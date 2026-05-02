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
	"oxsar/game-nova/internal/exchange"
	"oxsar/game-nova/internal/auth"
	"oxsar/game-nova/internal/automsg"
	"oxsar/game-nova/internal/balance"
	"oxsar/game-nova/internal/battle"
	billingclient "oxsar/game-nova/internal/billing/client"
	"oxsar/game-nova/internal/battlestats"
	"oxsar/game-nova/internal/building"
	"oxsar/game-nova/internal/catalog"
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
	"oxsar/game-nova/internal/monitor"
	"oxsar/game-nova/internal/notepad"
	"oxsar/game-nova/internal/officer"
	originalien "oxsar/game-nova/internal/origin/alien"
	"oxsar/game-nova/internal/planet"
	"oxsar/game-nova/internal/profession"
	"oxsar/game-nova/internal/records"
	"oxsar/game-nova/internal/referral"
	"oxsar/game-nova/internal/repair"
	"oxsar/game-nova/internal/repo"
	"oxsar/game-nova/internal/requirements"
	"oxsar/game-nova/internal/research"
	"oxsar/game-nova/internal/rocket"
	"oxsar/game-nova/internal/score"
	"oxsar/game-nova/internal/search"
	"oxsar/game-nova/internal/settings"
	"oxsar/game-nova/internal/battlereport"
	"oxsar/game-nova/internal/shipyard"
	"oxsar/game-nova/internal/simulator"
	"oxsar/game-nova/internal/techtree"
	"oxsar/game-nova/internal/storage"
	"oxsar/game-nova/internal/dailyquest"
	"oxsar/game-nova/internal/features"
	"oxsar/game-nova/internal/galaxyevent"
	"oxsar/game-nova/internal/health"
	"oxsar/game-nova/internal/universe"
	"oxsar/game-nova/internal/universeswitcher"
	"oxsar/game-nova/internal/wiki"
	"oxsar/game-nova/pkg/idempotency"
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

	univReg, err := universe.NewRegistry(filepath.Join(catalogDir, "universes.yaml"))
	if err != nil {
		return fmt.Errorf("universes.yaml: %w", err)
	}

	// План 72.1 часть 12: per-universe параметры (Speed, NumGalaxies, …,
	// Teleport*) живут в configs/universes.yaml. Подтягиваем их в cfg.Game
	// сразу после загрузки реестра — fail-fast если UNIVERSE_ID не найден.
	curUni, ok := univReg.ByID(cfg.Auth.UniverseID)
	if !ok {
		return fmt.Errorf("universes.yaml: UNIVERSE_ID=%q not found", cfg.Auth.UniverseID)
	}
	cfg.ApplyUniverse(config.UniverseParams{
		Name:                   curUni.Name,
		Speed:                  curUni.Speed,
		Deathmatch:             curUni.Deathmatch,
		NumGalaxies:            curUni.NumGalaxies,
		NumSystems:             curUni.NumSystems,
		MaxPlanets:             curUni.MaxPlanets,
		BashingPeriod:          curUni.BashingPeriod,
		BashingMaxAttacks:      curUni.BashingMaxAttacks,
		ProtectionPeriod:       curUni.ProtectionPeriod,
		StorageFactor:          curUni.StorageFactor,
		ResearchSpeedFactor:    curUni.ResearchSpeedFactor,
		EnergyProductionFactor: curUni.EnergyProductionFactor,
		TeleportCostOxsars:     curUni.TeleportCostOxsars,
		TeleportCooldownHours:  curUni.TeleportCooldownHours,
		TeleportDurationMin:    curUni.TeleportDurationMin,
	})

	// Per-universe balance (план 64). Для modern-вселенных (uni01, uni02
	// и любых, у которых нет configs/balance/<id>.yaml) bundle == чистый
	// дефолт; для origin-вселенной — applies override-файл.
	balanceLoader := balance.NewLoader(catalogDir)
	balanceBundle, err := balanceLoader.LoadForCtx(ctx, log, cfg.Auth.UniverseID)
	if err != nil {
		return err
	}
	cat := balanceBundle.Catalog

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
	starter := planet.NewStarter(db, cfg.Game.NumGalaxies, cfg.Game.NumSystems)

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

	// authMiddleware: RSA-валидация JWT от identity-service + lazy-create юзера
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

	simulatorH := simulator.NewHandler(db, cat)
	battleReportH := battlereport.NewHandler(db)

	repairSvc := repair.NewService(db, planetSvc, cat, reqs, cfg.Game.Speed).
		WithAutoMsg(automsgSvc).WithBundle(i18nBundle)
	repairH := repair.NewHandler(repairSvc)

	artefactSvc := artefact.NewService(db, cat).WithAutoMsg(automsgSvc).WithBundle(i18nBundle)
	artefactH := artefact.NewHandler(artefactSvc)

	galaxyH := galaxy.NewHandler(galaxy.NewRepository(pool), cfg.Game.NumGalaxies, cfg.Game.NumSystems)

	dailyQuestSvc := dailyquest.New(pool)
	dailyQuestH := dailyquest.NewHandler(dailyQuestSvc)

	transportSvc := fleet.NewTransportServiceWithConfig(db, cat, cfg.Game.Speed, artefactSvc, cfg.Game.NumGalaxies, cfg.Game.NumSystems, cfg.Game.MaxPlanets, cfg.Game.ProtectionPeriod).WithBundle(i18nBundle)
	transportSvc.SetBashingLimits(cfg.Game.BashingPeriod, cfg.Game.BashingMaxAttacks)
	transportSvc.SetDailyQuestSvc(dailyQuestSvc)
	fleetH := fleet.NewHandler(transportSvc, rdb)

	// План 72.1.20: /api/monitor-planet — мониторинг чужой планеты
	// через здание STAR_SURVEILLANCE (legacy MonitorPlanet.class.php).
	monitorH := monitor.NewHandler(db, transportSvc).
		WithAutoMsg(automsgSvc).WithBundle(i18nBundle)

	messageSvc := message.NewService(db)
	messageH := message.NewHandler(messageSvc)

	marketSvc := market.NewService(db).WithAutoMsg(automsgSvc).WithBundle(i18nBundle)
	marketH := market.NewHandler(marketSvc, rdb)

	rocketSvc := rocket.NewService(db, cat, cfg.Game.Speed, cfg.Game.NumGalaxies, cfg.Game.NumSystems).WithBundle(i18nBundle)
	rocketH := rocket.NewHandler(rocketSvc)

	// План 72.1.42: artmarket теперь использует automsg.Send (i18n
	// шаблон), bundle нужен на стороне automsg (он уже подключён выше).
	artMarketSvc := artmarket.NewService(db).WithAutoMsg(automsgSvc)
	artMarketH := artmarket.NewHandler(artMarketSvc, rdb)

	// План 68: биржа артефактов (P2P пакетный обмен на оксариты).
	exchangeRepo := exchange.NewPgRepo(db)
	exchangeSvc := exchange.NewService(db, exchangeRepo, exchange.DefaultConfig()).
		WithAutoMsg(automsgSvc).WithBundle(i18nBundle)
	exchangeH := exchange.NewHandler(exchangeSvc)

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

	allianceH := alliance.NewHandler(allianceSvc).WithRedis(rdb)

	// План 46 Ф.3 (149-ФЗ) → план 56: жалобы перенесены в portal-backend.
	// game-nova больше не владеет user_reports, см. portal/internal/report.

	professionSvc := profession.NewService(db, cat).WithAutoMsg(automsgSvc).WithBundle(i18nBundle)
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
	catalogH := catalog.NewHandler(cat)

	adminH := admin.NewHandler(db)
	alienH := alien.NewHandler(db)

	// План 77: billing-client для списания/возврата оксаров. URL пустой →
	// клиент возвращает ErrNotConfigured (премиум-фичи отключены, остальные
	// endpoint'ы работают). В production BILLING_URL обязателен.
	billingC := billingclient.New(cfg.Billing.URL)
	if cfg.Billing.URL == "" {
		log.WarnContext(ctx, "BILLING_URL not set; premium features (oxsar spend/refund) disabled")
	} else {
		log.InfoContext(ctx, "billing client configured", slog.String("url", cfg.Billing.URL))
	}

	// План 77 Ф.2: idempotency-middleware. Производственное TTL 24h.
	idemMW := idempotency.NewMiddleware(rdb, 0)

	// План 66 Ф.5: alien-buyout HTTP-handler (платный выкуп HOLDING
	// оксарами). Использует default-конфиг origin/alien (BuyoutBaseOxsars
	// = 100); per-universe override в Ф.5 не реализуется — конфиг
	// внутрипакетный, как у HoldingAIHandler в Ф.4.
	alienBuyoutH := originalien.NewBuyoutHandler(db, billingC, originalien.DefaultConfig())

	// План 65 Ф.6: телепорт планеты (премиум, оплата оксарами).
	// Idempotency-Key обязателен (R9). Параметры из cfg.Game.Teleport*;
	// per-universe override Ф.6 не вводит — modern-default == origin-default
	// (24h cooldown зеркалит legacy PLANET_TELEPORT_MIN_INTERVAL_TIME).
	planetTeleportH := planet.NewTeleportHandler(db, billingC, planet.TeleportConfig{
		CostOxsars:      cfg.Game.TeleportCostOxsars,
		CooldownHours:   cfg.Game.TeleportCooldownHours,
		DurationMinutes: cfg.Game.TeleportDurationMinutes,
		NumGalaxies:     cfg.Game.NumGalaxies,
		NumSystems:      cfg.Game.NumSystems,
	})

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
	// Регистрация и логин — только в identity-service. game-nova принимает RSA-JWT.
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

	// План 72.1 ч.20.11: публичный анонимный endpoint для просмотра
	// боевого отчёта по uuid (для отправки ссылок куда угодно).
	r.Get("/api/battle-reports/{id}", battleReportH.GetByID)

	r.Route("/api", func(pr chi.Router) {
		pr.Use(authMiddlewareFn)
		pr.Use(auth.LastSeenMiddleware(pool))
		pr.Get("/empire", empireH.GetAll)
		pr.Get("/settings", settingsH.Get)
		// План 72.1.50 ч.4B (rollout 77 Ф.3.2): social+fleet+planet
		// мутирующие endpoint'ы. Фронт уже шлёт Idempotency-Key для
		// каждого. Alliance — пропущено (handler-level idempotency
		// через FromRequest уже есть, план 67); fleet.Send тоже
		// (handler.go:56 FromRequest).
		pr.With(idemMW.Wrap).Put("/settings", settingsH.Update)
		// План 36 Critical-6: смена пароля — POST /auth/password в identity-service.
		pr.With(idemMW.Wrap).Post("/me/deletion/code", settingsH.RequestDeletionCode)
		pr.With(idemMW.Wrap).Delete("/me", settingsH.ConfirmDeletion)
		// План 72.1.30: cancel pending удалением в grace-period.
		pr.With(idemMW.Wrap).Post("/me/deletion/cancel", settingsH.CancelDeletion)
		pr.Get("/referrals", referralH.Mine)
		pr.Get("/notepad", notepadH.Get)
		pr.With(idemMW.Wrap).Put("/notepad", notepadH.Save)
		pr.Get("/search", searchH.Search)
		pr.Get("/techtree", techtreeH.Get)
		pr.Get("/battlestats", battlestatsH.List)
		pr.Get("/records", recordsH.List)
		// План 72 Ф.4 Spring 3 — catalog endpoints для origin info-страниц.
		pr.Get("/buildings/catalog/{type}", catalogH.BuildingByType)
		pr.Get("/units/catalog/{type}", catalogH.UnitByType)
		pr.Get("/artefacts/catalog/{type}", catalogH.ArtefactByType)
		pr.Get("/friends", friendsH.List)
		pr.With(idemMW.Wrap).Post("/friends/{userId}", friendsH.Add)
		pr.With(idemMW.Wrap).Post("/friends/{userId}/accept", friendsH.Accept)
		pr.With(idemMW.Wrap).Delete("/friends/{userId}", friendsH.Remove)
		pr.Get("/planets", planetH.List)
		pr.With(idemMW.Wrap).Patch("/planets/order", planetH.Reorder)
		pr.Get("/planets/{id}", planetH.Get)
		pr.With(idemMW.Wrap).Patch("/planets/{id}", planetH.Rename)
		pr.With(idemMW.Wrap).Post("/planets/{id}/set-home", planetH.SetHome)
		pr.With(idemMW.Wrap).Delete("/planets/{id}", planetH.Abandon)
		pr.Get("/planets/{id}/resource-report", planetH.ResourceReport)
		pr.With(idemMW.Wrap).Post("/planets/{id}/resource-update", planetH.ResourceUpdate)
		pr.Get("/planets/{id}/forecast", planetH.Forecast)

		pr.Get("/planets/{id}/buildings", buildingH.Levels)
		// План 72.1.50 ч.4 (rollout 77 Ф.3.2): все мутирующие
		// builds/research/shipyard/repair endpoint'ы оборачиваются в
		// idemMW.Wrap — фронт уже шлёт Idempotency-Key для каждого.
		pr.With(idemMW.Wrap).Post("/planets/{id}/buildings", buildingH.Enqueue)
		pr.Get("/planets/{id}/buildings/queue", buildingH.List)
		pr.With(idemMW.Wrap).Delete("/planets/{id}/buildings/queue/{taskId}", buildingH.Cancel)
		// План 72.1.44 cross-cut: VIP-instant старт стройки.
		pr.With(idemMW.Wrap).Post("/planets/{id}/buildings/queue/{taskId}/vip", buildingH.StartVIP)
		// План 72.1.33: legacy BuildingInfo::DEMOLISH_NOW.
		pr.With(idemMW.Wrap).Post("/planets/{id}/buildings/{unitId}/demolish", buildingH.Demolish)
		// План 72.1.33 ч.2: legacy BuildingInfo::PackConstruction +
		// PackResearch — упаковка здания/исследования в packed-артефакт.
		pr.With(idemMW.Wrap).Post("/planets/{id}/buildings/{unitId}/pack", artefactH.PackBuilding)
		pr.With(idemMW.Wrap).Post("/planets/{id}/research/{unitId}/pack", artefactH.PackResearch)

		pr.With(idemMW.Wrap).Post("/planets/{id}/research", researchH.Enqueue)
		pr.Get("/research", researchH.List)
		// План 72.1.39: legacy `Research::abort` — отмена с refund.
		pr.With(idemMW.Wrap).Delete("/research/{queueId}", researchH.Cancel)
		// План 72.1.44 cross-cut: VIP-instant старт исследования.
		pr.With(idemMW.Wrap).Post("/research/{queueId}/vip", researchH.StartVIP)

		// План 72.1.49: anti-double-submit через Idempotency-Key
		// (legacy `Shipyard.class.php:382` использовал 5-сек acquireLock).
		// Frontend `shipyard.ts:35` уже шлёт ключ — middleware теперь
		// обеспечивает дедуп на сервере.
		pr.With(idemMW.Wrap).Post("/planets/{id}/shipyard", shipyardH.Enqueue)
		pr.Get("/planets/{id}/shipyard/queue", shipyardH.List)
		pr.Get("/planets/{id}/shipyard/inventory", shipyardH.Inventory)
		// План 72.1.41: legacy `Shipyard` capacity-info (freeShield/Rocket).
		pr.Get("/planets/{id}/shipyard/capacity", shipyardH.Capacity)
		pr.With(idemMW.Wrap).Delete("/planets/{id}/shipyard/{queueId}", shipyardH.Cancel)
		// План 72.1.44 cross-cut: VIP-instant старт shipyard-задачи.
		pr.With(idemMW.Wrap).Post("/planets/{id}/shipyard/{queueId}/vip", shipyardH.StartVIP)

		pr.Post("/simulator/run", simulatorH.Run)

		pr.Get("/users/me/battles", battleReportH.ListMine)
		// /battle-reports/{id} зарегистрирован публично ниже —
		// анонимный просмотр по ссылке (план 72.1 ч.20.11).

		// План 72.1.49: Idempotency-middleware для repair/disassemble и
		// repair/repair (тот же anti-double-submit паттерн что и shipyard).
		pr.With(idemMW.Wrap).Post("/planets/{id}/repair/disassemble", repairH.EnqueueDisassemble)
		pr.With(idemMW.Wrap).Post("/planets/{id}/repair/repair", repairH.EnqueueRepair)
		pr.Get("/planets/{id}/repair/damaged", repairH.ListDamaged)
		pr.Get("/planets/{id}/repair/queue", repairH.List)
		pr.With(idemMW.Wrap).Delete("/planets/{id}/repair/queue/{queueId}", repairH.Cancel)
		// План 72.1.25: VIP-старт за credit (legacy startEventVIP).
		pr.With(idemMW.Wrap).Post("/planets/{id}/repair/queue/{queueId}/vip", repairH.StartVIP)

		pr.Get("/artefacts", artefactH.List)
		pr.With(idemMW.Wrap).Post("/artefacts/{id}/activate", artefactH.Activate)
		pr.With(idemMW.Wrap).Post("/artefacts/{id}/deactivate", artefactH.Deactivate)
		pr.With(idemMW.Wrap).Post("/artefacts/{id}/sell", artMarketH.ListForSale)
		// План 72.1.45 §2: история приобретений артефакта (legacy ArtefactInfo).
		pr.Get("/artefacts/info/{unitId}/history", artefactH.History)

		pr.Get("/artefact-market/offers", artMarketH.Offers)
		pr.Get("/artefact-market/credit", artMarketH.Credit)
		pr.With(idemMW.Wrap).Post("/artefact-market/offers/{id}/buy", artMarketH.Buy)
		pr.With(idemMW.Wrap).Delete("/artefact-market/offers/{id}", artMarketH.Cancel)

		// План 68: биржа артефактов (player-to-player пакетный обмен).
		exchangeH.Routes(pr)

		pr.Get("/achievements", achH.List)

		// План 17 D: daily quests.
		pr.Get("/daily-quests", dailyQuestH.List)
		pr.Post("/daily-quests/{id}/claim", dailyQuestH.Claim)

		pr.Get("/officers", officerH.List)
		pr.With(idemMW.Wrap).Post("/officers/{key}/activate", officerH.Activate)

		pr.Get("/professions", professionH.List)
		pr.Get("/professions/me", professionH.Get)
		pr.With(idemMW.Wrap).Post("/professions/me", professionH.Change)

		pr.Post("/ai-advisor/ask", aiAdvisorH.Ask)
		pr.Get("/ai-advisor/estimate", aiAdvisorH.Estimate)

		// План 38 Ф.5: /payment/order и /payment/history удалены. См. billing-service:
		//   POST /billing/orders, GET /billing/wallet/history.

		pr.With(idemMW.Wrap).Post("/alien/holding/{event_id}/pay", alienH.Pay)
		pr.Get("/alien/holdings/me", alienH.MyHoldings)

		// План 66 Ф.5: платный выкуп HOLDING-удержания за оксары.
		// Idempotency-Key обязателен (R9), middleware дедуплицирует
		// повторы по ключу + body-hash (план 77 Ф.2).
		pr.With(idemMW.Wrap).Post("/alien-missions/{mission_id}/buyout", alienBuyoutH.Buyout)

		// План 65 Ф.6: телепорт планеты на новые координаты (премиум,
		// оксары). Idempotency-Key обязателен; общий dedup-namespace по
		// ключу + body-hash, как у alien-buyout.
		pr.With(idemMW.Wrap).Post("/planets/{id}/teleport", planetTeleportH.Teleport)

		pr.Get("/galaxy/{g}/{s}", galaxyH.System)

		// План 72.1.50 ч.4B: fleet (без Send — handler-level FromRequest
		// уже есть в fleet/handler.go:56), market, rocket.
		pr.Post("/fleet", fleetH.Send)
		pr.Get("/fleet", fleetH.List)
		pr.Get("/fleet/incoming", fleetH.Incoming)
		pr.Get("/phalanx", fleetH.Phalanx)
		// План 72.1.20: monitor-planet (legacy `?go=MonitorPlanet&id=`).
		pr.Get("/monitor-planet", monitorH.Monitor)
		pr.With(idemMW.Wrap).Post("/stargate", fleetH.Stargate)
		pr.With(idemMW.Wrap).Post("/fleet/{id}/recall", fleetH.Recall)
		// План 72.1.47: load/unload для HOLDING-флотов (legacy
		// `Mission.class.php::loadResourcesToFleet/unloadResourcesFromFleet`).
		pr.With(idemMW.Wrap).Post("/fleet/{id}/load", fleetH.Load)
		pr.With(idemMW.Wrap).Post("/fleet/{id}/unload", fleetH.Unload)
		// План 72.1.48: formation (legacy `Mission.class.php::formation`).
		pr.With(idemMW.Wrap).Post("/fleet/{id}/promote-to-acs", fleetH.PromoteToACS)
		pr.With(idemMW.Wrap).Post("/acs/{groupId}/invite", fleetH.InviteACS)
		pr.Get("/acs/invitations", fleetH.ListACSInvitations)
		pr.With(idemMW.Wrap).Post("/acs/invitations/{groupId}/accept", fleetH.AcceptACSInvitation)

		pr.Get("/market/rates", marketH.Rates)
		pr.With(idemMW.Wrap).Post("/planets/{id}/market/exchange", marketH.Exchange)
		pr.With(idemMW.Wrap).Post("/planets/{id}/market/credit", marketH.ExchangeCredit)
		// План 72.1.28: multi-resource Credit_ex (legacy `Market::Credit_ex`).
		pr.With(idemMW.Wrap).Post("/planets/{id}/market/credit-multi", marketH.ExchangeCreditMulti)
		pr.Get("/market/lots", marketH.ListLots)
		pr.With(idemMW.Wrap).Post("/market/lots", marketH.CreateLot)
		pr.With(idemMW.Wrap).Delete("/market/lots/{id}", marketH.CancelLot)
		pr.With(idemMW.Wrap).Post("/market/lots/{id}/accept", marketH.AcceptLot)
		pr.Get("/market/fleet-lots", marketH.ListFleetLots)
		pr.With(idemMW.Wrap).Post("/planets/{id}/market/fleet-lots", marketH.CreateFleetLot)
		pr.With(idemMW.Wrap).Post("/market/fleet-lots/{lotId}/accept", marketH.AcceptFleetLot)
		pr.With(idemMW.Wrap).Delete("/market/fleet-lots/{lotId}", marketH.CancelFleetLot)

		pr.With(idemMW.Wrap).Post("/planets/{id}/rockets/launch", rocketH.Launch)
		pr.Get("/planets/{id}/rockets", rocketH.Stock)

		pr.Get("/highscore", scoreH.Highscore)
		pr.Get("/highscore/me", scoreH.MyRank)
		pr.Get("/highscore/alliances", scoreH.Alliances)
		pr.Get("/highscore/vacation", scoreH.Vacation)

		pr.Get("/alliances", allianceH.List)
		pr.Get("/alliances/me", allianceH.My)
		pr.Get("/alliances/{id}", allianceH.Get)
		pr.Get("/alliances/{id}/applications", allianceH.Applications)
		// План 72.1.43: legacy globalMail + updateAllyTag/Name.
		pr.Post("/alliances/{id}/broadcast", allianceH.BroadcastMail)
		pr.Patch("/alliances/{id}", allianceH.UpdateTagName)
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
		// План 67 Ф.2: расширения (3 описания, ranks CRUD, kick, audit).
		pr.Get("/alliances/{id}/descriptions", allianceH.GetDescriptions)
		pr.Patch("/alliances/{id}/descriptions", allianceH.UpdateDescriptions)
		pr.Get("/alliances/{id}/ranks", allianceH.ListRanks)
		pr.Post("/alliances/{id}/ranks", allianceH.CreateRank)
		pr.Patch("/alliances/{id}/ranks/{rank_id}", allianceH.UpdateRank)
		pr.Delete("/alliances/{id}/ranks/{rank_id}", allianceH.DeleteRank)
		pr.Patch("/alliances/{id}/members/{userID}/rank-id", allianceH.AssignMemberRank)
		pr.Delete("/alliances/{id}/members/{userID}", allianceH.Kick)
		pr.Get("/alliances/{id}/audit", allianceH.ListAudit)
		pr.Post("/alliances/{id}/transfer-leadership/code", allianceH.RequestTransferLeadership)
		pr.Post("/alliances/{id}/transfer-leadership", allianceH.ConfirmTransferLeadership)

		pr.Get("/chat/{kind}/history", chatH.History)
		pr.With(idemMW.Wrap).Post("/chat/{kind}/send", chatH.Send)
		pr.Get("/chat/{kind}/ws", chatH.Connect)
		pr.Get("/chat/{kind}/unread", chatH.UnreadCount)
		pr.With(idemMW.Wrap).Post("/chat/{kind}/read", chatH.MarkRead)
		pr.With(idemMW.Wrap).Patch("/chat/messages/{id}", chatH.EditMessage)
		pr.With(idemMW.Wrap).Delete("/chat/messages/{id}", chatH.DeleteMessage)

		pr.Get("/messages", messageH.Inbox)
		pr.Get("/messages/folders", messageH.Folders)
		pr.Get("/messages/sent", messageH.Sent)
		pr.With(idemMW.Wrap).Post("/messages", messageH.Compose)
		pr.With(idemMW.Wrap).Delete("/messages", messageH.DeleteAll)
		pr.With(idemMW.Wrap).Delete("/messages/{id}", messageH.Delete)
		pr.Get("/messages/unread-count", messageH.UnreadCount)
		pr.With(idemMW.Wrap).Post("/messages/{id}/read", messageH.MarkRead)
		// /battle-reports/{id} перенесён в публичный router (план 72.1 ч.20.11).
		pr.Get("/espionage-reports/{id}", messageH.GetEspionageReport)
		pr.Get("/expedition-reports/{id}", messageH.GetExpeditionReport)

		// План 56: POST /reports перенесён в portal-backend
		// (POST https://oxsar-nova.ru/api/reports). См. план 56.

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

			// План 56: модерация UGC-жалоб перенесена в admin-frontend
			// через admin-bff → portal-backend (/api/admin/reports*).

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
// Если NumSim ≥ 2, возвращает SimStats вместо Report.
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
	stats, _, err := battle.MultiRun(in, n)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, stats)
}

func decodeJSON(r *http.Request, into any) error {
	dec := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	dec.DisallowUnknownFields()
	return dec.Decode(into)
}
