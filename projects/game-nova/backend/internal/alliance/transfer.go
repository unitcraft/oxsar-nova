// План 67 Ф.3: Transfer-leadership с email-подтверждением.
//
// Двухшаговый поток:
//   1. POST /api/alliances/{id}/transfer-leadership/code
//      body: {"new_owner_id": uuid}
//      → генерирует 8-символьный код, хэширует, пишет в
//        alliance_leadership_codes (миграция 0078), отправляет код
//        текущему owner'у системным сообщением (folder=13 SYSTEM).
//   2. POST /api/alliances/{id}/transfer-leadership
//      body: {"new_owner_id": uuid, "code": "XXXXXXXX"}
//      → верифицирует код, в одной транзакции:
//        - alliances.owner_id = new_owner_id
//        - alliances.leadership_transferred_at = now()
//        - alliance_members.rank = 'owner' для new_owner
//        - alliance_members.rank = 'member' для прежнего owner
//        - audit-запись (action=leadership_transferred)
//        - удаление кода
//
// Защиты:
//   - Только текущий owner может запросить и подтвердить.
//   - new_owner должен быть членом этого же альянса.
//   - Запрос/подтверждение для разных new_owner — код инвалидируется.
//   - TTL 10 минут, max 5 неверных попыток (как в settings/delete.go).
//   - Rate-limit: 3 запроса кода в час на альянс.
//
// Образец — internal/settings/delete.go (D-003 в плане 44).

package alliance

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"oxsar/game-nova/internal/auth"
)

const (
	transferCodeLen   = 8
	transferTTL       = 10 * time.Minute
	transferMaxAtt    = 5
	transferRateLimit = 3 // запросов кода в час
)

// Алфавит без похожих символов (0/O, 1/I/l) — повторяет settings.codeAlphabet.
var transferCodeAlphabet = []byte("ABCDEFGHJKMNPQRSTUVWXYZ23456789")

var (
	ErrTransferCodeExpired      = errors.New("alliance: transfer code expired")
	ErrTransferCodeInvalid      = errors.New("alliance: transfer code invalid")
	ErrTransferTooManyAttempts  = errors.New("alliance: too many transfer attempts")
	ErrTransferNoCode           = errors.New("alliance: transfer code not requested")
	ErrTransferRateLimit        = errors.New("alliance: too many transfer code requests")
	ErrTransferTargetMismatch   = errors.New("alliance: code was issued for different new owner")
	ErrTransferTargetNotMember  = errors.New("alliance: target user is not a member of this alliance")
	ErrTransferTargetIsSelf     = errors.New("alliance: cannot transfer to yourself")
	ErrTransferOwnerChanged     = errors.New("alliance: ownership changed since code was issued")
)

// TransferCodeIssued — ответ на RequestTransferCode.
type TransferCodeIssued struct {
	ExpiresAt  time.Time `json:"expires_at"`
	TTLSeconds int       `json:"ttl_seconds"`
}

// RequestTransferCode генерирует код подтверждения и отправляет его
// текущему owner'у системным сообщением.
//
// Только owner альянса может запрашивать. new_owner_id должен быть
// членом этого альянса и не равен requester.
func (s *Service) RequestTransferCode(ctx context.Context, requesterID, allianceID, newOwnerID string) (TransferCodeIssued, error) {
	if newOwnerID == requesterID {
		return TransferCodeIssued{}, ErrTransferTargetIsSelf
	}

	// Pre-check без транзакции: rate-limit.
	var recent int
	if err := s.db.Pool().QueryRow(ctx, `
		SELECT COUNT(*) FROM alliance_leadership_codes
		WHERE alliance_id = $1 AND issued_at > now() - interval '1 hour'
	`, allianceID).Scan(&recent); err != nil {
		return TransferCodeIssued{}, fmt.Errorf("transfer code: rate-check: %w", err)
	}
	if recent >= transferRateLimit {
		return TransferCodeIssued{}, ErrTransferRateLimit
	}

	code, err := generateTransferCode()
	if err != nil {
		return TransferCodeIssued{}, fmt.Errorf("transfer code: gen: %w", err)
	}
	hash, err := auth.HashPassword(code)
	if err != nil {
		return TransferCodeIssued{}, fmt.Errorf("transfer code: hash: %w", err)
	}
	expiresAt := time.Now().Add(transferTTL)

	err = s.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		// Проверка: requester — owner альянса.
		var ownerID string
		if err := tx.QueryRow(ctx,
			`SELECT owner_id FROM alliances WHERE id=$1`, allianceID).Scan(&ownerID); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrNotFound
			}
			return fmt.Errorf("read alliance: %w", err)
		}
		if ownerID != requesterID {
			return ErrNotOwner
		}

		// Проверка: new_owner — член этого же альянса.
		var memAlliance string
		err := tx.QueryRow(ctx,
			`SELECT alliance_id FROM alliance_members WHERE user_id=$1`, newOwnerID).Scan(&memAlliance)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrTransferTargetNotMember
			}
			return fmt.Errorf("read target membership: %w", err)
		}
		if memAlliance != allianceID {
			return ErrTransferTargetNotMember
		}

		// UPSERT кода.
		_, err = tx.Exec(ctx, `
			INSERT INTO alliance_leadership_codes
				(alliance_id, requester_id, new_owner_id, code_hash, issued_at, expires_at, attempts)
			VALUES ($1, $2, $3, $4, now(), $5, 0)
			ON CONFLICT (alliance_id) DO UPDATE SET
				requester_id  = EXCLUDED.requester_id,
				new_owner_id  = EXCLUDED.new_owner_id,
				code_hash     = EXCLUDED.code_hash,
				issued_at     = now(),
				expires_at    = EXCLUDED.expires_at,
				attempts      = 0
		`, allianceID, requesterID, newOwnerID, hash, expiresAt)
		if err != nil {
			return fmt.Errorf("upsert code: %w", err)
		}
		return nil
	})
	if err != nil {
		return TransferCodeIssued{}, err
	}

	// Отправить код системным сообщением (folder=13 SYSTEM).
	// Best-effort: ошибка не пробрасывается — игрок увидит код в
	// inbox'е, но запрос уже зафиксирован и можно подтвердить.
	if s.automsg != nil {
		title := s.tr("alliance", "transferLeadership.codeTitle", nil)
		body := s.tr("alliance", "transferLeadership.codeBody", map[string]string{
			"code":      code,
			"expiresAt": expiresAt.Format("15:04 02.01.2006"),
		})
		_ = s.automsg.SendDirect(ctx, nil, requesterID, 13, title, body)
	}

	return TransferCodeIssued{
		ExpiresAt:  expiresAt.UTC(),
		TTLSeconds: int(transferTTL.Seconds()),
	}, nil
}

// ConfirmTransferLeadership верифицирует код и переносит лидерство в
// одной транзакции.
func (s *Service) ConfirmTransferLeadership(ctx context.Context, requesterID, allianceID, newOwnerID, code string) error {
	if newOwnerID == requesterID {
		return ErrTransferTargetIsSelf
	}
	if code == "" {
		return ErrTransferCodeInvalid
	}

	return s.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		// 1. Прочесть код под FOR UPDATE.
		var (
			storedRequester string
			storedNewOwner  string
			codeHash        string
			expiresAt       time.Time
			attempts        int
		)
		err := tx.QueryRow(ctx, `
			SELECT requester_id, new_owner_id, code_hash, expires_at, attempts
			FROM alliance_leadership_codes
			WHERE alliance_id = $1
			FOR UPDATE
		`, allianceID).Scan(&storedRequester, &storedNewOwner, &codeHash, &expiresAt, &attempts)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrTransferNoCode
			}
			return fmt.Errorf("read transfer code: %w", err)
		}

		// 2. TTL.
		if expiresAt.Before(time.Now()) {
			_, _ = tx.Exec(ctx,
				`DELETE FROM alliance_leadership_codes WHERE alliance_id=$1`, allianceID)
			return ErrTransferCodeExpired
		}

		// 3. Лимит попыток.
		if attempts >= transferMaxAtt {
			_, _ = tx.Exec(ctx,
				`DELETE FROM alliance_leadership_codes WHERE alliance_id=$1`, allianceID)
			return ErrTransferTooManyAttempts
		}

		// 4. Код выпускался для другого requester / new_owner.
		if storedRequester != requesterID {
			return ErrTransferOwnerChanged
		}
		if storedNewOwner != newOwnerID {
			return ErrTransferTargetMismatch
		}

		// 5. Проверка hash.
		ok, err := auth.VerifyPassword(code, codeHash)
		if err != nil {
			return fmt.Errorf("verify transfer code: %w", err)
		}
		if !ok {
			if _, err := tx.Exec(ctx, `
				UPDATE alliance_leadership_codes SET attempts = attempts + 1
				WHERE alliance_id=$1
			`, allianceID); err != nil {
				return fmt.Errorf("bump transfer attempts: %w", err)
			}
			return ErrTransferCodeInvalid
		}

		// 6. Финальная проверка ownership под FOR UPDATE — могло
		// измениться между issue и confirm (другой transfer
		// выполнился, disband, etc.).
		var currentOwner string
		if err := tx.QueryRow(ctx,
			`SELECT owner_id FROM alliances WHERE id=$1 FOR UPDATE`, allianceID).Scan(&currentOwner); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrNotFound
			}
			return fmt.Errorf("recheck owner: %w", err)
		}
		if currentOwner != requesterID {
			_, _ = tx.Exec(ctx,
				`DELETE FROM alliance_leadership_codes WHERE alliance_id=$1`, allianceID)
			return ErrTransferOwnerChanged
		}

		// 7. Проверка new_owner всё ещё член альянса.
		var targetAlliance string
		err = tx.QueryRow(ctx,
			`SELECT alliance_id FROM alliance_members WHERE user_id=$1 FOR UPDATE`, newOwnerID).Scan(&targetAlliance)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrTransferTargetNotMember
			}
			return fmt.Errorf("recheck target membership: %w", err)
		}
		if targetAlliance != allianceID {
			return ErrTransferTargetNotMember
		}

		// 8. Сама передача.
		if _, err := tx.Exec(ctx, `
			UPDATE alliances
			SET owner_id = $1, leadership_transferred_at = now()
			WHERE id = $2
		`, newOwnerID, allianceID); err != nil {
			return fmt.Errorf("update alliance owner: %w", err)
		}
		// Прежний owner становится member'ом (rank), новый — owner'ом.
		// rank_id оставляем как есть; UI/owner может позже выставить
		// конкретный custom-rank.
		if _, err := tx.Exec(ctx, `
			UPDATE alliance_members SET rank='member'
			WHERE alliance_id=$1 AND user_id=$2
		`, allianceID, requesterID); err != nil {
			return fmt.Errorf("demote old owner: %w", err)
		}
		if _, err := tx.Exec(ctx, `
			UPDATE alliance_members SET rank='owner'
			WHERE alliance_id=$1 AND user_id=$2
		`, allianceID, newOwnerID); err != nil {
			return fmt.Errorf("promote new owner: %w", err)
		}

		// 9. Audit.
		writeAuditTx(ctx, tx, allianceID, requesterID,
			ActionLeadershipTransferred, TargetKindUser, newOwnerID,
			map[string]any{"new_owner_id": newOwnerID, "previous_owner_id": requesterID})

		// 10. Удалить код.
		if _, err := tx.Exec(ctx,
			`DELETE FROM alliance_leadership_codes WHERE alliance_id=$1`, allianceID); err != nil {
			return fmt.Errorf("delete transfer code: %w", err)
		}

		return nil
	})
}

func generateTransferCode() (string, error) {
	buf := make([]byte, transferCodeLen)
	raw := make([]byte, transferCodeLen)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	for i, b := range raw {
		buf[i] = transferCodeAlphabet[int(b)%len(transferCodeAlphabet)]
	}
	return string(buf), nil
}
