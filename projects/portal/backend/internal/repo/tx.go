// Package repo оборачивает pgxpool в слой-транзактор.
//
// Все сервисы получают repo.Exec (интерфейс, закрывающий Query/Exec
// и InTx). Интерфейс нужен, чтобы подменять pgx в unit-тестах.
package repo

// DUPLICATE: этот файл скопирован между Go-модулями oxsar/game-nova,
// oxsar/auth, oxsar/portal и oxsar/billing. При любом изменении
// синхронизируйте КОПИИ:
//   - projects/game-nova/backend/internal/repo/tx.go
//   - projects/auth/backend/internal/repo/tx.go
//   - projects/portal/backend/internal/repo/tx.go
//   - projects/billing/backend/internal/repo/tx.go
// Причина дубля: каждый домен — отдельный go.mod, без shared-модуля.

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TxFunc — тело транзакции. Если возвращает ошибку — rollback.
type TxFunc func(ctx context.Context, tx pgx.Tx) error

// Exec — минимальный интерфейс, который сервисы требуют от БД.
type Exec interface {
	InTx(ctx context.Context, fn TxFunc) error
	Pool() *pgxpool.Pool
}

// PG реализует Exec поверх pgxpool.
type PG struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *PG { return &PG{pool: pool} }

func (p *PG) Pool() *pgxpool.Pool { return p.pool }

// InTx выполняет fn внутри транзакции. При ошибке fn или commit'а делает
// rollback. Паники в fn — rollback + проброс panic (сохранение стека).
func (p *PG) InTx(ctx context.Context, fn TxFunc) error {
	tx, err := p.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if rec := recover(); rec != nil {
			_ = tx.Rollback(ctx)
			panic(rec)
		}
	}()

	if err := fn(ctx, tx); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("tx fn: %w (rollback: %v)", err, rbErr)
		}
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}
