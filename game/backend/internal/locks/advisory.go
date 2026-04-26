// Package locks — distributed locking через Postgres advisory lock.
//
// Используется в scheduler'е (план 32) для гарантии single-execution
// при N≥2 worker'ах. Также подходит для любых ad-hoc singleton-задач:
// миграции данных при первом запуске, single-run cron-job'ов и т.п.
//
// Postgres advisory lock работает per-session: lock держится, пока
// connection открыт. Если процесс упал — Postgres освободит lock при
// disconnect. Это делает их более надёжными, чем Redis-locks без TTL.
//
// Важно: lock привязан к конкретному соединению. Поэтому мы выделяем
// один pgx.Conn из пула, держим его до конца критической секции,
// потом возвращаем в пул. Использовать pool напрямую нельзя: следующий
// запрос может попасть на другой connection, и lock не отпустится.
package locks

import (
	"context"
	"fmt"
	"hash/fnv"

	"github.com/jackc/pgx/v5/pgxpool"
)

// TryRun пытается взять advisory lock с именем lockName и при успехе
// выполняет fn. Возвращает acquired=true если lock был взят (fn
// выполнена), acquired=false если lock уже у кого-то (fn пропущена,
// err=nil). Любая ошибка fn возвращается как err при acquired=true.
//
// При acquired=false fn НЕ вызывается — это позволяет caller'у
// логировать skip / инкрементить metrics.
//
//	acquired, err := locks.TryRun(ctx, pool, "alien_spawn", func(ctx context.Context) error {
//	    return alienSvc.Spawn(ctx)
//	})
//	if !acquired {
//	    log.Debug("skipped: another worker holds lock")
//	}
func TryRun(ctx context.Context, pool *pgxpool.Pool, lockName string, fn func(ctx context.Context) error) (acquired bool, err error) {
	if lockName == "" {
		return false, fmt.Errorf("locks: empty name")
	}
	if pool == nil {
		return false, fmt.Errorf("locks: nil pool")
	}

	conn, err := pool.Acquire(ctx)
	if err != nil {
		return false, fmt.Errorf("locks: acquire conn: %w", err)
	}
	defer conn.Release()

	key := hashLockName(lockName)

	var ok bool
	if err := conn.QueryRow(ctx, "SELECT pg_try_advisory_lock($1)", key).Scan(&ok); err != nil {
		return false, fmt.Errorf("locks: try_advisory_lock %q: %w", lockName, err)
	}
	if !ok {
		return false, nil
	}

	// При panic в fn defer вернёт connection в пул, а Postgres сам
	// сбросит lock при close session (если pool закроет connection
	// из-за ошибки). Чтобы гарантировать unlock в нормальном случае,
	// делаем явный unlock в defer.
	defer func() {
		if _, uerr := conn.Exec(context.Background(), "SELECT pg_advisory_unlock($1)", key); uerr != nil {
			// Не маскируем ошибку fn ошибкой unlock'а — только если fn ОК.
			if err == nil {
				err = fmt.Errorf("locks: advisory_unlock %q: %w", lockName, uerr)
			}
		}
	}()

	if ferr := fn(ctx); ferr != nil {
		return true, ferr
	}
	return true, nil
}

// hashLockName переводит имя lock'а в стабильный int64 для
// pg_try_advisory_lock. FNV-64 даёт хорошее распределение, разные
// имена практически не пересекаются на нашем масштабе (десятки
// jobs).
func hashLockName(name string) int64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(name))
	return int64(h.Sum64())
}
