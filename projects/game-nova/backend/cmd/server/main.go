// Command server — HTTP/WS вход oxsar-nova. Запускает API на SERVER_ADDR.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"

	"oxsar/game-nova/internal/achievement"
	"oxsar/game-nova/internal/admin"
	"oxsar/game-nova/internal/aiadvisor"
	"oxsar/game-nova/internal/alien"
	"oxsar/game-nova/internal/alliance"
	"oxsar/game-nova/internal/chat"
	"oxsar/game-nova/internal/artefact"
	"oxsar/game-nova/internal/artmarket"
	"oxsar/game-nova/internal/auth"
	"oxsar/game-nova/internal/automsg"
	"oxsar/game-nova/internal/battle"
	"oxsar/game-nova/internal/battlestats"
	"oxsar/game-nova/internal/building"
	"oxsar/game-nova/internal/config"
	"oxsar/game-nova/internal/empire"
	"oxsar/game-nova/internal/fleet"
	"oxsar/game-nova/internal/friends"
	"oxsar/game-nova/internal/galaxy"
	"oxsar/game-nova/internal/httpx"
	"oxsar/game-nova/internal/i18n"
	"oxsar/game-nova/internal/market"
	"oxsar/game-nova/internal/message"
	"oxsar/game-nova/internal/moderation"
	"oxsar/game-nova/internal/notepad"
	"oxsar/game-nova/internal/officer"
	"oxsar/game-nova/internal/planet"
	"oxsar/game-nova/internal/profession"
	"oxsar/game-nova/internal/records"
	"oxsar/game-nova/internal/referral"
	"oxsar/game-nova/internal/report"
	"oxsar/game-nova/internal/repair"
	"oxsar/game-nova/internal/repo"
	"oxsar/game-nova/internal/requirements"
	"oxsar/game-nova/internal/research"
	"oxsar/game-nova/internal/rocket"
	"oxsar/game-nova/internal/score"
	"oxsar/game-nova/internal/search"
	"oxsar/game-nova/internal/settings"
	"oxsar/game-nova/internal/shipyard"
	"oxsar/game-nova/internal/techtree"
	"oxsar/game-nova/internal/storage"
	"oxsar/game-nova/internal/dailyquest"
	"oxsar/game-nova/internal/features"
	"oxsar/game-nova/internal/galaxyevent"
	"oxsar/game-nova/internal/health"
	"oxsar/game-nova/internal/universe"
	"oxsar/game-nova/internal/universeswitcher"
	"oxsar/game-nova/internal/wiki"
	"oxsar/game-nova/pkg/metrics"
)

// buildVersion — версия билда. Можно перебить через -ldflags
// "-X main.buildVersion=1.2.3" в build-pipeline; по умолчанию dev.
var buildVersion = "dev"

// drainDelay — пауза между SetDraining и srv.Shutdown. Цель —
// дать nginx/балансировщику время убрать backend из upstream
// (типичный health-check-interval = 5-10s).
const drainDelay = 10 * time.Second

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

	healthState := health.NewState("backend", buildVersion)

	catalogDir := os.Getenv("CATALOG_DIR")
	if catalogDir == "" {
		catalogDir = "../configs"
	}

	univReg, univRegErr := universe.NewRegistry(filepath.Join(catalogDir, "universes.yaml"))
	if univRegErr != nil {
		log.WarnContext(ctx, "universes.yaml not loaded", slog.String("err", univRegErr.Error()))
		univReg, _ = universe.NewRegistryFromSlice(nil)
	}

	cat, err := config.LoadCatalog(catalogDir)
	if err != nil {
		return err
	}
	log.InfoContext(ctx, "catalog loaded",
		slog.Int("buildings", len(cat.Buildings.Buildings)),
		slog.Int("ships", len(cat.Ships.Ships)))

	// Feature flags. Отсутствие файла — не ошибка (все флаги false). План 31 Ф.2.
	featuresPath := os.Getenv("FEATURES_FILE")
	if featuresPath == "" {
		featuresPath = filepath.Join(catalogDir, "features.yaml")
	}
	featureSet, err := features.Load(featuresPath)
	if err != nil {
		log.WarnContext(ctx, "features load failed, using empty set",
			slog.String("path", featuresPath), slog.String("err", err.Error()))
		featureSet, _ = features.ParseBytes(nil)
	}
	log.InfoContext(ctx, "features loaded",
		slog.String("path", featuresPath),
		slog.Any("enabled", features.EnabledKeys(featureSet)))
	featureH := features.NewHandler(featureSet)

	pool, err := storage.OpenPostgres(ctx, cfg.DB.URL)
	if err != nil {
		return err
	}
	defer pool.Close()

	rdb, err := storage.OpenRedis(ctx, cfg.Redis.URL)
	if err != nil {
		log.WarnContext(ctx, "redis unavailable, continuing without cache", slog.String("err", err.Error()))
	}

	// План 36 Ф.12: /api/auth/login|register|refresh удалены —
	// аутентификация только через identity-service.

	db := repo.New(pool)

	// IDENTITY_JWKS_URL обязателен (legacy AUTH_JWKS_URL читается как
	// fallback, см. config.envFallback). Запуск без него = fail-fast.
	if cfg.Auth.JWKSUrl == "" {
		return fmt.Errorf("IDENTITY_JWKS_URL is required (legacy AUTH_JWKS_URL also accepted)")
	}
	log.InfoContext(ctx, "auth mode: RSA-256 via JWKS",
		slog.String("jwks_url", cfg.Auth.JWKSUrl),
		slog.String("universe_id", cfg.Auth.UniverseID))


	planetRepo := planet.NewRepository(pool)
	planetSvc := planet.NewServiceWithFactors(db, planetRepo, cat, cfg.Game.StorageFactor, cfg.Game.EnergyProductionFactor)

	// План 17 F: galaxy events. Wire-up в planet до handler.
	galaxyEventSvc := galaxyevent.New(pool)
	galaxyEventH := galaxyevent.NewHandler(galaxyEventSvc)
	planetSvc.SetGalaxyEventReader(galaxyEventSvc)

	planetH := planet.NewHandler(planetSvc)
	starter := planet.NewStarter(db)

	// i18n bundle — загружаем до сервисов, чтобы automsg мог читать
	// тексты шаблонов. Если директория не найдена — bundle=nil,
	// automsg.Send вернёт ErrNoBundle (нет критического падения).
	i18nDir := os.Getenv("I18N_DIR")
	if i18nDir == "" {
		i18nDir = filepath.Join(catalogDir, "i18n")
	}
	var i18nBundle *i18n.Bundle
	var i18nH *i18n.Handler
	if bundle, err := i18n.Load(i18nDir, i18n.LangRu); err != nil {
		log.WarnContext(ctx, "i18n not loaded", slog.String("dir", i18nDir), slog.String("err", err.Error()))
	} else {
		log.InfoContext(ctx, "i18n loaded", slog.Any("langs", bundle.Languages()))
		i18nBundle = bundle
		i18nH = i18n.NewHandler(bundle)
	}

	// automsg нужен auth (welcome/starterGuide при регистрации),
	// поэтому инициализируем до auth.NewService.
	automsgSvc := automsg.NewService(db).WithBundle(i18nBundle)

	referralSvc := referral.NewService(db).WithAutoMsg(automsgSvc).WithBundle(i18nBundle)
	_ = referralSvc // referral больше не подключается через auth.Service (она удалена в Ф.12);
	// referral будет вызываться напрямую из EnsureUserMiddleware/bootstrap (TODO).

	// План 36 Ф.12: auth.Service (Register/Login/Refresh) удалён. Здесь только
	// /api/me и /api/me/vacation — handler берёт данные напрямую из БД.
	authH := auth.NewHandler(pool)

	// authMiddleware: RSA-валидация JWT от auth-service + lazy-create юзера
	// в game-db при первом запросе (план 36 Ф.12).
	rsaVer, loadErr := auth.LoadVerifier(ctx, cfg.Auth.JWKSUrl)
	if loadErr != nil {
		return loadErr
	}
	rsaMW := auth.RSAMiddleware(rsaVer)
	ensureMW := auth.EnsureUserMiddleware(auth.EnsureUserConfig{
		Pool:           pool,
		Starter:        starter,
		Automsg:        automsgSvc,
		UniverseID:     cfg.Auth.UniverseID,
		AuthServiceURL: cfg.Auth.AuthServiceURL,
	})
	authMiddlewareFn := func(next http.Handler) http.Handler {
		return rsaMW(ensureMW(next))
	}

	switcherH := universeswitcher.New(cfg.Auth.AuthServiceURL, cfg.Auth.UniverseID, univReg)

	reqs := requirements.New(cat)

	buildingSvc := building.NewService(db, planetSvc, cat, reqs, cfg.Game.Speed)
	buildingH := building.NewHandler(buildingSvc)

	researchSvc := research.NewServiceWithFactors(db, planetSvc, cat, reqs, cfg.Game.Speed, cfg.Game.ResearchSpeedFactor)
	researchH := research.NewHandler(researchSvc)

	shipyardSvc := shipyard.NewService(db, planetSvc, cat, reqs, cfg.Game.Speed)
	shipyardH := shipyard.NewHandler(shipyardSvc)

	repairSvc := repair.NewService(db, planetSvc, cat, reqs, cfg.Game.Speed)
	repairH := repair.NewHandler(repairSvc)

	artefactSvc := artefact.NewService(db, cat).WithAutoMsg(automsgSvc).WithBundle(i18nBundle)
	artefactH := artefact.NewHandler(artefactSvc)

	galaxyH := galaxy.NewHandler(galaxy.NewRepository(pool))

	dailyQuestSvc := dailyquest.New(pool)
	dailyQuestH := dailyquest.NewHandler(dailyQuestSvc)

	transportSvc := fleet.NewTransportServiceWithConfig(db, cat, cfg.Game.Speed, artefactSvc, cfg.Game.MaxPlanets, cfg.Game.ProtectionPeriod).WithBundle(i18nBundle)
	transportSvc.SetBashingLimits(cfg.Game.BashingPeriod, cfg.Game.BashingMaxAttacks)
	transportSvc.SetDailyQuestSvc(dailyQuestSvc)
	fleetH := fleet.NewHandler(transportSvc, rdb)

	messageSvc := message.NewService(db)
	messageH := message.NewHandler(messageSvc)

	marketSvc := market.NewService(db)
	marketH := market.NewHandler(marketSvc, rdb)

	rocketSvc := rocket.NewService(db, cat, cfg.Game.Speed).WithBundle(i18nBundle)
	rocketH := rocket.NewHandler(rocketSvc)

	artMarketSvc := artmarket.NewService(db)
	artMarketH := artmarket.NewHandler(artMarketSvc, rdb)

	achSvc := achievement.NewService(db).WithBundle(i18nBundle)
	achH := achievement.NewHandler(achSvc)

	officerSvc := officer.NewService(db).WithBundle(i18nBundle)
	officerH := officer.NewHandler(officerSvc)

	scoreSvc := score.NewServiceWithCoeffs(db, cat, cfg.Game.Points)
	scoreH := score.NewHandlerWithDB(scoreSvc, db)

	allianceSvc := alliance.NewService(db).WithAutoMsg(automsgSvc).WithBundle(i18nBundle)

	// План 46 (149-ФЗ): UGC-blacklist для tag/name альянса и для чата.
	// Путь — env MODERATION_BLACKLIST (default: общий конфиг в корне репо).
	// Отсутствие файла — warning, не fatal (на dev/test допустимо).
	blPath := os.Getenv("MODERATION_BLACKLIST")
	if blPath == "" {
		blPath = filepath.Join(catalogDir, "moderation", "blacklist.yaml")
	}
	var ugcBlacklist *moderation.Blacklist
	if bl, blErr := moderation.LoadBlacklist(blPath); blErr == nil {
		ugcBlacklist = bl
		allianceSvc = allianceSvc.WithBlacklist(bl)
		log.InfoContext(ctx, "moderation blacklist loaded",
			slog.String("path", blPath), slog.Int("roots", bl.Size()))
	} else {
		log.WarnContext(ctx, "moderation blacklist not loaded; UGC checks disabled",
			slog.String("path", blPath), slog.String("err", blErr.Error()))
	}

	allianceH := alliance.NewHandler(allianceSvc)

	// План 46 Ф.3 (149-ФЗ): жалобы пользователей на UGC.
	reportSvc := report.NewService(db)
	reportH := report.NewHandler(reportSvc)

	professionSvc := profession.NewService(db, cat)
	professionH := profession.NewHandler(professionSvc)

	aiAdvisorSvc := aiadvisor.NewService(db, cfg.AIAdvisor)
	aiAdvisorH := aiadvisor.NewHandler(aiAdvisorSvc)

	// План 38 Ф.5: payments переехали в billing-service (отдельный микросервис).
	// internal/payment/ удалён, см. docs/plans/38-billing-service.md.

	empireH := empire.NewHandler(pool)
	settingsH := settings.NewHandler(pool).WithAutoMsg(automsgSvc).WithBundle(i18nBundle)
	referralH := referral.NewHandler(pool)
	notepadH := notepad.NewHandler(pool)
	searchH := search.NewHandler(pool)
	techtreeH := techtree.NewHandler(pool, cat)
	battlestatsH := battlestats.NewHandler(pool)
	friendsH := friends.NewHandler(pool)
	recordsH := records.NewHandler(pool, cat)

	adminH := admin.NewHandler(db)
	alienH := alien.NewHandler(db)

	// План 32 Ф.5: chat.Hub использует Redis pub/sub для multi-instance
	// fan-out'а. При rdb=nil деградирует до single-instance broadcast.
	chatHub := chat.NewHubWithRedis(ctx, rdb, log)
	defer chatHub.Close()
	chatH := chat.NewHandler(chatHub, db).WithBlacklist(ugcBlacklist)

	r := httpx.NewRouter(httpx.RouterDeps{Log: log})

	// Health/ready endpoints — без auth, без middleware. Используются
	// orchestrator'ом / nginx upstream health-check для решения, слать
	// ли запросы на этот backend instance. План 31 Ф.1.
	r.Get("/api/health", healthState.HealthHandler())
	r.Get("/api/ready", healthState.ReadyHandler(pool))

	// Feature flags для UI — публично читаются без auth: фронтенд при
	// загрузке решает, какой UI рисовать. План 31 Ф.2.
	r.Get("/api/features", featureH.List)

	// План 36 Ф.12: /api/auth/login|register|refresh удалены из game-nova.
	// Регистрация и логин — только в auth-service. game-nova принимает RSA-JWT.
	// authRL и authH.Register/Login/Refresh пока оставлены в коде как мёртвый
	// код (плановая чистка после удаления legacy HS256 из service.go).
	r.With(authMiddlewareFn).Get("/api/me", authH.Me)
	r.With(authMiddlewareFn).Post("/api/me/vacation", authH.SetVacation)
	r.With(authMiddlewareFn).Delete("/api/me/vacation", authH.UnsetVacation)
	r.Post("/api/battle-sim", battleSimHandler)
	r.Get("/api/stats", scoreH.Stats)
	r.With(authMiddlewareFn).Get("/api/stats/resource-transfers", scoreH.ResourceTransfers)
	// План 38 Ф.5: /api/payment/* удалены. Платежи и пакеты — в billing-service:
	//   GET  /billing/packages   (публичный)
	//   POST /billing/orders     (создаёт payment_order и payURL)
	//   POST /billing/webhooks/{provider}  (HMAC-verified webhook)
	// Frontend дёргает billing напрямую через vite proxy /billing/*.

	// i18n доступна без авторизации (логин-экран тоже использует).
	if i18nH != nil {
		r.Get("/api/i18n", i18nH.Languages)
		r.Get("/api/i18n/{lang}", i18nH.Locale)
	}

	// Prometheus /metrics — внутренний endpoint для мониторинга.
	// Не требует авторизации: закрывается firewall-ом на сетевом уровне.
	r.Handle("/metrics", metrics.Register())

	// Wiki (план 19) — публичная, читает docs/wiki/ru/.
	wikiRoot := os.Getenv("WIKI_ROOT")
	if wikiRoot == "" {
		wikiRoot = "docs/wiki/ru"
	}
	wikiH := wiki.NewHandler(wiki.NewService(wikiRoot))
	r.Get("/api/wiki", wikiH.Index)
	r.Get("/api/wiki/{category}", wikiH.Category)
	r.Get("/api/wiki/{category}/{slug}", wikiH.Page)

	// План 17 F: galaxy events — public read, admin create/cancel.
	r.Get("/api/galaxy-event", galaxyEventH.Active)

	// Universe list (публичный) и handoff-receiver (публичный — вызывается браузером при редиректе).
	r.Get("/api/universes", switcherH.ListUniverses)
	r.Get("/auth/handoff", switcherH.HandoffReceive)
	// Переключение вселенной — требует аутентификации.
	r.With(authMiddlewareFn).Get("/api/universes/switch", switcherH.SwitchUniverse)

	r.Route("/api", func(pr chi.Router) {
		pr.Use(authMiddlewareFn)
		pr.Use(auth.LastSeenMiddleware(pool))
		pr.Get("/empire", empireH.GetAll)
		pr.Get("/settings", settingsH.Get)
		pr.Put("/settings", settingsH.Update)
		// План 36 Critical-6: смена пароля — POST /auth/password в auth-service.
		pr.Post("/me/deletion/code", settingsH.RequestDeletionCode)
		pr.Delete("/me", settingsH.ConfirmDeletion)
		pr.Get("/referrals", referralH.Mine)
		pr.Get("/notepad", notepadH.Get)
		pr.Put("/notepad", notepadH.Save)
		pr.Get("/search", searchH.Search)
		pr.Get("/techtree", techtreeH.Get)
		pr.Get("/battlestats", battlestatsH.List)
		pr.Get("/records", recordsH.List)
		pr.Get("/friends", friendsH.List)
		pr.Post("/friends/{userId}", friendsH.Add)
		pr.Delete("/friends/{userId}", friendsH.Remove)
		pr.Get("/planets", planetH.List)
		pr.Patch("/planets/order", planetH.Reorder)
		pr.Get("/planets/{id}", planetH.Get)
		pr.Patch("/planets/{id}", planetH.Rename)
		pr.Post("/planets/{id}/set-home", planetH.SetHome)
		pr.Delete("/planets/{id}", planetH.Abandon)
		pr.Get("/planets/{id}/resource-report", planetH.ResourceReport)
		pr.Post("/planets/{id}/resource-update", planetH.ResourceUpdate)
		pr.Get("/planets/{id}/forecast", planetH.Forecast)

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

		// План 17 D: daily quests.
		pr.Get("/daily-quests", dailyQuestH.List)
		pr.Post("/daily-quests/{id}/claim", dailyQuestH.Claim)

		pr.Get("/officers", officerH.List)
		pr.Post("/officers/{key}/activate", officerH.Activate)

		pr.Get("/professions", professionH.List)
		pr.Get("/professions/me", professionH.Get)
		pr.Post("/professions/me", professionH.Change)

		pr.Post("/ai-advisor/ask", aiAdvisorH.Ask)
		pr.Get("/ai-advisor/estimate", aiAdvisorH.Estimate)

		// План 38 Ф.5: /payment/order и /payment/history удалены. См. billing-service:
		//   POST /billing/orders, GET /billing/wallet/history.

		pr.Post("/alien/holding/{event_id}/pay", alienH.Pay)
		pr.Get("/alien/holdings/me", alienH.MyHoldings)

		pr.Get("/galaxy/{g}/{s}", galaxyH.System)

		pr.Post("/fleet", fleetH.Send)
		pr.Get("/fleet", fleetH.List)
		pr.Get("/fleet/incoming", fleetH.Incoming)
		pr.Get("/phalanx", fleetH.Phalanx)
		pr.Post("/stargate", fleetH.Stargate)
		pr.Post("/fleet/{id}/recall", fleetH.Recall)

		pr.Get("/market/rates", marketH.Rates)
		pr.Post("/planets/{id}/market/exchange", marketH.Exchange)
		pr.Post("/planets/{id}/market/credit", marketH.ExchangeCredit)
		pr.Get("/market/lots", marketH.ListLots)
		pr.Post("/market/lots", marketH.CreateLot)
		pr.Delete("/market/lots/{id}", marketH.CancelLot)
		pr.Post("/market/lots/{id}/accept", marketH.AcceptLot)
		pr.Get("/market/fleet-lots", marketH.ListFleetLots)
		pr.Post("/planets/{id}/market/fleet-lots", marketH.CreateFleetLot)
		pr.Post("/market/fleet-lots/{lotId}/accept", marketH.AcceptFleetLot)
		pr.Delete("/market/fleet-lots/{lotId}", marketH.CancelFleetLot)

		pr.Post("/planets/{id}/rockets/launch", rocketH.Launch)
		pr.Get("/planets/{id}/rockets", rocketH.Stock)

		pr.Get("/highscore", scoreH.Highscore)
		pr.Get("/highscore/me", scoreH.MyRank)
		pr.Get("/highscore/alliances", scoreH.Alliances)
		pr.Get("/highscore/vacation", scoreH.Vacation)

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
		pr.Get("/messages/sent", messageH.Sent)
		pr.Post("/messages", messageH.Compose)
		pr.Delete("/messages", messageH.DeleteAll)
		pr.Delete("/messages/{id}", messageH.Delete)
		pr.Get("/messages/unread-count", messageH.UnreadCount)
		pr.Post("/messages/{id}/read", messageH.MarkRead)
		pr.Get("/battle-reports/{id}", messageH.GetReport)
		pr.Get("/espionage-reports/{id}", messageH.GetEspionageReport)
		pr.Get("/expedition-reports/{id}", messageH.GetExpeditionReport)

		// План 46 Ф.3: жалобы пользователей.
		pr.Post("/reports", reportH.Create)

		pr.Route("/admin", func(ar chi.Router) {
			// Ф.8.1 RBAC: на уровне префикса — минимум support (модератор),
			// на destructive операции — admin или superadmin.
			ar.Use(admin.RequireRole(db, admin.RoleSupport))
			// AuditMiddleware: для write-запросов (не-GET) после 2xx-ответа
			// асинхронно пишет запись в admin_audit_log. См. Ф.1.2 план 14.
			ar.Use(admin.AuditMiddleware(db))
			// RateLimitMiddleware: защита от human-error (например,
			// случайный banAll). 100 write-действий/час на админа,
			// in-memory. См. Ф.8.2 план 14.
			ar.Use(admin.RateLimitMiddleware())

			// Read-only + лёгкая модерация — доступно support+.
			ar.Get("/stats", adminH.Stats)
			ar.Get("/users", adminH.ListUsers)
			ar.Get("/users/{id}", adminH.GetUserProfile)
			ar.Post("/users/{id}/ban", adminH.Ban)
			ar.Post("/users/{id}/unban", adminH.Unban)
			ar.Get("/events", adminH.EventsList)
			ar.Get("/events/stats", adminH.EventsStats)
			ar.Get("/events/dead", adminH.ListDeadEvents)
			ar.Get("/audit", adminH.ListAudit)

			// План 46 Ф.3: модерация UGC-жалоб (support+).
			ar.Get("/reports", reportH.AdminList)
			ar.Post("/reports/{id}/resolve", reportH.AdminResolve)

			// Destructive — admin+.
			ar.With(admin.RequireRole(db, admin.RoleAdmin)).Post("/users/{id}/credit", adminH.Credit)
			ar.With(admin.RequireRole(db, admin.RoleAdmin)).Post("/users/{id}/resources", adminH.GrantResources)
			ar.With(admin.RequireRole(db, admin.RoleAdmin)).Post("/users/{id}/artefacts/grant", adminH.GrantArtefact)
			ar.With(admin.RequireRole(db, admin.RoleAdmin)).Delete("/users/{id}/artefacts/{aid}", adminH.DeleteArtefact)
			ar.With(admin.RequireRole(db, admin.RoleAdmin)).Post("/events/{id}/retry", adminH.EventRetry)
			ar.With(admin.RequireRole(db, admin.RoleAdmin)).Post("/events/{id}/cancel", adminH.EventCancel)
			ar.With(admin.RequireRole(db, admin.RoleAdmin)).Post("/events/dead/{id}/resurrect", adminH.ResurrectDeadEvent)

			// План 14 Ф.2.4-2.6 — force-recall, planet-management, user-delete.
			fleetAdminH := admin.NewFleetAdminHandler(transportSvc, db)
			ar.With(admin.RequireRole(db, admin.RoleAdmin)).Post("/fleets/{fleet_id}/recall", fleetAdminH.ForceRecall)
			ar.With(admin.RequireRole(db, admin.RoleAdmin)).Post("/planets/{id}/rename", adminH.PlanetRename)
			ar.With(admin.RequireRole(db, admin.RoleAdmin)).Post("/planets/{id}/transfer", adminH.PlanetTransfer)
			ar.With(admin.RequireRole(db, admin.RoleAdmin)).Delete("/planets/{id}", adminH.PlanetDelete)
			ar.With(admin.RequireRole(db, admin.RoleAdmin)).Delete("/users/{id}", adminH.UserSoftDelete)
			ar.With(admin.RequireRole(db, admin.RoleAdmin)).Post("/users/{id}/restore", adminH.UserRestore)

			// План 17 F: галактические события (admin создаёт/отменяет).
			ar.With(admin.RequireRole(db, admin.RoleAdmin)).Post("/galaxy-events", galaxyEventH.Create)
			ar.With(admin.RequireRole(db, admin.RoleAdmin)).Delete("/galaxy-events/{id}", galaxyEventH.Cancel)

			// Только superadmin может менять роли (privilege escalation).
			ar.With(admin.RequireRole(db, admin.RoleSuperadmin)).Post("/users/{id}/role", adminH.SetRole)
		})
	})

	srv := &http.Server{
		Addr:              cfg.Server.Addr,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// Все зависимости подняты — сервер готов принимать запросы.
	healthState.SetReady()

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

	// Phase 1 (drain): /api/health начинает отвечать 503, чтобы nginx/
	// балансировщик убрал backend из upstream до фактического shutdown.
	// Это устраняет 502 во время выкатки. План 31 Ф.1.
	healthState.SetDraining()
	log.InfoContext(ctx, "draining", slog.Duration("delay", drainDelay))
	select {
	case <-time.After(drainDelay):
	case <-context.Background().Done():
		// Никогда не сработает — фоновый context не отменяется. Здесь
		// чисто на случай будущих доработок (force-shutdown signal).
	}

	// Phase 2 (shutdown): graceful shutdown с timeout. Активные
	// запросы завершаются, новые отклоняются, listener закрывается.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	log.InfoContext(ctx, "shutting down http server")
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
