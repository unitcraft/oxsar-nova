import { Suspense, lazy, useState, useEffect, useRef } from 'react';
import { useQuery } from '@tanstack/react-query';
import { useAuthStore } from './stores/auth';
import { api } from './api/client';
import type { Planet } from './api/types';
import { LoginScreen } from './features/auth/LoginScreen';
import { OverviewScreen } from './features/overview/OverviewScreen';
import { useTranslation } from './i18n/i18n';
import { ToastProvider } from './ui/Toast';
import { ResourceTicker } from './ui/ResourceTicker';
import { Countdown } from './ui/Countdown';
import { ScreenSkeleton } from './ui/Skeleton';
import { useKeyboardShortcuts } from './lib/useKeyboardShortcuts';

const BuildingsScreen    = lazy(() => import('./features/buildings/BuildingsScreen').then(m => ({ default: m.BuildingsScreen })));
const ResearchScreen     = lazy(() => import('./features/research/ResearchScreen').then(m => ({ default: m.ResearchScreen })));
const ShipyardScreen     = lazy(() => import('./features/shipyard/ShipyardScreen').then(m => ({ default: m.ShipyardScreen })));
const ArtefactsScreen    = lazy(() => import('./features/artefacts/ArtefactsScreen').then(m => ({ default: m.ArtefactsScreen })));
const BattleSimScreen    = lazy(() => import('./features/battle-sim/BattleSimScreen').then(m => ({ default: m.BattleSimScreen })));
const GalaxyScreen       = lazy(() => import('./features/galaxy/GalaxyScreen').then(m => ({ default: m.GalaxyScreen })));
const FleetScreen        = lazy(() => import('./features/fleet/FleetScreen').then(m => ({ default: m.FleetScreen })));
const RepairScreen       = lazy(() => import('./features/repair/RepairScreen').then(m => ({ default: m.RepairScreen })));
const MessagesScreen     = lazy(() => import('./features/messages/MessagesScreen').then(m => ({ default: m.MessagesScreen })));
const MarketScreen       = lazy(() => import('./features/market/MarketScreen').then(m => ({ default: m.MarketScreen })));
const RocketsScreen      = lazy(() => import('./features/rockets/RocketsScreen').then(m => ({ default: m.RocketsScreen })));
const ArtefactMarketScreen = lazy(() => import('./features/artmarket/ArtefactMarketScreen').then(m => ({ default: m.ArtefactMarketScreen })));
const OfficersScreen     = lazy(() => import('./features/officers/OfficersScreen').then(m => ({ default: m.OfficersScreen })));
const ScoreScreen        = lazy(() => import('./features/score/ScoreScreen').then(m => ({ default: m.ScoreScreen })));
const AllianceScreen     = lazy(() => import('./features/alliance/AllianceScreen').then(m => ({ default: m.AllianceScreen })));
const AchievementsScreen = lazy(() => import('./features/achievements/AchievementsScreen').then(m => ({ default: m.AchievementsScreen })));
const ChatScreen         = lazy(() => import('./features/chat/ChatScreen').then(m => ({ default: m.ChatScreen })));
const PlanetOptionsScreen = lazy(() => import('./features/planet-options/PlanetOptionsScreen').then(m => ({ default: m.PlanetOptionsScreen })));
const ResourceScreen     = lazy(() => import('./features/resource/ResourceScreen').then(m => ({ default: m.ResourceScreen })));
const AdminScreen        = lazy(() => import('./features/admin/AdminScreen').then(m => ({ default: m.AdminScreen })));
const UnitInfoScreen     = lazy(() => import('./features/unit-info/UnitInfoScreen').then(m => ({ default: m.UnitInfoScreen })));

type Tab =
  | 'overview' | 'buildings' | 'research' | 'shipyard' | 'repair'
  | 'artefacts' | 'galaxy' | 'fleet' | 'market' | 'rockets'
  | 'art-market' | 'officers' | 'achievements' | 'score'
  | 'messages' | 'alliance' | 'chat' | 'sim' | 'admin' | 'planet-options' | 'resource'
  | 'unit-info';

const VALID_TABS = new Set<string>([
  'overview', 'buildings', 'research', 'shipyard', 'repair',
  'artefacts', 'galaxy', 'fleet', 'market', 'rockets',
  'art-market', 'officers', 'achievements', 'score',
  'messages', 'alliance', 'chat', 'sim', 'admin', 'planet-options', 'resource',
  'unit-info',
]);

type InfoUnit = { kind: 'building' | 'research' | 'ship' | 'defense'; id: number; level: number; fromTab: Tab };

const INFO_KINDS = new Set(['building', 'research', 'ship', 'defense']);

function parseHash(): { tab: Tab; infoUnit: InfoUnit | null } {
  const hash = window.location.hash.replace('#', '');
  const parts = hash.split('/');
  if (parts[0] === 'unit-info' && parts[1] && INFO_KINDS.has(parts[1]) && parts[2]) {
    const id = parseInt(parts[2], 10);
    if (!isNaN(id)) {
      return { tab: 'unit-info', infoUnit: { kind: parts[1] as InfoUnit['kind'], id, level: 0, fromTab: 'overview' } };
    }
  }
  const tab = VALID_TABS.has(parts[0] ?? '') ? (parts[0] as Tab) : 'overview';
  return { tab, infoUnit: null };
}

export function App() {
  const token = useAuthStore((s) => s.accessToken);
  return (
    <ToastProvider>
      {token ? <AuthenticatedApp /> : <LoginScreen />}
    </ToastProvider>
  );
}

function AuthenticatedApp() {
  const [tab, setTab] = useState<Tab>(() => parseHash().tab);
  const { t } = useTranslation();
  const logout = useAuthStore((s) => s.logout);

  const navigateTo = (next: Tab) => {
    setTab(next);
    history.pushState(null, '', `#${next}`);
  };

  useEffect(() => {
    const onPop = () => {
      const parsed = parseHash();
      setTab(parsed.tab);
      if (parsed.infoUnit) setInfoUnit(parsed.infoUnit);
    };
    window.addEventListener('popstate', onPop);
    return () => window.removeEventListener('popstate', onPop);
  }, []);

  useKeyboardShortcuts([
    {
      key: 'h',
      alt: true,
      handler: () => navigateTo('overview'),
      description: 'Alt+H — главный экран',
    },
    {
      key: 'b',
      alt: true,
      handler: () => navigateTo('buildings'),
      description: 'Alt+B — постройки',
    },
    {
      key: 'r',
      alt: true,
      handler: () => navigateTo('research'),
      description: 'Alt+R — исследования',
    },
    {
      key: 'm',
      alt: true,
      handler: () => navigateTo('messages'),
      description: 'Alt+M — сообщения',
    },
    {
      key: 'Escape',
      handler: () => navigateTo('overview'),
      description: 'Esc — вернуться на главный экран',
    },
  ]);

  const planets = useQuery({
    queryKey: ['planets'],
    queryFn: () => api.get<{ planets: Planet[] }>('/api/planets'),
    refetchInterval: 60000,
  });
  const unread = useQuery({
    queryKey: ['messages-unread'],
    queryFn: () => api.get<{ count: number }>('/api/messages/unread-count'),
    refetchInterval: 15000,
  });
  const me = useQuery({
    queryKey: ['me'],
    queryFn: () => api.get<{ user_id: string; username: string; role: string; credit: number }>('/api/me'),
    staleTime: 60000,
  });
  const incoming = useQuery({
    queryKey: ['fleets-incoming'],
    queryFn: () => api.get<{ fleets: Array<{ id: string; dst_galaxy: number; dst_system: number; dst_position: number; dst_is_moon: boolean; arrive_at: string }> }>('/api/fleet/incoming'),
    refetchInterval: 15000,
  });

  const unreadCount = unread.data?.count ?? 0;
  const isAdmin = me.data?.role === 'admin' || me.data?.role === 'superadmin';
  const [currentPlanetId, setCurrentPlanetId] = useState<string | null>(null);
  const [fleetDst, setFleetDst] = useState<{ g: number; s: number; pos: number; isMoon: boolean; mission: number } | undefined>();
  const [infoUnit, setInfoUnit] = useState<InfoUnit | null>(() => parseHash().infoUnit);

  function openInfo(kind: InfoUnit['kind'], id: number, level: number) {
    const unit: InfoUnit = { kind, id, level, fromTab: tab };
    setInfoUnit(unit);
    setTab('unit-info');
    history.pushState(null, '', `#unit-info/${kind}/${id}`);
  }

  const list = planets.data?.planets ?? [];
  const planet = list.find((p) => p.id === currentPlanetId) ?? list[0];

  if (planets.isLoading) return <LoadingSkeleton />;
  if (planets.isError) return (
    <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', height: '100vh', flexDirection: 'column', gap: 16 }}>
      <div style={{ fontSize: 40 }}>⚠️</div>
      <div style={{ color: 'var(--ox-fg-dim)' }}>Ошибка загрузки. Попробуйте обновить страницу.</div>
      <button type="button" className="btn" onClick={() => window.location.reload()}>Обновить</button>
    </div>
  );
  if (!planet) return <LoadingSkeleton />;

  const navItems = buildNavItems(t, unreadCount, isAdmin);

  const incomingFleets = (incoming.data?.fleets ?? []).filter(
    (f) => new Date(f.arrive_at).getTime() > Date.now()
  );

  return (
    <div className="ox-layout">
      <Header
        planet={planet}
        planets={list}
        homePlanetId={list[0]?.id}
        onPlanetChange={setCurrentPlanetId}
        onLogout={logout}
        username={me.data?.username ?? ''}
        {...(me.data?.credit !== undefined ? { credit: me.data.credit } : {})}
      />

      {incomingFleets.length > 0 && (
        <div style={{ padding: '0 12px 8px', display: 'flex', flexDirection: 'column', gap: 6 }}>
          {incomingFleets.map((f) => (
            <div key={f.id} style={{
              padding: '8px 14px', borderRadius: 6,
              background: 'rgba(239,68,68,0.12)',
              border: '1px solid rgba(239,68,68,0.6)',
              display: 'flex', alignItems: 'center', gap: 10,
              animation: 'ox-pulse-border 1.5s ease-in-out infinite',
              fontSize: 13,
            }}>
              <span style={{ fontSize: 18, flexShrink: 0 }}>⚠️</span>
              <span style={{ color: 'var(--ox-danger)', fontWeight: 600, flex: 1 }}>Атака на [{f.dst_galaxy}:{f.dst_system}:{f.dst_position}]{f.dst_is_moon ? ' 🌑' : ''}</span>
              <span style={{ fontSize: 12, fontFamily: 'var(--ox-mono)', color: 'var(--ox-fg-dim)', flexShrink: 0 }}>
                <Countdown finishAt={f.arrive_at} />
              </span>
            </div>
          ))}
        </div>
      )}

      <div className="ox-body">
        {/* Десктоп сайдбар */}
        <nav className="ox-sidebar">
          {navItems.map((item) =>
            item.sep
              ? <div key={item.key} className="ox-sidebar-sep" />
              : item.groupLabel
              ? <div key={item.key} className="ox-sidebar-group-label">{item.groupLabel}</div>
              : (
                <button
                  key={item.key}
                  type="button"
                  className="ox-nav-btn"
                  aria-pressed={tab === item.key}
                  onClick={() => navigateTo(item.key as Tab)}
                  title={item.label}
                >
                  <span className="icon">{item.icon}</span>
                  <span className="label">{item.label}</span>
                  {!!item.badge && <span className="badge">{item.badge}</span>}
                </button>
              )
          )}
        </nav>

        {/* Контент */}
        <main className="ox-content">
          <Suspense fallback={<ScreenSkeleton />}>
            {tab === 'overview'   && <OverviewScreen onShowPlanetOptions={() => navigateTo('planet-options')} />}
            {tab === 'buildings'  && <BuildingsScreen planet={planet} onOpenInfo={(id, lvl) => openInfo('building', id, lvl)} />}
            {tab === 'research'   && <ResearchScreen planet={planet} onOpenInfo={(id, lvl) => openInfo('research', id, lvl)} />}
            {tab === 'unit-info'  && infoUnit && (
              <UnitInfoScreen
                kind={infoUnit.kind}
                unitId={infoUnit.id}
                currentLevel={infoUnit.level}
                planetId={planet.id}
              />
            )}
            {tab === 'shipyard'   && <ShipyardScreen planet={planet} onOpenInfo={(kind, id) => openInfo(kind, id, 0)} />}
            {tab === 'repair'     && <RepairScreen planet={planet} />}
            {tab === 'galaxy'     && <GalaxyScreen homePlanet={planet} userId={me.data?.user_id ?? ''} planets={list} onFleetMission={(fg, fs, fpos, fMoon, fMission) => { setFleetDst({ g: fg, s: fs, pos: fpos, isMoon: fMoon, mission: fMission }); navigateTo('fleet'); }} />}
            {tab === 'fleet'      && <FleetScreen planet={planet} {...(fleetDst ? { initialDst: fleetDst } : {})} />}
            {tab === 'market'     && <MarketScreen planet={planet} />}
            {tab === 'rockets'    && <RocketsScreen planet={planet} />}
            {tab === 'messages'   && <MessagesScreen onFleetMission={(g, s, pos, isMoon, mission) => { setFleetDst({ g, s, pos, isMoon, mission }); navigateTo('fleet'); }} />}
            {tab === 'artefacts'  && <ArtefactsScreen />}
            {tab === 'art-market' && <ArtefactMarketScreen />}
            {tab === 'officers'   && <OfficersScreen />}
            {tab === 'achievements' && <AchievementsScreen />}
            {tab === 'score'      && <ScoreScreen />}
            {tab === 'alliance'   && <AllianceScreen />}
            {tab === 'chat'       && <ChatScreen />}
            {tab === 'sim'        && <BattleSimScreen />}
            {tab === 'admin'      && isAdmin && <AdminScreen />}
            {tab === 'planet-options' && <PlanetOptionsScreen planet={planet} planets={list} homePlanetId={list[0]?.id ?? null} onBack={() => navigateTo('overview')} />}
            {tab === 'resource'   && <ResourceScreen planetId={planet.id} />}
          </Suspense>
        </main>
      </div>

      {/* Мобильная нижняя навигация */}
      <BottomNav tab={tab} setTab={navigateTo} unreadCount={unreadCount} />

      <footer className="ox-footer">
        <small>oxsar-nova v0.1.0 — dev preview</small>
      </footer>
    </div>
  );
}

/* ── Server Clock ── */
function useServerClock() {
  const [time, setTime] = useState(() => new Date());
  useEffect(() => {
    const id = setInterval(() => setTime(new Date()), 1000);
    return () => clearInterval(id);
  }, []);
  return time;
}

/* ── Header ── */
function Header({
  planet, planets, homePlanetId, onPlanetChange, onLogout, username, credit,
}: {
  planet: Planet;
  planets: Planet[];
  homePlanetId: string | undefined;
  onPlanetChange: (id: string) => void;
  onLogout: () => void;
  username: string;
  credit?: number | undefined;
}) {
  const metal    = planet.metal    ?? 0;
  const silicon  = planet.silicon  ?? 0;
  const hydrogen = planet.hydrogen ?? 0;
  const metalRate    = planet.metal_per_sec    ?? 0;
  const siliconRate  = planet.silicon_per_sec  ?? 0;
  const hydrogenRate = planet.hydrogen_per_sec ?? 0;
  const metalCap    = planet.metal_cap    ?? 0;
  const siliconCap  = planet.silicon_cap  ?? 0;
  const hydrogenCap = planet.hydrogen_cap ?? 0;
  const energyProd      = planet.energy_prod      ?? 0;
  const energyRemaining = planet.energy_remaining ?? 0;
  const clock    = useServerClock();
  const timeStr  = clock.toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit', second: '2-digit' });

  function fmtCap(v: number): string {
    if (v >= 1_000_000) return `${(v / 1_000_000).toFixed(1)}M`;
    if (v >= 1_000) return `${Math.round(v / 1_000)}k`;
    return String(Math.round(v));
  }

  function capColor(cur: number, cap: number): string {
    if (cap <= 0) return 'var(--ox-fg-muted)';
    const ratio = cur / cap;
    if (ratio >= 1) return 'var(--ox-danger)';
    if (ratio >= 0.9) return 'var(--ox-warn, #f59e0b)';
    return 'var(--ox-fg-muted)';
  }

  return (
    <header className="ox-header">
      <div className="ox-header-logo">✦ OXSAR</div>

      <div className="ox-header-resources">
        <div className="ox-res-item">
          <span className="icon">🟠</span>
          <span className="label-sm">Мет</span>
          <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'flex-end', lineHeight: 1.1 }}>
            <ResourceTicker value={metal} ratePerSec={metalRate} cap={metalCap} />
            {metalCap > 0 && <span style={{ fontSize: 10, color: capColor(metal, metalCap), fontFamily: 'var(--ox-mono)' }}>{fmtCap(metalCap)}</span>}
          </div>
        </div>
        <div className="ox-res-item">
          <span className="icon">💎</span>
          <span className="label-sm">Крем</span>
          <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'flex-end', lineHeight: 1.1 }}>
            <ResourceTicker value={silicon} ratePerSec={siliconRate} cap={siliconCap} />
            {siliconCap > 0 && <span style={{ fontSize: 10, color: capColor(silicon, siliconCap), fontFamily: 'var(--ox-mono)' }}>{fmtCap(siliconCap)}</span>}
          </div>
        </div>
        <div className="ox-res-item">
          <span className="icon">💧</span>
          <span className="label-sm">Водор</span>
          <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'flex-end', lineHeight: 1.1 }}>
            <ResourceTicker value={hydrogen} ratePerSec={hydrogenRate} cap={hydrogenCap} />
            {hydrogenCap > 0 && <span style={{ fontSize: 10, color: capColor(hydrogen, hydrogenCap), fontFamily: 'var(--ox-mono)' }}>{fmtCap(hydrogenCap)}</span>}
          </div>
        </div>
        <div className="ox-res-item ox-energy-item" style={{ color: energyRemaining < 0 ? 'var(--ox-danger)' : 'var(--ox-fg-dim)' }}>
          <span className="icon">⚡</span>
          <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'flex-end', lineHeight: 1.1 }}>
            <span style={{ fontSize: 13, fontFamily: 'var(--ox-mono)', fontWeight: 600 }}>{Math.round(energyProd)}</span>
            {energyProd > 0 && (
              <span style={{ fontSize: 10, fontFamily: 'var(--ox-mono)', color: energyRemaining < 0 ? 'var(--ox-danger)' : 'var(--ox-success, #22c55e)' }}>
                {energyRemaining >= 0 ? '+' : ''}{Math.round(energyRemaining)}
              </span>
            )}
          </div>
        </div>
      </div>

      <div className="ox-header-right">
        {credit !== undefined && (
          <div className="ox-res-item" title="Кредиты">
            <span className="icon">💳</span>
            <span style={{ fontFamily: 'var(--ox-mono)', fontSize: 13, fontWeight: 600, color: 'var(--ox-accent)' }}>
              {credit % 1 === 0 ? credit : credit.toFixed(2)}
            </span>
          </div>
        )}
        <span style={{ fontSize: 12, fontFamily: 'var(--ox-mono)', color: 'var(--ox-fg-dim)', letterSpacing: '0.04em' }}>
          {timeStr}
        </span>
        <PlanetSwitcher planet={planet} planets={planets} homePlanetId={homePlanetId} onChange={onPlanetChange} />
        {username && (
          <span style={{ fontSize: 12, color: 'var(--ox-fg-dim)', marginLeft: 4 }}>
            {username}
          </span>
        )}
        <button type="button" className="btn-ghost btn-sm" onClick={onLogout}>Выйти</button>
      </div>
    </header>
  );
}

/* ── Planet Switcher ── */
function PlanetSwitcher({
  planet, planets, homePlanetId, onChange,
}: {
  planet: Planet;
  planets: Planet[];
  homePlanetId: string | undefined;
  onChange: (id: string) => void;
}) {
  const [open, setOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!open) return;
    const close = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false);
    };
    document.addEventListener('mousedown', close);
    return () => document.removeEventListener('mousedown', close);
  }, [open]);

  return (
    <div style={{ position: 'relative' }} ref={ref}>
      <div
        className="ox-planet-switcher"
        onClick={() => setOpen((v) => !v)}
        role="button"
        tabIndex={0}
        onKeyDown={(e) => e.key === 'Enter' && setOpen((v) => !v)}
      >
        <span className="ox-planet-switcher-icon">{planet.is_moon ? '🌑' : '🪐'}</span>
        <span style={{ fontSize: 13, fontWeight: 600 }}>{planet.name}</span>
        {planet.id === homePlanetId && <span style={{ fontSize: 11 }}>🏠</span>}
        <span className="ox-planet-switcher-coords">
          [{planet.galaxy}:{planet.system}:{planet.position}]
        </span>
        <span style={{ fontSize: 10, color: 'var(--ox-fg-muted)', marginLeft: 2 }}>▾</span>
      </div>

      {open && planets.length > 1 && (
        <div className="ox-planet-dropdown">
          {planets.map((p) => (
            <div
              key={p.id}
              className={`ox-planet-option${p.id === planet.id ? ' active' : ''}`}
              onClick={() => { onChange(p.id); setOpen(false); }}
            >
              <span>{p.is_moon ? '🌑' : '🪐'}</span>
              <span style={{ flex: 1 }}>{p.name}{p.id === homePlanetId && ' 🏠'}</span>
              <span style={{ fontFamily: 'var(--ox-mono)', fontSize: 11, color: 'var(--ox-fg-muted)' }}>
                [{p.galaxy}:{p.system}:{p.position}]
              </span>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

/* ── Bottom Nav (mobile) ── */
const BOTTOM_ITEMS: Array<{ key: Tab; icon: string; label: string }> = [
  { key: 'overview',  icon: '🏠', label: 'Обзор' },
  { key: 'galaxy',    icon: '🌌', label: 'Галакт.' },
  { key: 'fleet',     icon: '🛸', label: 'Флот' },
  { key: 'messages',  icon: '📨', label: 'Сообщ.' },
];

function BottomNav({ tab, setTab, unreadCount }: { tab: Tab; setTab: (t: Tab) => void; unreadCount: number }) {
  const [sheetOpen, setSheetOpen] = useState(false);
  return (
    <>
      <nav className="ox-bottom-nav">
        {BOTTOM_ITEMS.map((item) => (
          <button
            key={item.key}
            type="button"
            className="ox-bottom-nav-btn"
            aria-pressed={tab === item.key}
            onClick={() => setTab(item.key)}
          >
            <span className="nav-icon">{item.icon}</span>
            <span>{item.label}</span>
            {item.key === 'messages' && unreadCount > 0 && (
              <span className="badge">{unreadCount}</span>
            )}
          </button>
        ))}
        <button type="button" className="ox-bottom-nav-btn" onClick={() => setSheetOpen(true)}>
          <span className="nav-icon">⋯</span>
          <span>Ещё</span>
        </button>
      </nav>

      {sheetOpen && (
        <div className="ox-modal-overlay" onClick={() => setSheetOpen(false)}>
          <div
            className="ox-modal"
            style={{ maxWidth: '100%', borderRadius: '16px 16px 0 0', padding: '16px 0' }}
            onClick={(e) => e.stopPropagation()}
          >
            <div style={{ padding: '0 16px 12px', fontSize: 11, fontWeight: 700, letterSpacing: '0.08em', textTransform: 'uppercase', color: 'var(--ox-fg-muted)' }}>
              Навигация
            </div>
            <MoreSheet tab={tab} setTab={(t) => { setTab(t); setSheetOpen(false); }} />
          </div>
        </div>
      )}
    </>
  );
}

const ALL_NAV: Array<{ key: Tab; icon: string; label: string }> = [
  { key: 'overview',    icon: '🏠', label: 'Обзор' },
  { key: 'resource',    icon: '⚙️', label: 'Сырьё' },
  { key: 'buildings',   icon: '🏗', label: 'Постройки' },
  { key: 'research',    icon: '🔬', label: 'Исследования' },
  { key: 'shipyard',    icon: '🚀', label: 'Верфь' },
  { key: 'repair',      icon: '🔧', label: 'Ремонт' },
  { key: 'galaxy',      icon: '🌌', label: 'Галактика' },
  { key: 'fleet',       icon: '🛸', label: 'Флот' },
  { key: 'rockets',     icon: '💥', label: 'Ракеты' },
  { key: 'messages',    icon: '📨', label: 'Сообщения' },
  { key: 'chat',        icon: '💬', label: 'Чат' },
  { key: 'alliance',    icon: '🤝', label: 'Альянс' },
  { key: 'market',      icon: '💱', label: 'Рынок' },
  { key: 'artefacts',   icon: '💎', label: 'Артефакты' },
  { key: 'art-market',  icon: '🏪', label: 'Арт-рынок' },
  { key: 'officers',    icon: '⭐', label: 'Офицеры' },
  { key: 'score',       icon: '🏆', label: 'Рейтинг' },
  { key: 'achievements',icon: '🥇', label: 'Достижения' },
  { key: 'sim',         icon: '⚔️', label: 'Симулятор' },
];

function MoreSheet({ tab, setTab }: { tab: Tab; setTab: (t: Tab) => void }) {
  return (
    <div style={{ maxHeight: '60vh', overflowY: 'auto' }}>
      {ALL_NAV.map((item) => (
        <button
          key={item.key}
          type="button"
          onClick={() => setTab(item.key)}
          style={{
            display: 'flex', alignItems: 'center', gap: 12,
            width: '100%', padding: '12px 20px',
            background: tab === item.key ? 'var(--ox-bg-active)' : 'transparent',
            border: 'none', borderRadius: 0,
            color: tab === item.key ? 'var(--ox-accent)' : 'var(--ox-fg)',
            fontSize: 15, fontWeight: 600, cursor: 'pointer',
            textAlign: 'left', minHeight: 48,
          }}
        >
          <span style={{ fontSize: 20, width: 28, textAlign: 'center' }}>{item.icon}</span>
          {item.label}
        </button>
      ))}
    </div>
  );
}

/* ── Навигационные пункты ── */
function buildNavItems(t: (ns: string, key: string, fb?: string) => string, unreadCount: number, isAdmin: boolean) {
  return [
    { key: 'planet', groupLabel: 'Планета' },
    { key: 'overview',   icon: '🏠', label: t('global','MENU_MAIN') },
    { key: 'resource',   icon: '⚙️', label: t('global','MENU_RESOURCE') },
    { key: 'buildings',  icon: '🏗', label: t('global','MENU_CONSTRUCTIONS') },
    { key: 'research',   icon: '🔬', label: t('global','MENU_RESEARCH') },
    { key: 'shipyard',   icon: '🚀', label: t('global','MENU_SHIPYARD') },
    { key: 'repair',     icon: '🔧', label: t('global','MENU_REPAIR') },
    { key: 's1', sep: true },
    { key: 'space', groupLabel: 'Космос' },
    { key: 'galaxy',     icon: '🌌', label: t('global','MENU_GALAXY') },
    { key: 'fleet',      icon: '🛸', label: t('global','MENU_FLEET') },
    { key: 'rockets',    icon: '💥', label: t('global','MENU_ROCKETS') },
    { key: 's2', sep: true },
    { key: 'social', groupLabel: 'Общение' },
    { key: 'messages',   icon: '📨', label: t('global','MENU_MESSAGES'), badge: unreadCount || undefined },
    { key: 'chat',       icon: '💬', label: 'Чат' },
    { key: 'alliance',   icon: '🤝', label: t('global','MENU_ALLIANCE') || 'Альянс' },
    { key: 's3', sep: true },
    { key: 'trade', groupLabel: 'Торговля' },
    { key: 'market',     icon: '💱', label: t('global','MENU_MARKET') },
    { key: 'artefacts',  icon: '💎', label: t('global','MENU_ARTEFACTS') },
    { key: 'art-market', icon: '🏪', label: t('global','MENU_ART_MARKET') },
    { key: 'officers',   icon: '⭐', label: t('global','MENU_OFFICERS') },
    { key: 's4', sep: true },
    { key: 'stats', groupLabel: 'Статистика' },
    { key: 'score',      icon: '🏆', label: t('global','MENU_HIGHSCORE') || 'Рейтинг' },
    { key: 'achievements',icon:'🥇', label: t('global','MENU_ACHIEVEMENTS') || 'Достижения' },
    { key: 'sim',        icon: '⚔️', label: t('global','MENU_SIMULATOR') },
    ...(isAdmin ? [
      { key: 's5', sep: true },
      { key: 'admin', icon: '🛠', label: 'Админ' },
    ] : []),
  ];
}

/* ── Скелетоны загрузки ── */
function LoadingSkeleton() {
  return (
    <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', minHeight: '100vh', flexDirection: 'column', gap: 16 }}>
      <div style={{ fontSize: 32, fontWeight: 700, letterSpacing: '0.1em', color: 'var(--ox-accent)', textShadow: '0 0 30px rgba(56,189,248,0.4)', animation: 'ox-float 2s ease-in-out infinite' }}>
        ✦ OXSAR
      </div>
      <div style={{ color: 'var(--ox-fg-muted)', fontSize: 13, letterSpacing: '0.1em', textTransform: 'uppercase' }}>
        Загрузка вселенной…
      </div>
    </div>
  );
}
