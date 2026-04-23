// Package battlestats — история боёв текущего игрока с фильтрами.
package battlestats

import (
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/oxsar/nova/backend/internal/auth"
	"github.com/oxsar/nova/backend/internal/httpx"
)

type Handler struct {
	pool *pgxpool.Pool
}

func NewHandler(pool *pgxpool.Pool) *Handler { return &Handler{pool: pool} }

type battleRow struct {
	ID             string  `json:"id"`
	At             string  `json:"at"`
	Winner         string  `json:"winner"`
	Rounds         int     `json:"rounds"`
	Role           string  `json:"role"` // attacker | defender
	Opponent       string  `json:"opponent"`
	OpponentID     *string `json:"opponent_id,omitempty"`
	PlanetName     *string `json:"planet_name,omitempty"`
	DebrisMetal    float64 `json:"debris_metal"`
	DebrisSilicon  float64 `json:"debris_silicon"`
	LootMetal      float64 `json:"loot_metal"`
	LootSilicon    float64 `json:"loot_silicon"`
	LootHydrogen   float64 `json:"loot_hydrogen"`
}

type response struct {
	Battles []battleRow `json:"battles"`
	Total   int         `json:"total"`
	Wins    int         `json:"wins"`
	Losses  int         `json:"losses"`
	Draws   int         `json:"draws"`
}

// List GET /api/battlestats?role=attacker|defender|any&result=win|loss|draw|any&from=YYYY-MM-DD&to=YYYY-MM-DD&limit=N
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}

	role := r.URL.Query().Get("role")
	result := r.URL.Query().Get("result")
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")

	var where []string
	args := []any{uid}
	where = append(where, "(attacker_user_id = $1 OR defender_user_id = $1)")

	if role == "attacker" {
		where = append(where, "attacker_user_id = $1")
	} else if role == "defender" {
		where = append(where, "defender_user_id = $1")
	}

	if from != "" {
		if t, err := time.Parse("2006-01-02", from); err == nil {
			args = append(args, t)
			where = append(where, "at >= $"+itoa(len(args)))
		}
	}
	if to != "" {
		if t, err := time.Parse("2006-01-02", to); err == nil {
			args = append(args, t.Add(24*time.Hour))
			where = append(where, "at < $"+itoa(len(args)))
		}
	}

	query := `
		SELECT br.id, br.at, br.winner, br.rounds,
		       br.attacker_user_id, br.defender_user_id,
		       COALESCE(a.username::text, ''), COALESCE(d.username::text, ''),
		       p.name,
		       br.debris_metal, br.debris_silicon,
		       br.loot_metal, br.loot_silicon, br.loot_hydrogen
		FROM battle_reports br
		LEFT JOIN users a ON a.id = br.attacker_user_id
		LEFT JOIN users d ON d.id = br.defender_user_id
		LEFT JOIN planets p ON p.id = br.planet_id
		WHERE ` + strings.Join(where, " AND ") + `
		ORDER BY br.at DESC
		LIMIT 200
	`

	rows, err := h.pool.Query(r.Context(), query, args...)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	defer rows.Close()

	resp := response{Battles: []battleRow{}}
	for rows.Next() {
		var br battleRow
		var at time.Time
		var attackerID, defenderID *string
		var attackerName, defenderName string
		var planetName *string
		if err := rows.Scan(
			&br.ID, &at, &br.Winner, &br.Rounds,
			&attackerID, &defenderID,
			&attackerName, &defenderName, &planetName,
			&br.DebrisMetal, &br.DebrisSilicon,
			&br.LootMetal, &br.LootSilicon, &br.LootHydrogen,
		); err != nil {
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
			return
		}
		br.At = at.UTC().Format(time.RFC3339)
		br.PlanetName = planetName

		// Определить роль.
		if attackerID != nil && *attackerID == uid {
			br.Role = "attacker"
			br.Opponent = defenderName
			br.OpponentID = defenderID
		} else {
			br.Role = "defender"
			br.Opponent = attackerName
			br.OpponentID = attackerID
		}

		// Фильтр по результату.
		isWin := (br.Role == "attacker" && br.Winner == "attackers") ||
			(br.Role == "defender" && br.Winner == "defenders")
		isLoss := (br.Role == "attacker" && br.Winner == "defenders") ||
			(br.Role == "defender" && br.Winner == "attackers")
		isDraw := br.Winner == "draw"

		switch result {
		case "win":
			if !isWin {
				continue
			}
		case "loss":
			if !isLoss {
				continue
			}
		case "draw":
			if !isDraw {
				continue
			}
		}

		if isWin {
			resp.Wins++
		} else if isLoss {
			resp.Losses++
		} else if isDraw {
			resp.Draws++
		}
		resp.Total++
		resp.Battles = append(resp.Battles, br)
	}

	httpx.WriteJSON(w, r, http.StatusOK, resp)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var s [8]byte
	i := len(s)
	for n > 0 {
		i--
		s[i] = byte('0' + n%10)
		n /= 10
	}
	return string(s[i:])
}
