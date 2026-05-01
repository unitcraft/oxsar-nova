package galaxy

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"

	"oxsar/game-nova/internal/auth"
	"oxsar/game-nova/internal/httpx"
)

// Handler — HTTP-адаптер чтения галактики.
//
// План 72.1.24: legacy `Galaxy.class.php::subtractHydrogen` списывает
// 10H при просмотре системы, отличной от текущей планеты. Эндпоинт
// принимает опциональный `?from_planet_id=` — если он задан, бэкенд
// списывает 10H с этой планеты (legacy эквивалент). Без параметра —
// backwards-compat: просто отдаёт view без списания.
type Handler struct {
	repo        *Repository
	numGalaxies int
	numSystems  int
}

// NewHandler — план 72.1 часть 12: лимиты вселенной приходят из
// configs/universes.yaml через cfg.Game.{NumGalaxies,NumSystems}.
func NewHandler(repo *Repository, numGalaxies, numSystems int) *Handler {
	return &Handler{repo: repo, numGalaxies: numGalaxies, numSystems: numSystems}
}

// galaxyViewHydrogenCost — стоимость одного просмотра чужой системы
// (legacy `STAR_SURVEILLANCE_CONSUMPTION` для phalanx — 5000H, но для
// просто галактики `Galaxy::subtractHydrogen` использует 10H).
const galaxyViewHydrogenCost int64 = 10

// ErrInsufficientHydrogen — недостаточно водорода на источнике.
var ErrInsufficientHydrogen = errors.New("galaxy: not enough hydrogen (10H required to view another system)")

// System GET /api/galaxy/{g}/{s}?from_planet_id=<uuid>
func (h *Handler) System(w http.ResponseWriter, r *http.Request) {
	g, err1 := strconv.Atoi(chi.URLParam(r, "g"))
	s, err2 := strconv.Atoi(chi.URLParam(r, "s"))
	if err1 != nil || err2 != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "invalid coords"))
		return
	}
	if err := (Coords{Galaxy: g, System: s, Position: 1}).Validate(h.numGalaxies, h.numSystems); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
		return
	}
	uid, _ := auth.UserID(r.Context())

	// План 72.1.24: hydrogen-cost при просмотре чужой системы.
	if fromPlanetID := r.URL.Query().Get("from_planet_id"); fromPlanetID != "" && uid != "" {
		if err := h.chargeHydrogenIfRemote(r, uid, fromPlanetID, g, s); err != nil {
			switch {
			case errors.Is(err, ErrInsufficientHydrogen):
				httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
			default:
				httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
			}
			return
		}
	}

	view, err := h.repo.ReadSystem(r.Context(), g, s, uid)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, view)
}

// chargeHydrogenIfRemote проверяет планету источника и списывает 10H
// если target-система не совпадает с текущей системой источника.
// Транзакция, чтобы не было race в SELECT .. UPDATE.
func (h *Handler) chargeHydrogenIfRemote(r *http.Request, userID, fromPlanetID string, g, s int) error {
	ctx := r.Context()
	tx, err := h.repo.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var (
		ownerID    string
		srcG, srcS int
		hydrogen   float64
	)
	err = tx.QueryRow(ctx, `
		SELECT user_id, galaxy, system, hydrogen
		FROM planets WHERE id = $1 AND destroyed_at IS NULL
		FOR UPDATE
	`, fromPlanetID).Scan(&ownerID, &srcG, &srcS, &hydrogen)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Не наша/несуществующая планета — игнорируем cost (UI просто
			// показывает view без списания, легаси редирект на main).
			return nil
		}
		return err
	}
	if ownerID != userID {
		return nil // чужая планета как источник — не списываем
	}
	if srcG == g && srcS == s {
		return nil // та же система — без cost (legacy)
	}
	if int64(hydrogen) < galaxyViewHydrogenCost {
		return ErrInsufficientHydrogen
	}
	if _, err := tx.Exec(ctx,
		`UPDATE planets SET hydrogen = hydrogen - $1 WHERE id = $2`,
		galaxyViewHydrogenCost, fromPlanetID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
