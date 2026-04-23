// Command server — HTTP/WS вход oxsar-nova. Запускает API на SERVER_ADDR.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/oxsar/nova/backend/internal/achievement"
	"github.com/oxsar/nova/backend/internal/admin"
	"github.com/oxsar/nova/backend/internal/alliance"
	"github.com/oxsar/nova/backend/internal/chat"
	"github.com/oxsar/nova/backend/internal/artefact"
	"github.com/oxsar/nova/backend/internal/artmarket"
	"github.com/oxsar/nova/backend/internal/auth"
	"github.com/oxsar/nova/backend/internal/automsg"
	"github.com/oxsar/nova/backend/internal/battle"
	"github.com/oxsar/nova/backend/internal/building"
	"github.com/oxsar/nova/backend/internal/config"
	"github.com/oxsar/nova/backend/internal/fleet"
	"github.com/oxsar/nova/backend/internal/galaxy"
	"github.com/oxsar/nova/backend/internal/httpx"
	"github.com/oxsar/nova/backend/internal/i18n"
	"github.com/oxsar/nova/backend/internal/market"
	"github.com/oxsar/nova/backend/internal/message"
	"github.com/oxsar/nova/backend/internal/officer"
	"github.com/oxsar/nova/backend/internal/planet"
	"github.com/oxsar/nova/backend/internal/repair"
	"github.com/oxsar/nova/backend/internal/repo"
	"github.com/oxsar/nova/backend/internal/requirements"
	"github.com/oxsar/nova/backend/internal/research"
	"github.com/oxsar/nova/backend/internal/rocket"
	"github.com/oxsar/nova/backend/internal/score"
	"github.com/oxsar/nova/backend/internal/shipyard"
	"github.com/oxsar/nova/backend/internal/storage"
)

func main() {
	if err := run(); err != nil {
		slog.Error("server exit", slog.String("err", err.Error()))
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	log := newLogger(cfg.Server.LogLevel)
	slog.SetDefault(log)
	log.InfoContext(ctx, "starting server", slog.String("env", cfg.Server.Env))

	catalogDir := os.Getenv("CATALOG_DIR")
	if catalogDir == "" {
		catalogDir = "../configs"
	}
	cat, err := config.LoadCatalog(catalogDir)
	if err != nil {
		return err
	}
	log.InfoContext(ctx, "catalog loaded",
		slog.Int("buildings", len(cat.Buildings.Buildings)),
		slog.Int("ships", len(cat.Ships.Ships)))

	pool, err := storage.OpenPostgres(ctx, cfg.DB.URL)
	if err != nil {
		return err
	}
	defer pool.Close()

	rdb, err := storage.OpenRedis(ctx, cfg.Redis.URL)
	if err != nil {
		log.WarnContext(ctx, "redis unavailable, continuing without cache", slog.String("err", err.Error()))
	}

	// Rate-limit для /api/auth/login и /api/auth/register: 20 req/min per IP.
	authRL := auth.NewIPRateLimiter(rdb, 20, time.Minute)

	db := repo.New(pool)

	jwt := auth.NewJWTIssuer(cfg.Auth.JWTSecret, cfg.Auth.AccessTTL, cfg.Auth.RefreshTTL)

	planetRepo := planet.NewRepository(pool)
	planetSvc := planet.NewService(db, planetRepo, cat)
	planetH := planet.NewHandler(planetSvc)
	starter := planet.NewStarter(db)

	// automsg нужен auth (WELCOME/STARTER_GUIDE при регистрации),
	// поэтому инициализируем до auth.NewService.
	automsgSvc := automsg.NewService(db)

	authSvc := auth.NewService(db, jwt, starter, automsgSvc)
	authH := auth.NewHandler(authSvc, pool)

	reqs := requirements.New(cat)

	buildingSvc := building.NewService(db, planetSvc, cat, reqs, cfg.Game.Speed)
	buildingH := building.NewHandler(buildingSvc)

	researchSvc := research.NewService(db, planetSvc, cat, reqs, cfg.Game.Speed)
	researchH := research.NewHandler(researchSvc)

	shipyardSvc := shipyard.NewService(db, planetSvc, cat, reqs, cfg.Game.Speed)
	shipyardH := shipyard.NewHandler(shipyardSvc)

	repairSvc := repair.NewService(db, planetSvc, cat, reqs, cfg.Game.Speed)
	repairH := repair.NewHandler(repairSvc)

	artefactSvc := artefact.NewService(db, cat)
	artefactH := artefact.NewHandler(artefactSvc)

	galaxyH := galaxy.NewHandler(galaxy.NewRepository(pool))

	transportSvc := fleet.NewTransportService(db, cat, cfg.Game.Speed, artefactSvc)
	fleetH := fleet.NewHandler(transportSvc)

	messageSvc := message.NewService(db)
	messageH := message.NewHandler(messageSvc)

	marketSvc := market.NewService(db)
	marketH := market.NewHandler(marketSvc)

	rocketSvc := rocket.NewService(db, cat, cfg.Game.Speed)
	rocketH := rocket.NewHandler(rocketSvc)

	artMarketSvc := artmarket.NewService(db)
	artMarketH := artmarket.NewHandler(artMarketSvc)

	achSvc := achievement.NewService(db)
	achH := achievement.NewHandler(achSvc)

	officerSvc := officer.NewService(db)
	officerH := officer.NewHandler(officerSvc)

	scoreSvc := score.NewService(db, cat)
	scoreH := score.NewHandlerWithDB(scoreSvc, db)

	allianceSvc := alliance.NewService(db)
	allianceH := alliance.NewHandler(allianceSvc)

	adminH := admin.NewHandler(db)

	chatHub := chat.NewHub()
	chatH := chat.NewHandler(chatHub, db)

	// i18n: папка необязательна — если её нет или пустая, i18n просто
	// пропускается, HTTP-эндпоинты не регистрируются. Это ожидаемо
	// до первого прогона cmd/tools/import-phrases (§10.3 ТЗ).
	i18nDir := os.Getenv("I18N_DIR")
	if i18nDir == "" {
		i18nDir = filepath.Join(catalogDir, "i18n")
	}
	var i18nH *i18n.Handler
	if bundle, err := i18n.Load(i18nDir, i18n.LangRu); err != nil {
		log.WarnContext(ctx, "i18n not loaded", slog.String("dir", i18nDir), slog.String("err", err.Error()))
	} else {
		log.InfoContext(ctx, "i18n loaded", slog.Any("langs", bundle.Languages()))
		i18nH = i18n.NewHandler(bundle)
	}

	r := httpx.NewRouter(httpx.RouterDeps{Log: log})

	r.With(authRL.Middleware).Post("/api/auth/register", authH.Register)
	r.With(authRL.Middleware).Post("/api/auth/login", authH.Login)
	r.With(authRL.Middleware).Post("/api/auth/refresh", authH.Refresh)
	r.With(auth.Middleware(jwt)).Get("/api/me", authH.Me)
	r.Post("/api/battle-sim", battleSimHandler)
	r.Get("/api/stats", scoreH.Stats)

	// i18n доступна без авторизации (логин-экран тоже использует).
	if i18nH != nil {
		r.Get("/api/i18n", i18nH.Languages)
		r.Get("/api/i18n/{lang}", i18nH.Locale)
	}

	r.Route("/api", func(pr chi.Router) {
		pr.Use(auth.Middleware(jwt))
		pr.Use(auth.LastSeenMiddleware(pool))
		pr.Get("/planets", planetH.List)
		pr.Get("/planets/{id}", planetH.Get)
		pr.Patch("/planets/{id}", planetH.Rename)
		pr.Post("/planets/{id}/set-home", planetH.SetHome)
		pr.Delete("/planets/{id}", planetH.Abandon)
		pr.Get("/planets/{id}/resource-report", planetH.ResourceReport)
		pr.Post("/planets/{id}/resource-update", planetH.ResourceUpdate)

		pr.Get("/planets/{id}/buildings", buildingH.Levels)
		pr.Post("/planets/{id}/buildings", buildingH.Enqueue)
		pr.Get("/planets/{id}/buildings/queue", buildingH.List)
		pr.Delete("/planets/{id}/buildings/queue/{taskId}", buildingH.Cancel)

		pr.Post("/planets/{id}/research", researchH.Enqueue)
		pr.Get("/research", researchH.List)

		pr.Post("/planets/{id}/shipyard", shipyardH.Enqueue)
		pr.Get("/planets/{id}/shipyard/queue", shipyardH.List)
		pr.Get("/planets/{id}/shipyard/inventory", shipyardH.Inventory)
		pr.Delete("/planets/{id}/shipyard/{queueId}", shipyardH.Cancel)

		pr.Post("/planets/{id}/repair/disassemble", repairH.EnqueueDisassemble)
		pr.Post("/planets/{id}/repair/repair", repairH.EnqueueRepair)
		pr.Get("/planets/{id}/repair/damaged", repairH.ListDamaged)
		pr.Get("/planets/{id}/repair/queue", repairH.List)
		pr.Delete("/planets/{id}/repair/queue/{queueId}", repairH.Cancel)

		pr.Get("/artefacts", artefactH.List)
		pr.Post("/artefacts/{id}/activate", artefactH.Activate)
		pr.Post("/artefacts/{id}/deactivate", artefactH.Deactivate)
		pr.Post("/artefacts/{id}/sell", artMarketH.ListForSale)

		pr.Get("/artefact-market/offers", artMarketH.Offers)
		pr.Get("/artefact-market/credit", artMarketH.Credit)
		pr.Post("/artefact-market/offers/{id}/buy", artMarketH.Buy)
		pr.Delete("/artefact-market/offers/{id}", artMarketH.Cancel)

		pr.Get("/achievements", achH.List)

		pr.Get("/officers", officerH.List)
		pr.Post("/officers/{key}/activate", officerH.Activate)

		pr.Get("/galaxy/{g}/{s}", galaxyH.System)

		pr.Post("/fleet", fleetH.Send)
		pr.Get("/fleet", fleetH.List)
		pr.Get("/fleet/incoming", fleetH.Incoming)
		pr.Post("/fleet/{id}/recall", fleetH.Recall)

		pr.Get("/market/rates", marketH.Rates)
		pr.Post("/planets/{id}/market/exchange", marketH.Exchange)
		pr.Get("/market/lots", marketH.ListLots)
		pr.Post("/market/lots", marketH.CreateLot)
		pr.Delete("/market/lots/{id}", marketH.CancelLot)
		pr.Post("/market/lots/{id}/accept", marketH.AcceptLot)

		pr.Post("/planets/{id}/rockets/launch", rocketH.Launch)
		pr.Get("/planets/{id}/rockets", rocketH.Stock)

		pr.Get("/highscore", scoreH.Highscore)
		pr.Get("/highscore/me", scoreH.MyRank)

		pr.Get("/alliances", allianceH.List)
		pr.Get("/alliances/me", allianceH.My)
		pr.Get("/alliances/{id}", allianceH.Get)
		pr.Get("/alliances/{id}/applications", allianceH.Applications)
		pr.Post("/alliances", allianceH.Create)
		pr.Post("/alliances/{id}/join", allianceH.Join)
		pr.Patch("/alliances/{id}/open", allianceH.SetOpen)
		pr.Post("/alliances/leave", allianceH.Leave)
		pr.Delete("/alliances/{id}", allianceH.Disband)
		pr.Post("/alliances/applications/{appID}/approve", allianceH.Approve)
		pr.Delete("/alliances/applications/{appID}", allianceH.Reject)
		pr.Get("/alliances/{id}/relations", allianceH.GetRelations)
		pr.Put("/alliances/{id}/relations/{target_id}", allianceH.ProposeRelation)
		pr.Post("/alliances/{id}/relations/{initiator_id}/accept", allianceH.AcceptRelation)
		pr.Delete("/alliances/{id}/relations/{initiator_id}", allianceH.RejectRelation)
		pr.Patch("/alliances/{id}/members/{userID}/rank", allianceH.SetMemberRank)

		pr.Get("/chat/{kind}/history", chatH.History)
		pr.Post("/chat/{kind}/send", chatH.Send)
		pr.Get("/chat/{kind}/ws", chatH.Connect)
		pr.Patch("/chat/messages/{id}", chatH.EditMessage)
		pr.Delete("/chat/messages/{id}", chatH.DeleteMessage)

		pr.Get("/messages", messageH.Inbox)
		pr.Post("/messages", messageH.Compose)
		pr.Delete("/messages", messageH.DeleteAll)
		pr.Delete("/messages/{id}", messageH.Delete)
		pr.Get("/messages/unread-count", messageH.UnreadCount)
		pr.Post("/messages/{id}/read", messageH.MarkRead)
		pr.Get("/battle-reports/{id}", messageH.GetReport)
		pr.Get("/espionage-reports/{id}", messageH.GetEspionageReport)
		pr.Get("/expedition-reports/{id}", messageH.GetExpeditionReport)

		pr.Route("/admin", func(ar chi.Router) {
			ar.Use(admin.AdminOnly(db))
			ar.Get("/stats", adminH.Stats)
			ar.Get("/users", adminH.ListUsers)
			ar.Post("/users/{id}/ban", adminH.Ban)
			ar.Post("/users/{id}/unban", adminH.Unban)
			ar.Post("/users/{id}/credit", adminH.Credit)
			ar.Post("/users/{id}/role", adminH.SetRole)
			ar.Get("/automsgs", adminH.ListAutomsgs)
			ar.Put("/automsgs/{key}", adminH.UpdateAutomsg)
		})
	})

	srv := &http.Server{
		Addr:              cfg.Server.Addr,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.InfoContext(ctx, "listening", slog.String("addr", srv.Addr))
		errCh <- srv.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		log.InfoContext(ctx, "shutdown requested")
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}

func newLogger(level string) *slog.Logger {
	lvl := slog.LevelInfo
	switch level {
	case "debug":
		lvl = slog.LevelDebug
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lvl}))
}

// battleSimHandler — офлайновый симулятор (§5.7 ТЗ). Чистая функция:
// вход/выход полностью через JSON, БД не затрагивается.
// Если NumSim >= 2, возвращает SimStats вместо Report.
func battleSimHandler(w http.ResponseWriter, r *http.Request) {
	var in battle.Input
	if err := decodeJSON(r, &in); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
		return
	}
	n := in.NumSim
	if n < 2 {
		report, err := battle.Calculate(in)
		if err != nil {
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
			return
		}
		httpx.WriteJSON(w, r, http.StatusOK, report)
		return
	}
	if n > 20 {
		n = 20
	}
	var wins, draws int
	var totalRounds int
	seed0 := in.Seed
	for i := range n {
		in.Seed = seed0 + uint64(i)
		rep, err := battle.Calculate(in)
		if err != nil {
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
			return
		}
		totalRounds += rep.Rounds
		switch rep.Winner {
		case "attackers":
			wins++
		case "draw":
			draws++
		}
	}
	stats := battle.SimStats{
		NumSim:    n,
		WinRate:   float64(wins) / float64(n),
		DrawRate:  float64(draws) / float64(n),
		AvgRounds: float64(totalRounds) / float64(n),
	}
	httpx.WriteJSON(w, r, http.StatusOK, stats)
}

func decodeJSON(r *http.Request, into any) error {
	dec := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	dec.DisallowUnknownFields()
	return dec.Decode(into)
}
