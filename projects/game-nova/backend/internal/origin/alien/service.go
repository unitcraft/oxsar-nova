package alien

import (
	"oxsar/game-nova/internal/config"
	"oxsar/game-nova/internal/i18n"
)

// Service — обёртка для регистрации event.Handler'ов в worker'е.
//
// Содержит зависимости (catalog, bundle, loader, config). Stateless
// между событиями; каждое событие обрабатывается в своей транзакции.
//
// R0-исключение: Service одинаков для всех вселенных (uni01/uni02 +
// origin), как и весь пакет origin/alien (см. doc.go).
type Service struct {
	cfg    Config
	cat    *config.Catalog
	bundle *i18n.Bundle
	loader Loader
}

// NewService возвращает Service с дефолтным Config (1-в-1 с
// origin consts.php). Для override Config — `WithConfig`.
//
// loader может быть nil — тогда handlers, требующие load (только
// FlyUnknown.replan, Spawner) вернут ошибку. Для работы только
// post-event handlers (FlyUnknown без replan, ChangeMissionAI)
// loader не нужен.
func NewService(cat *config.Catalog, loader Loader) *Service {
	return &Service{
		cfg:    DefaultConfig(),
		cat:    cat,
		loader: loader,
	}
}

// WithBundle подключает i18n bundle для сообщений игроку.
func (s *Service) WithBundle(b *i18n.Bundle) *Service {
	s.bundle = b
	return s
}

// WithConfig перекрывает дефолтный Config (per-universe override —
// после Ф.3 будет грузиться из configs/balance/<universe>.yaml).
func (s *Service) WithConfig(cfg Config) *Service {
	s.cfg = cfg
	return s
}

// Config возвращает текущий Config (для тестов).
func (s *Service) Config() Config { return s.cfg }

// Catalog возвращает справочник (для тестов и helper'ов).
func (s *Service) Catalog() *config.Catalog { return s.cat }

// tr — короткое название trans-функции, как в internal/alien.Service.
func (s *Service) tr(group, key string, vars map[string]string) string {
	if s.bundle == nil {
		return "[" + group + "." + key + "]"
	}
	return s.bundle.Tr(i18n.LangRu, group, key, vars)
}
