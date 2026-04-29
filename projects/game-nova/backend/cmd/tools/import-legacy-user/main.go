// Command import-legacy-user — оффлайн-инструмент для копирования одного
// тестового юзера из legacy MySQL (oxsar2-PHP, na_*) в nova Postgres
// (game-nova-db.users + planets + ... и identity-db.users).
//
// Зачем: для side-by-side сравнения экранов origin-фронта (план 72) с
// legacy-php (план 43) — юзер test (userid=1) должен иметь семантически
// эквивалентные данные в обоих БД. Legacy-сторона заполняется
// projects/game-legacy-php/tools/apply-test-user-fixture.sh, nova-сторона —
// этим CLI (план 87).
//
// Идемпотентен: при повторном запуске удаляет ранее импортированного
// юзера во всех target-таблицах и заливает заново.
//
// Запуск (defaults для local docker stack):
//
//	cd projects/game-nova/backend
//	go run ./cmd/tools/import-legacy-user
//
// Полный список флагов: --help. Также есть Makefile-target
// `make import-legacy-user`.
//
// План 87 Ф.1-Ф.6.
package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"oxsar/game-nova/internal/auth"
	"oxsar/game-nova/pkg/ids"
)

// Defaults для local Docker stack (см. deploy/docker-compose.yml + project
// game-legacy-php docker-compose). Пользователь может переопределить через
// флаги или через ENV-переменные в Makefile.
const (
	defaultLegacyDSN   = "root:root_pass@tcp(127.0.0.1:3307)/oxsar_db"
	defaultIdentityDSN = "postgres://identitysvc:identitysvc@127.0.0.1:5435/identitysvc?sslmode=disable"
	defaultGameDSN     = "postgres://oxsar:oxsar@127.0.0.1:5433/oxsar?sslmode=disable"
)

type config struct {
	legacyDSN      string
	identityDSN    string
	gameDSN        string
	legacyUserID   int
	targetUsername string
	targetPassword string
	targetUniverse string
	messageLimit   int
}

func main() {
	cfg := parseFlags()

	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(log)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if err := run(ctx, cfg); err != nil {
		slog.Error("import-legacy-user failed", slog.String("err", err.Error()))
		os.Exit(1)
	}
	slog.Info("import-legacy-user: done")
}

func parseFlags() config {
	var cfg config
	flag.StringVar(&cfg.legacyDSN, "legacy-db", defaultLegacyDSN, "MySQL DSN to legacy oxsar_db (game-legacy-php docker)")
	flag.StringVar(&cfg.identityDSN, "identity-db", defaultIdentityDSN, "Postgres URL to identity-db")
	flag.StringVar(&cfg.gameDSN, "game-db", defaultGameDSN, "Postgres URL to game-nova-db")
	flag.IntVar(&cfg.legacyUserID, "legacy-userid", 1, "userid in legacy na_user to import")
	flag.StringVar(&cfg.targetUsername, "target-username", "test", "username to use in nova (identity + game)")
	flag.StringVar(&cfg.targetPassword, "target-password", "test1234", "password for the imported user (argon2id-hashed in identity-db)")
	flag.StringVar(&cfg.targetUniverse, "target-universe", "uni01", "universe_id (used for identity universe_memberships)")
	flag.IntVar(&cfg.messageLimit, "message-limit", 50, "limit messages copied per direction (inbox/outbox)")
	flag.Parse()
	return cfg
}

func run(ctx context.Context, cfg config) error {
	legacy, err := sql.Open("mysql", cfg.legacyDSN)
	if err != nil {
		return fmt.Errorf("open legacy mysql: %w", err)
	}
	defer legacy.Close()
	if err := legacy.PingContext(ctx); err != nil {
		return fmt.Errorf("ping legacy mysql: %w", err)
	}

	identityPool, err := pgxpool.New(ctx, cfg.identityDSN)
	if err != nil {
		return fmt.Errorf("open identity-db: %w", err)
	}
	defer identityPool.Close()
	if err := identityPool.Ping(ctx); err != nil {
		return fmt.Errorf("ping identity-db: %w", err)
	}

	gamePool, err := pgxpool.New(ctx, cfg.gameDSN)
	if err != nil {
		return fmt.Errorf("open game-nova-db: %w", err)
	}
	defer gamePool.Close()
	if err := gamePool.Ping(ctx); err != nil {
		return fmt.Errorf("ping game-nova-db: %w", err)
	}

	slog.InfoContext(ctx, "connected to all 3 databases",
		slog.Int("legacy_userid", cfg.legacyUserID),
		slog.String("target_username", cfg.targetUsername))

	src, err := loadLegacyUser(ctx, legacy, cfg.legacyUserID)
	if err != nil {
		return fmt.Errorf("load legacy user: %w", err)
	}
	slog.InfoContext(ctx, "legacy user loaded",
		slog.String("username", src.user.username),
		slog.Int("planets", len(src.planets)),
		slog.Int("buildings", len(src.buildings)),
		slog.Int("researches", len(src.researches)),
		slog.Int("ships", len(src.ships)),
		slog.Int("artefacts", len(src.artefacts)))

	if err := cleanupTestUser(ctx, identityPool, gamePool, cfg.targetUsername); err != nil {
		return fmt.Errorf("cleanup: %w", err)
	}

	userID := ids.New() // UUIDv7

	if err := insertIdentityUser(ctx, identityPool, userID, cfg); err != nil {
		return fmt.Errorf("insert identity user: %w", err)
	}

	stats, err := insertGameUser(ctx, gamePool, userID, cfg, src)
	if err != nil {
		return fmt.Errorf("insert game user: %w", err)
	}

	slog.InfoContext(ctx, "import summary",
		slog.String("user_id", userID),
		slog.String("username", cfg.targetUsername),
		slog.Int("planets", stats.planets),
		slog.Int("buildings", stats.buildings),
		slog.Int("research_levels", stats.research),
		slog.Int("ships", stats.ships),
		slog.Int("defense", stats.defense),
		slog.Int("artefacts", stats.artefacts),
		slog.Int("messages", stats.messages),
		slog.Int("alliance_members", stats.allianceMembers),
		slog.Int("skipped_unknown_units", stats.skippedUnknownUnits))

	return nil
}

// =====================================================================
// Legacy data loading
// =====================================================================

type legacySource struct {
	user       legacyUser
	planets    []legacyPlanet
	galaxy     map[int]legacyGalaxyCell // planetid → cell
	buildings  []legacyBuilding
	researches []legacyResearch
	ships      []legacyShipRow // covers ships, defense, rockets — split by id
	artefacts  []legacyArtefact
	alliance   *legacyAlliance
	notepad    string
	messages   []legacyMessage
}

type legacyUser struct {
	id              int
	username        string
	email           string
	profession      int
	credit          float64
	exchangeRate    float64
	researchFactor  float64
	points          float64
	maxPoints       float64
	curPlanet       int
	languageID      int
	regtime         int64
	last            int64
	umode           int
	umodemin        int
	protectionTime  int64
	planetTeleport  int64
	observer        int
	tutorialState   int
	chatLanguageID  int
	lastChat        int64
	lastChatAlly    int64
}

type legacyPlanet struct {
	planetid           int
	ismoon             int
	planetname         string
	diameter           int
	picture            string
	temperature        int
	last               int64
	metal              float64
	silicon            float64
	hydrogen           float64
	solarSatelliteProd int
	buildFactor        float64
	researchFactor     float64
	produceFactor      float64
	energyFactor       float64
	storageFactor      float64
}

type legacyGalaxyCell struct {
	galaxy   int
	system   int
	position int
}

type legacyBuilding struct {
	planetid   int
	buildingid int
	level      int
}

type legacyResearch struct {
	buildingid int
	level      int
}

type legacyShipRow struct {
	planetid     int
	unitid       int
	quantity     int64
	damaged      int64
	shellPercent float64
}

type legacyArtefact struct {
	artid     int
	typeid    int
	planetid  int
	active    int
	timesLeft int
	deleted   int64
}

type legacyAlliance struct {
	aid          int
	tag          string
	name         string
	founder      int
	textextern   string
	textintern   string
	homepage     string
	founderName  string
	memberRank   int
	joinDate     int64
}

type legacyMessage struct {
	msgid    int64
	mode     int
	time     int64
	sender   sql.NullInt64
	receiver int64
	message  string
	subject  string
	readed   int
}

func loadLegacyUser(ctx context.Context, db *sql.DB, userID int) (*legacySource, error) {
	src := &legacySource{galaxy: map[int]legacyGalaxyCell{}}

	// na_user
	row := db.QueryRowContext(ctx, `
		SELECT userid, username, email, profession, credit, exchange_rate,
		       research_factor, points, max_points, curplanet, languageid,
		       regtime, last, umode, umodemin, protection_time,
		       planet_teleport_time, observer, tutorial_state,
		       chat_languageid, last_chat, last_chatally
		FROM na_user WHERE userid = ?`, userID)
	u := legacyUser{}
	if err := row.Scan(&u.id, &u.username, &u.email, &u.profession, &u.credit,
		&u.exchangeRate, &u.researchFactor, &u.points, &u.maxPoints,
		&u.curPlanet, &u.languageID, &u.regtime, &u.last, &u.umode,
		&u.umodemin, &u.protectionTime, &u.planetTeleport, &u.observer,
		&u.tutorialState, &u.chatLanguageID, &u.lastChat, &u.lastChatAlly); err != nil {
		return nil, fmt.Errorf("select na_user: %w", err)
	}
	src.user = u

	// na_planet
	planetRows, err := db.QueryContext(ctx, `
		SELECT planetid, ismoon, planetname, diameter, picture, temperature,
		       last, metal, silicon, hydrogen, solar_satellite_prod,
		       build_factor, research_factor, produce_factor,
		       energy_factor, storage_factor
		FROM na_planet WHERE userid = ?`, userID)
	if err != nil {
		return nil, fmt.Errorf("select na_planet: %w", err)
	}
	for planetRows.Next() {
		var p legacyPlanet
		var picture []byte
		if err := planetRows.Scan(&p.planetid, &p.ismoon, &p.planetname, &p.diameter,
			&picture, &p.temperature, &p.last, &p.metal, &p.silicon, &p.hydrogen,
			&p.solarSatelliteProd, &p.buildFactor, &p.researchFactor,
			&p.produceFactor, &p.energyFactor, &p.storageFactor); err != nil {
			planetRows.Close()
			return nil, fmt.Errorf("scan na_planet: %w", err)
		}
		p.picture = string(picture)
		src.planets = append(src.planets, p)
	}
	planetRows.Close()

	// na_galaxy — координаты планет/лун. Только те cells, где planetid или
	// moonid принадлежит юзеру; чтобы не тащить всю галактику.
	planetIDs := make([]int, 0, len(src.planets))
	for _, p := range src.planets {
		planetIDs = append(planetIDs, p.planetid)
	}
	if len(planetIDs) > 0 {
		placeholders := strings.Repeat("?,", len(planetIDs))
		placeholders = placeholders[:len(placeholders)-1]
		args := make([]any, 0, 2*len(planetIDs))
		for _, id := range planetIDs {
			args = append(args, id)
		}
		for _, id := range planetIDs {
			args = append(args, id)
		}
		q := fmt.Sprintf(`
			SELECT galaxy, system, position, planetid, moonid
			FROM na_galaxy
			WHERE planetid IN (%s) OR moonid IN (%s)`, placeholders, placeholders)
		gRows, err := db.QueryContext(ctx, q, args...)
		if err != nil {
			return nil, fmt.Errorf("select na_galaxy: %w", err)
		}
		for gRows.Next() {
			var g, s, pos int
			var pid sql.NullInt64
			var mid sql.NullInt64
			if err := gRows.Scan(&g, &s, &pos, &pid, &mid); err != nil {
				gRows.Close()
				return nil, fmt.Errorf("scan na_galaxy: %w", err)
			}
			if pid.Valid {
				src.galaxy[int(pid.Int64)] = legacyGalaxyCell{g, s, pos}
			}
			if mid.Valid {
				src.galaxy[int(mid.Int64)] = legacyGalaxyCell{g, s, pos}
			}
		}
		gRows.Close()
	}

	// na_building2planet
	bRows, err := db.QueryContext(ctx, `
		SELECT b.planetid, b.buildingid, b.level
		FROM na_building2planet b
		JOIN na_planet p ON p.planetid = b.planetid
		WHERE p.userid = ?`, userID)
	if err != nil {
		return nil, fmt.Errorf("select na_building2planet: %w", err)
	}
	for bRows.Next() {
		var b legacyBuilding
		if err := bRows.Scan(&b.planetid, &b.buildingid, &b.level); err != nil {
			bRows.Close()
			return nil, fmt.Errorf("scan na_building2planet: %w", err)
		}
		src.buildings = append(src.buildings, b)
	}
	bRows.Close()

	// na_research2user
	rRows, err := db.QueryContext(ctx, `
		SELECT buildingid, level FROM na_research2user WHERE userid = ?`, userID)
	if err != nil {
		return nil, fmt.Errorf("select na_research2user: %w", err)
	}
	for rRows.Next() {
		var r legacyResearch
		if err := rRows.Scan(&r.buildingid, &r.level); err != nil {
			rRows.Close()
			return nil, fmt.Errorf("scan na_research2user: %w", err)
		}
		src.researches = append(src.researches, r)
	}
	rRows.Close()

	// na_unit2shipyard — корабли, оборона, ракеты лежат вместе.
	sRows, err := db.QueryContext(ctx, `
		SELECT u.planetid, u.unitid, u.quantity, u.damaged, u.shell_percent
		FROM na_unit2shipyard u
		JOIN na_planet p ON p.planetid = u.planetid
		WHERE p.userid = ?`, userID)
	if err != nil {
		return nil, fmt.Errorf("select na_unit2shipyard: %w", err)
	}
	for sRows.Next() {
		var s legacyShipRow
		if err := sRows.Scan(&s.planetid, &s.unitid, &s.quantity,
			&s.damaged, &s.shellPercent); err != nil {
			sRows.Close()
			return nil, fmt.Errorf("scan na_unit2shipyard: %w", err)
		}
		src.ships = append(src.ships, s)
	}
	sRows.Close()

	// na_artefact2user — берём только не-удалённые на бирже (lot_id=0) и
	// принадлежащие юзеру.
	aRows, err := db.QueryContext(ctx, `
		SELECT artid, typeid, planetid, active, times_left, deleted
		FROM na_artefact2user
		WHERE userid = ? AND lot_id = 0`, userID)
	if err != nil {
		return nil, fmt.Errorf("select na_artefact2user: %w", err)
	}
	for aRows.Next() {
		var a legacyArtefact
		if err := aRows.Scan(&a.artid, &a.typeid, &a.planetid, &a.active,
			&a.timesLeft, &a.deleted); err != nil {
			aRows.Close()
			return nil, fmt.Errorf("scan na_artefact2user: %w", err)
		}
		src.artefacts = append(src.artefacts, a)
	}
	aRows.Close()

	// na_alliance + na_user2ally
	allyRow := db.QueryRowContext(ctx, `
		SELECT a.aid, a.tag, a.name, a.founder, a.textextern, a.textintern,
		       a.homepage, ua.rank, ua.joindate
		FROM na_user2ally ua
		JOIN na_alliance a ON a.aid = ua.aid
		WHERE ua.userid = ?
		LIMIT 1`, userID)
	var ally legacyAlliance
	var textextern, textintern, homepage []byte
	err = allyRow.Scan(&ally.aid, &ally.tag, &ally.name, &ally.founder,
		&textextern, &textintern, &homepage, &ally.memberRank, &ally.joinDate)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		// нет альянса
	case err != nil:
		return nil, fmt.Errorf("select na_alliance: %w", err)
	default:
		ally.textextern = string(textextern)
		ally.textintern = string(textintern)
		ally.homepage = string(homepage)
		src.alliance = &ally
	}

	// na_notes
	notesRow := db.QueryRowContext(ctx, `SELECT notes FROM na_notes WHERE user_id = ?`, userID)
	var notes sql.NullString
	if err := notesRow.Scan(&notes); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("select na_notes: %w", err)
	}
	if notes.Valid {
		src.notepad = notes.String
	}

	// na_message — limit на receiver=userID (входящие). В legacy sender может
	// быть NULL для системных сообщений.
	mRows, err := db.QueryContext(ctx, `
		SELECT msgid, mode, time, sender, receiver, message, subject, readed
		FROM na_message
		WHERE receiver = ?
		ORDER BY time DESC
		LIMIT 200`, userID)
	if err != nil {
		return nil, fmt.Errorf("select na_message: %w", err)
	}
	for mRows.Next() {
		var m legacyMessage
		if err := mRows.Scan(&m.msgid, &m.mode, &m.time, &m.sender, &m.receiver,
			&m.message, &m.subject, &m.readed); err != nil {
			mRows.Close()
			return nil, fmt.Errorf("scan na_message: %w", err)
		}
		src.messages = append(src.messages, m)
	}
	mRows.Close()

	return src, nil
}

// =====================================================================
// Cleanup
// =====================================================================

// cleanupTestUser удаляет ранее импортированного юзера в обоих БД, чтобы
// повторный запуск был идемпотентным. Связи с ON DELETE CASCADE удалят
// планеты/постройки/исследования/корабли/оборону/артефакты/messages
// автоматически.
//
// Альянс (alliances) НЕ удаляется — его могут делить другие юзеры. Если
// импортируемый юзер был owner — alliance остаётся «висеть»; для testing
// flow это допустимо (план 87 §«Что НЕ делаем»).
func cleanupTestUser(ctx context.Context, identityPool, gamePool *pgxpool.Pool, username string) error {
	// game-nova: ищем по username, удаляем по id (cascade добивает остальное).
	var gameID *string
	err := gamePool.QueryRow(ctx, `SELECT id::text FROM users WHERE username = $1`, username).Scan(&gameID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("lookup game user: %w", err)
	}
	if gameID != nil {
		// alliance_members привязан к юзеру, cascade удалит. Если юзер был
		// owner альянса — alliances.owner_id RESTRICT, надо вручную сменить
		// или удалить альянс. Здесь — удаляем alliances где он owner.
		if _, err := gamePool.Exec(ctx,
			`DELETE FROM alliances WHERE owner_id = $1`, *gameID); err != nil {
			return fmt.Errorf("delete owned alliances: %w", err)
		}
		if _, err := gamePool.Exec(ctx,
			`DELETE FROM users WHERE id = $1`, *gameID); err != nil {
			return fmt.Errorf("delete game user: %w", err)
		}
		slog.InfoContext(ctx, "cleanup: removed existing game user", slog.String("user_id", *gameID))
	}

	// identity: ищем по username, удаляем по id (cascade удалит refresh_tokens,
	// universe_memberships, user_consents).
	var identityID *string
	err = identityPool.QueryRow(ctx, `SELECT id::text FROM users WHERE username = $1`, username).Scan(&identityID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("lookup identity user: %w", err)
	}
	if identityID != nil {
		if _, err := identityPool.Exec(ctx,
			`DELETE FROM users WHERE id = $1`, *identityID); err != nil {
			return fmt.Errorf("delete identity user: %w", err)
		}
		slog.InfoContext(ctx, "cleanup: removed existing identity user", slog.String("user_id", *identityID))
	}

	return nil
}

// =====================================================================
// Identity-side insert
// =====================================================================

func insertIdentityUser(ctx context.Context, pool *pgxpool.Pool, userID string, cfg config) error {
	// identity-service ожидает argon2id (см. projects/identity/backend/internal/auth/password.go);
	// раньше CLI писал bcrypt, что давало 500 «invalid hash format» при логине.
	hash, err := auth.HashPassword(cfg.targetPassword)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	email := cfg.targetUsername + "@test.local"

	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `
		INSERT INTO users (id, username, email, password_hash, roles)
		VALUES ($1, $2, $3, $4, ARRAY['player'])`,
		userID, cfg.targetUsername, email, hash); err != nil {
		return fmt.Errorf("insert identity.users: %w", err)
	}

	// universe_memberships — игрок зарегистрирован в указанной вселенной.
	if _, err := tx.Exec(ctx, `
		INSERT INTO universe_memberships (user_id, universe_id) VALUES ($1, $2)`,
		userID, cfg.targetUniverse); err != nil {
		return fmt.Errorf("insert universe_memberships: %w", err)
	}

	// user_consents (план 44) — синтетическое согласие на обработку ПД +
	// принятие пользовательского соглашения. Без них identity-сервис может
	// требовать повторного подтверждения при логине.
	for _, ct := range []string{"pdn", "user_agreement"} {
		if _, err := tx.Exec(ctx, `
			INSERT INTO user_consents (user_id, consent_type, consent_text_version)
			VALUES ($1, $2, 'imported-from-legacy')`,
			userID, ct); err != nil {
			return fmt.Errorf("insert user_consents (%s): %w", ct, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit identity tx: %w", err)
	}
	slog.InfoContext(ctx, "identity user inserted",
		slog.String("user_id", userID),
		slog.String("username", cfg.targetUsername),
		slog.String("email", email))
	return nil
}

// =====================================================================
// Game-side insert
// =====================================================================

type insertStats struct {
	planets             int
	buildings           int
	research            int
	ships               int
	defense             int
	artefacts           int
	messages            int
	allianceMembers     int
	skippedUnknownUnits int
}

func insertGameUser(ctx context.Context, pool *pgxpool.Pool, userID string, cfg config, src *legacySource) (*insertStats, error) {
	stats := &insertStats{}

	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("begin game tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// users
	profession, ok := ProfessionMapping[src.user.profession]
	if !ok {
		profession = "none"
	}
	regtime := unixOrNow(src.user.regtime)
	lastSeen := unixOrNow(src.user.last)
	var protectedUntil *time.Time
	if src.user.protectionTime > 0 {
		t := time.Unix(src.user.protectionTime, 0).UTC()
		protectedUntil = &t
	}
	var lastTeleport *time.Time
	if src.user.planetTeleport > 0 {
		t := time.Unix(src.user.planetTeleport, 0).UTC()
		lastTeleport = &t
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO users (
			id, username, email, password_hash, language, timezone,
			points, max_points, credit, research_factor, ipcheck,
			umode, tutorial_state, regtime, last_seen, profession,
			is_observer, protected_until_at, last_planet_teleport_at,
			last_global_chat_read_at, last_ally_chat_read_at, created_at
		) VALUES ($1, $2, $3, NULL, 'ru', 'UTC',
			$4, $5, $6, $7, false,
			$8, $9, $10, $11, $12,
			$13, $14, $15,
			$16, $17, now())`,
		userID, cfg.targetUsername, cfg.targetUsername+"@test.local",
		src.user.points, src.user.maxPoints, src.user.credit, src.user.researchFactor,
		boolFromInt(src.user.umode), src.user.tutorialState, regtime, lastSeen, profession,
		boolFromInt(src.user.observer), protectedUntil, lastTeleport,
		unixOrZeroNullable(src.user.lastChat), unixOrZeroNullable(src.user.lastChatAlly),
	); err != nil {
		return nil, fmt.Errorf("insert game.users: %w", err)
	}

	// planets
	planetIDByLegacy := map[int]string{}
	for _, p := range src.planets {
		cell, ok := src.galaxy[p.planetid]
		if !ok {
			// Планета без cell в na_galaxy — orphan/удалённая. Пропускаем,
			// иначе UNIQUE(galaxy, system, position) рухнет.
			slog.WarnContext(ctx, "skip planet without galaxy cell",
				slog.Int("legacy_planetid", p.planetid),
				slog.String("name", p.planetname))
			continue
		}
		// nova ожидает координаты в диапазоне galaxy 1..16, system 1..999,
		// position 1..15. legacy может содержать большие system'ы; зажимаем.
		if cell.galaxy < 1 || cell.galaxy > 16 || cell.system < 1 || cell.system > 999 ||
			cell.position < 1 || cell.position > 15 {
			slog.WarnContext(ctx, "skip planet with out-of-range coords",
				slog.Int("legacy_planetid", p.planetid),
				slog.Int("galaxy", cell.galaxy),
				slog.Int("system", cell.system),
				slog.Int("position", cell.position))
			continue
		}
		pid := ids.New()
		planetIDByLegacy[p.planetid] = pid
		ptype := mapPlanetType(p.picture, p.ismoon == 1)
		// legacy temperature → nova temperature_min/max (диапазон ±20°C
		// относительно базовой; точная математика см. план 39).
		tempMin := p.temperature - 20
		tempMax := p.temperature + 20
		lastResUpdate := unixOrNow(p.last)
		if _, err := tx.Exec(ctx, `
			INSERT INTO planets (
				id, user_id, is_moon, name, galaxy, system, position,
				diameter, used_fields, temperature_min, temperature_max,
				metal, silicon, hydrogen, last_res_update,
				solar_satellite_prod, build_factor, research_factor,
				produce_factor, energy_factor, storage_factor,
				picture, planet_type, created_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7,
				$8, 0, $9, $10,
				$11, $12, $13, $14,
				$15, $16, $17,
				$18, $19, $20,
				$21, $22, now())`,
			pid, userID, p.ismoon == 1, p.planetname,
			cell.galaxy, cell.system, cell.position,
			p.diameter, tempMin, tempMax,
			p.metal, p.silicon, p.hydrogen, lastResUpdate,
			p.solarSatelliteProd, p.buildFactor, p.researchFactor,
			p.produceFactor, p.energyFactor, p.storageFactor,
			p.picture, ptype,
		); err != nil {
			return nil, fmt.Errorf("insert planet %d: %w", p.planetid, err)
		}
		stats.planets++
	}

	// cur_planet_id из na_user.curplanet
	if curID, ok := planetIDByLegacy[src.user.curPlanet]; ok {
		if _, err := tx.Exec(ctx,
			`UPDATE users SET cur_planet_id = $1 WHERE id = $2`, curID, userID); err != nil {
			return nil, fmt.Errorf("set cur_planet_id: %w", err)
		}
	} else if len(planetIDByLegacy) > 0 {
		// fallback на любую существующую планету
		for _, pid := range planetIDByLegacy {
			if _, err := tx.Exec(ctx,
				`UPDATE users SET cur_planet_id = $1 WHERE id = $2`, pid, userID); err != nil {
				return nil, fmt.Errorf("set cur_planet_id fallback: %w", err)
			}
			break
		}
	}

	// buildings + moon_buildings
	for _, b := range src.buildings {
		pid, ok := planetIDByLegacy[b.planetid]
		if !ok {
			continue
		}
		// Поищем сначала в обоих маппингах; lookup-результат — один и тот
		// же id, но это документирует тип здания. Если ни тут, ни там —
		// skip с warning.
		nid, ok := BuildingMapping[b.buildingid]
		if !ok {
			nid, ok = MoonBuildingMapping[b.buildingid]
		}
		if !ok {
			slog.WarnContext(ctx, "skip unknown building",
				slog.Int("legacy_buildingid", b.buildingid))
			stats.skippedUnknownUnits++
			continue
		}
		level := b.level
		if level < 0 {
			level = 0
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO buildings (planet_id, unit_id, level)
			VALUES ($1, $2, $3)
			ON CONFLICT (planet_id, unit_id) DO UPDATE SET level = EXCLUDED.level`,
			pid, nid, level); err != nil {
			return nil, fmt.Errorf("insert building: %w", err)
		}
		stats.buildings++
	}

	// research
	for _, r := range src.researches {
		nid, ok := ResearchMapping[r.buildingid]
		if !ok {
			slog.WarnContext(ctx, "skip unknown research",
				slog.Int("legacy_buildingid", r.buildingid))
			stats.skippedUnknownUnits++
			continue
		}
		level := r.level
		if level < 0 {
			level = 0
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO research (user_id, unit_id, level)
			VALUES ($1, $2, $3)
			ON CONFLICT (user_id, unit_id) DO UPDATE SET level = EXCLUDED.level`,
			userID, nid, level); err != nil {
			return nil, fmt.Errorf("insert research: %w", err)
		}
		stats.research++
	}

	// ships + defense + rockets + planet shields
	for _, s := range src.ships {
		pid, ok := planetIDByLegacy[s.planetid]
		if !ok {
			continue
		}
		// Маршрутизация по таблице по диапазону legacy unit-id:
		switch {
		case isInMap(FleetMapping, s.unitid):
			// ships
			shellPct := float32(s.shellPercent)
			if _, err := tx.Exec(ctx, `
				INSERT INTO ships (planet_id, unit_id, count, damaged_count, shell_percent)
				VALUES ($1, $2, $3, $4, $5)
				ON CONFLICT (planet_id, unit_id) DO UPDATE
				  SET count = EXCLUDED.count,
				      damaged_count = EXCLUDED.damaged_count,
				      shell_percent = EXCLUDED.shell_percent`,
				pid, FleetMapping[s.unitid], s.quantity, s.damaged, shellPct); err != nil {
				return nil, fmt.Errorf("insert ship: %w", err)
			}
			stats.ships++
		case isInMap(RocketMapping, s.unitid):
			// nova: ракеты живут в ships (план 22 Ф.2.3)
			if _, err := tx.Exec(ctx, `
				INSERT INTO ships (planet_id, unit_id, count, damaged_count, shell_percent)
				VALUES ($1, $2, $3, 0, 0)
				ON CONFLICT (planet_id, unit_id) DO UPDATE SET count = EXCLUDED.count`,
				pid, RocketMapping[s.unitid], s.quantity); err != nil {
				return nil, fmt.Errorf("insert rocket: %w", err)
			}
			stats.ships++
		case isInMap(DefenseMapping, s.unitid):
			if _, err := tx.Exec(ctx, `
				INSERT INTO defense (planet_id, unit_id, count)
				VALUES ($1, $2, $3)
				ON CONFLICT (planet_id, unit_id) DO UPDATE SET count = EXCLUDED.count`,
				pid, DefenseMapping[s.unitid], s.quantity); err != nil {
				return nil, fmt.Errorf("insert defense: %w", err)
			}
			stats.defense++
		case isInMap(PlanetShieldMapping, s.unitid):
			if _, err := tx.Exec(ctx, `
				INSERT INTO defense (planet_id, unit_id, count)
				VALUES ($1, $2, $3)
				ON CONFLICT (planet_id, unit_id) DO UPDATE SET count = EXCLUDED.count`,
				pid, PlanetShieldMapping[s.unitid], s.quantity); err != nil {
				return nil, fmt.Errorf("insert planet shield: %w", err)
			}
			stats.defense++
		default:
			slog.WarnContext(ctx, "skip unknown unit (not ship/defense/rocket)",
				slog.Int("legacy_unitid", s.unitid))
			stats.skippedUnknownUnits++
		}
	}

	// artefacts
	for _, a := range src.artefacts {
		nid, ok := ArtefactMapping[a.typeid]
		if !ok {
			slog.WarnContext(ctx, "skip artefact: type not in nova catalog",
				slog.Int("legacy_typeid", a.typeid))
			stats.skippedUnknownUnits++
			continue
		}
		state := "held"
		if a.active != 0 {
			state = "active"
		} else if a.deleted != 0 {
			state = "expired"
		}
		var pid *string
		if a.planetid > 0 {
			if id, ok := planetIDByLegacy[a.planetid]; ok {
				pid = &id
			}
		}
		artID := ids.New()
		if _, err := tx.Exec(ctx, `
			INSERT INTO artefacts_user (id, user_id, planet_id, unit_id, state, acquired_at, payload)
			VALUES ($1, $2, $3, $4, $5, now(), '{}'::jsonb)`,
			artID, userID, pid, nid, state); err != nil {
			return nil, fmt.Errorf("insert artefact: %w", err)
		}
		stats.artefacts++
	}

	// alliance + member
	if src.alliance != nil {
		// alliances.tag должен быть в диапазоне 3-5 символов; legacy
		// не валидирует, поэтому прайвинг.
		tag := strings.TrimSpace(src.alliance.tag)
		if tag == "" {
			tag = "LGCY"
		}
		if len(tag) > 5 {
			tag = tag[:5]
		}
		// Имя альянса в nova UNIQUE; для предотвращения коллизий с
		// существующими альянсами в БД суффиксуем коротким хексом.
		name := strings.TrimSpace(src.alliance.name)
		if name == "" {
			name = "Imported alliance"
		}
		suffix, _ := randHex(4)
		nameUniq := fmt.Sprintf("%s [%s]", name, suffix)
		tagUniq := tag
		if len(tagUniq) > 0 {
			// делаем tag уникальным аналогично — добавляем 1 hex-символ если
			// длина <=4. Иначе оставляем как есть; UNIQUE-конфликт всё равно
			// ловится через ON CONFLICT (но nova таблица ON CONFLICT не имеет
			// для UNIQUE поля, поэтому используем простой суффикс).
			if len(tagUniq) <= 4 {
				h, _ := randHex(1)
				tagUniq = tagUniq + h
			}
		}
		aid := uuid.NewString() // alliances.id имеет DEFAULT gen_random_uuid(), но передаём явно для cleanup
		if _, err := tx.Exec(ctx, `
			INSERT INTO alliances (id, tag, name, description, owner_id)
			VALUES ($1, $2, $3, $4, $5)`,
			aid, tagUniq, nameUniq, src.alliance.textextern, userID); err != nil {
			return nil, fmt.Errorf("insert alliance: %w", err)
		}
		// rank: 0 в legacy = founder/owner; >0 — обычный member.
		rank := "member"
		if src.alliance.memberRank == 0 || src.alliance.founder == src.user.id {
			rank = "owner"
		}
		joinedAt := unixOrNow(src.alliance.joinDate)
		if _, err := tx.Exec(ctx, `
			INSERT INTO alliance_members (alliance_id, user_id, rank, joined_at)
			VALUES ($1, $2, $3, $4)`,
			aid, userID, rank, joinedAt); err != nil {
			return nil, fmt.Errorf("insert alliance_member: %w", err)
		}
		if _, err := tx.Exec(ctx,
			`UPDATE users SET alliance_id = $1 WHERE id = $2`, aid, userID); err != nil {
			return nil, fmt.Errorf("set users.alliance_id: %w", err)
		}
		stats.allianceMembers = 1
	}

	// notepad
	if src.notepad != "" {
		// 0078_notepad_check_length добавил CHECK на размер; зажимаем.
		content := src.notepad
		const maxNotepad = 64 * 1024
		if len(content) > maxNotepad {
			content = content[:maxNotepad]
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO user_notepad (user_id, content) VALUES ($1, $2)
			ON CONFLICT (user_id) DO UPDATE SET content = EXCLUDED.content`,
			userID, content); err != nil {
			return nil, fmt.Errorf("insert notepad: %w", err)
		}
	}

	// messages — берём не более messageLimit последних, складываем в inbox
	// (folder=0). sender NULL → from_user_id NULL (системные).
	limit := cfg.messageLimit
	if limit > len(src.messages) {
		limit = len(src.messages)
	}
	for i := 0; i < limit; i++ {
		m := src.messages[i]
		mid := ids.New()
		// Тело legacy содержит HTML/BBCode; nova messages хранит это «как
		// есть» — origin-фронт умеет их рендерить.
		body := m.message
		const maxBody = 32 * 1024
		if len(body) > maxBody {
			body = body[:maxBody]
		}
		subject := m.subject
		if len(subject) > 256 {
			subject = subject[:256]
		}
		var readAt *time.Time
		if m.readed != 0 {
			t := time.Unix(m.time, 0).UTC()
			readAt = &t
		}
		createdAt := unixOrNow(m.time)
		if _, err := tx.Exec(ctx, `
			INSERT INTO messages (id, to_user_id, from_user_id, folder, subject, body, created_at, read_at)
			VALUES ($1, $2, NULL, $3, $4, $5, $6, $7)`,
			mid, userID, m.mode, subject, body, createdAt, readAt); err != nil {
			return nil, fmt.Errorf("insert message: %w", err)
		}
		stats.messages++
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit game tx: %w", err)
	}
	return stats, nil
}

// =====================================================================
// Helpers
// =====================================================================

func isInMap(m map[int]int, k int) bool {
	_, ok := m[k]
	return ok
}

func boolFromInt(i int) bool { return i != 0 }

func unixOrNow(ts int64) time.Time {
	if ts <= 0 {
		return time.Now().UTC()
	}
	return time.Unix(ts, 0).UTC()
}

func unixOrZeroNullable(ts int64) *time.Time {
	if ts <= 0 {
		return nil
	}
	t := time.Unix(ts, 0).UTC()
	return &t
}

// mapPlanetType конвертирует legacy `na_planet.picture` ("dschjungelplanet05",
// "moon", "wasserplanet09") в nova `planet_type` (без числового суффикса).
// Если picture не распознан — fallback по is_moon флагу.
func mapPlanetType(picture string, isMoon bool) string {
	pic := strings.ToLower(strings.TrimSpace(picture))
	// snip trailing digits ("dschjungelplanet05" → "dschjungelplanet")
	for i := len(pic); i > 0; i-- {
		c := pic[i-1]
		if c < '0' || c > '9' {
			pic = pic[:i]
			break
		}
	}
	if t, ok := LegacyPlanetTypeMapping[pic]; ok {
		return t
	}
	if isMoon {
		return "moon"
	}
	return "normaltempplanet"
}

func randHex(nBytes int) (string, error) {
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
