package balance

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"gopkg.in/yaml.v3"

	"oxsar/game-nova/internal/config"
	"oxsar/game-nova/pkg/metrics"
)

// ErrInvalidOverride — override-файл найден, но не парсится / содержит
// невалидные значения. Sentinel-ошибка, чтобы вызывающие могли отличать
// «no override» (норма) от «override broken» (фатально на старте).
var ErrInvalidOverride = errors.New("balance: invalid override file")

// Loader загружает балансовый bundle для вселенной по её ID.
//
// Loader потокобезопасен: in-memory-кеш per-universe, защищённый
// мьютексом. Один процесс обычно работает только с одной вселенной
// (cfg.Auth.UniverseID), но Loader поддерживает многоарендность для
// тестов и cmd/tools/* (battle-sim, resync, wiki-gen), которые могут
// итерировать несколько вселенных.
//
// Замена override после старта не поддерживается (immutable bundle —
// см. план 64 §1, override определяется наличием файла на старте,
// hot-reload откладывается).
type Loader struct {
	configsDir string

	mu      sync.RWMutex
	defOnce sync.Once
	defB    *Bundle
	defErr  error
	cache   map[string]*Bundle
}

// NewLoader создаёт loader, читающий YAML-справочники из configsDir.
//
// configsDir указывает на корень configs/ (где лежат buildings.yml,
// units.yml и подкаталог balance/). Существующая ENV-переменная
// CATALOG_DIR — тот же самый путь, который передаётся в config.LoadCatalog.
func NewLoader(configsDir string) *Loader {
	return &Loader{
		configsDir: configsDir,
		cache:      make(map[string]*Bundle),
	}
}

// LoadDefaults — загружает дефолтный bundle (modern-баланс без override).
//
// Возвращает Bundle{UniverseID: "", HasOverride: false}. Все modern-
// вселенные (uni01, uni02 и любые будущие, у которых нет override-
// файла) используют этот bundle.
//
// Возвращаемая ссылка — shared (один объект на процесс): bundle
// immutable, безопасно использовать из нескольких goroutine. Кеширует
// результат — повторные вызовы не перечитывают YAML.
func (l *Loader) LoadDefaults() (*Bundle, error) {
	l.defOnce.Do(func() {
		cat, err := config.LoadCatalog(l.configsDir)
		if err != nil {
			l.defErr = fmt.Errorf("balance: load catalog: %w", err)
			return
		}
		l.defB = &Bundle{
			UniverseID:  "",
			HasOverride: false,
			Catalog:     cat,
			Globals:     ModernGlobals(),
		}
	})
	return l.defB, l.defErr
}

// LoadForCtx — то же что LoadFor, но с context для slog/metrics. Возвращает
// тот же Bundle, что LoadFor (cache shared). Если ctx уже отменён, не
// прерывает чтение — YAML-файл локальный, всё равно быстрее ms; ctx
// нужен только для логирования.
func (l *Loader) LoadForCtx(ctx context.Context, log *slog.Logger, universeID string) (*Bundle, error) {
	start := time.Now()
	b, err := l.LoadFor(universeID)
	dur := time.Since(start)

	metrics.RegisterBalance()
	status := "ok"
	override := false
	if err != nil {
		status = "error"
	} else if b != nil {
		override = b.HasOverride
	}
	if metrics.BalanceLoadTotal != nil {
		metrics.BalanceLoadTotal.WithLabelValues(universeID, status, strconv.FormatBool(override)).Inc()
	}
	if metrics.BalanceLoadDuration != nil && status == "ok" {
		metrics.BalanceLoadDuration.WithLabelValues(universeID).Observe(dur.Seconds())
	}

	if log != nil {
		if err != nil {
			log.ErrorContext(ctx, "balance load failed",
				slog.String("universe_id", universeID),
				slog.String("err", err.Error()),
				slog.Duration("duration", dur))
		} else {
			log.InfoContext(ctx, "balance bundle loaded",
				slog.String("universe_id", universeID),
				slog.Bool("override_applied", override),
				slog.Int("buildings", len(b.Catalog.Buildings.Buildings)),
				slog.Int("ships", len(b.Catalog.Ships.Ships)),
				slog.Duration("duration", dur))
		}
	}
	return b, err
}

// LoadFor — загружает bundle для вселенной с заданным ID.
//
// Алгоритм:
//
//  1. Если для universeID существует configs/balance/<id>.yaml — он
//     применяется поверх дефолта (deep merge). Возвращается bundle
//     с HasOverride=true.
//  2. Иначе — возвращается дефолтный bundle (LoadDefaults), HasOverride=false.
//
// Ошибки:
//   - read default catalog → пробрасывается из LoadDefaults
//   - parse override yaml → ErrInvalidOverride wrapping fmt.Errorf
//   - merge override → ErrInvalidOverride wrapping причину
//
// Кеширует bundle per-universe. Два вызова LoadFor("origin") вернут
// тот же *Bundle (по указателю).
func (l *Loader) LoadFor(universeID string) (*Bundle, error) {
	if universeID == "" {
		return l.LoadDefaults()
	}
	l.mu.RLock()
	if b, ok := l.cache[universeID]; ok {
		l.mu.RUnlock()
		return b, nil
	}
	l.mu.RUnlock()

	l.mu.Lock()
	defer l.mu.Unlock()
	if b, ok := l.cache[universeID]; ok {
		return b, nil
	}

	def, err := l.LoadDefaults()
	if err != nil {
		return nil, err
	}

	overridePath := filepath.Join(l.configsDir, "balance", universeID+".yaml")
	data, err := os.ReadFile(overridePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// Нет override → возвращаем дефолт с проставленным UniverseID.
			b := &Bundle{
				UniverseID:  universeID,
				HasOverride: false,
				Catalog:     def.Catalog,
				Globals:     def.Globals,
			}
			l.cache[universeID] = b
			return b, nil
		}
		return nil, fmt.Errorf("balance: read override %q: %w", overridePath, err)
	}

	var ov override
	if err := yaml.Unmarshal(data, &ov); err != nil {
		return nil, fmt.Errorf("%w: parse %q: %v", ErrInvalidOverride, overridePath, err)
	}
	if ov.Universe != "" && ov.Universe != universeID {
		return nil, fmt.Errorf("%w: %q has universe=%q, expected %q",
			ErrInvalidOverride, overridePath, ov.Universe, universeID)
	}

	merged, err := applyOverride(def, &ov)
	if err != nil {
		return nil, fmt.Errorf("%w: %s: %v", ErrInvalidOverride, overridePath, err)
	}
	merged.UniverseID = universeID
	merged.HasOverride = true
	l.cache[universeID] = merged
	return merged, nil
}
