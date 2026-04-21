import { Suspense, lazy, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { useAuthStore } from './stores/auth';
import { api } from './api/client';
import type { Planet } from './api/types';
import { LoginScreen } from './features/auth/LoginScreen';
import { OverviewScreen } from './features/overview/OverviewScreen';
import { useTranslation } from './i18n/i18n';

// Лениво грузим тяжёлые экраны: экран активируется — только тогда
// Vite/браузер подтягивают соответствующий чанк. При HMR правка одного
// экрана не влияет на модули остальных — пересборка и обновление
// браузера идут быстрее. Overview остаётся синхронным: это стартовая
// вкладка и самое частое место входа.
const BuildingsScreen = lazy(() =>
  import('./features/buildings/BuildingsScreen').then((m) => ({ default: m.BuildingsScreen })),
);
const ResearchScreen = lazy(() =>
  import('./features/research/ResearchScreen').then((m) => ({ default: m.ResearchScreen })),
);
const ShipyardScreen = lazy(() =>
  import('./features/shipyard/ShipyardScreen').then((m) => ({ default: m.ShipyardScreen })),
);
const ArtefactsScreen = lazy(() =>
  import('./features/artefacts/ArtefactsScreen').then((m) => ({ default: m.ArtefactsScreen })),
);
const BattleSimScreen = lazy(() =>
  import('./features/battle-sim/BattleSimScreen').then((m) => ({ default: m.BattleSimScreen })),
);
const GalaxyScreen = lazy(() =>
  import('./features/galaxy/GalaxyScreen').then((m) => ({ default: m.GalaxyScreen })),
);
const FleetScreen = lazy(() =>
  import('./features/fleet/FleetScreen').then((m) => ({ default: m.FleetScreen })),
);
const RepairScreen = lazy(() =>
  import('./features/repair/RepairScreen').then((m) => ({ default: m.RepairScreen })),
);

type Tab =
  | 'overview'
  | 'buildings'
  | 'research'
  | 'shipyard'
  | 'repair'
  | 'artefacts'
  | 'galaxy'
  | 'fleet'
  | 'sim';

export function App() {
  const token = useAuthStore((s) => s.accessToken);
  if (!token) {
    return (
      <Layout>
        <LoginScreen />
      </Layout>
    );
  }
  return (
    <Layout>
      <AuthenticatedApp />
    </Layout>
  );
}

function Layout({ children }: { children: React.ReactNode }) {
  const token = useAuthStore((s) => s.accessToken);
  const logout = useAuthStore((s) => s.logout);
  const { t } = useTranslation();
  return (
    <div className="ox-layout">
      <header className="ox-header">
        <h1>oxsar-nova</h1>
        <div className="ox-right">
          {token && (
            <button type="button" onClick={logout}>
              {t('global', 'MENU_LOGOUT')}
            </button>
          )}
        </div>
      </header>
      <main className="ox-main">{children}</main>
      <footer className="ox-footer">
        <small>v0.1.0 — dev preview</small>
      </footer>
    </div>
  );
}

function AuthenticatedApp() {
  const [tab, setTab] = useState<Tab>('overview');
  const { t } = useTranslation();
  const planets = useQuery({
    queryKey: ['planets'],
    queryFn: () => api.get<{ planets: Planet[] }>('/api/planets'),
    refetchInterval: 5000,
  });
  const [currentPlanetId, setCurrentPlanetId] = useState<string | null>(null);

  if (planets.isLoading) return <p>…</p>;
  if (planets.error)
    return (
      <p className="ox-error">
        {t('global', 'ERROR')}: {planets.error instanceof Error ? planets.error.message : ''}
      </p>
    );

  const list = planets.data?.planets ?? [];
  const planet = list.find((p) => p.id === currentPlanetId) ?? list[0];

  if (!planet) {
    return <p>{t('global', 'ERROR')}: no starter planet</p>;
  }

  return (
    <div>
      <PlanetSwitcher list={list} currentId={planet.id} onChange={setCurrentPlanetId} />

      <div className="ox-tabs">
        <TabButton current={tab} value="overview" onClick={setTab} label={t('global', 'MENU_MAIN')} />
        <TabButton
          current={tab}
          value="buildings"
          onClick={setTab}
          label={t('global', 'MENU_CONSTRUCTIONS')}
        />
        <TabButton
          current={tab}
          value="research"
          onClick={setTab}
          label={t('global', 'MENU_RESEARCH')}
        />
        <TabButton
          current={tab}
          value="shipyard"
          onClick={setTab}
          label={t('global', 'MENU_SHIPYARD')}
        />
        <TabButton
          current={tab}
          value="repair"
          onClick={setTab}
          label={t('global', 'MENU_REPAIR')}
        />
        <TabButton
          current={tab}
          value="galaxy"
          onClick={setTab}
          label={t('global', 'MENU_GALAXY')}
        />
        <TabButton
          current={tab}
          value="fleet"
          onClick={setTab}
          label={t('global', 'MENU_FLEET')}
        />
        <TabButton
          current={tab}
          value="artefacts"
          onClick={setTab}
          label={t('global', 'MENU_ARTEFACTS')}
        />
        <TabButton
          current={tab}
          value="sim"
          onClick={setTab}
          label={t('global', 'MENU_SIMULATOR')}
        />
      </div>

      <Suspense fallback={<p>…</p>}>
        {tab === 'overview' && <OverviewScreen />}
        {tab === 'buildings' && <BuildingsScreen planet={planet} />}
        {tab === 'research' && <ResearchScreen planet={planet} />}
        {tab === 'shipyard' && <ShipyardScreen planet={planet} />}
        {tab === 'repair' && <RepairScreen planet={planet} />}
        {tab === 'galaxy' && <GalaxyScreen homePlanet={planet} />}
        {tab === 'fleet' && <FleetScreen planet={planet} />}
        {tab === 'artefacts' && <ArtefactsScreen />}
        {tab === 'sim' && <BattleSimScreen />}
      </Suspense>
    </div>
  );
}

function PlanetSwitcher({
  list,
  currentId,
  onChange,
}: {
  list: Planet[];
  currentId: string;
  onChange: (id: string) => void;
}) {
  if (list.length < 2) {
    const p = list[0];
    return p ? (
      <div style={{ marginBottom: 12 }}>
        Планета: <b>{p.name}</b> [{p.galaxy}:{p.system}:{p.position}]
      </div>
    ) : null;
  }
  return (
    <div style={{ marginBottom: 12 }}>
      Планета:{' '}
      <select value={currentId} onChange={(e) => onChange(e.target.value)}>
        {list.map((p) => (
          <option key={p.id} value={p.id}>
            {p.name} [{p.galaxy}:{p.system}:{p.position}]
          </option>
        ))}
      </select>
    </div>
  );
}

function TabButton({
  current,
  value,
  onClick,
  label,
}: {
  current: Tab;
  value: Tab;
  onClick: (v: Tab) => void;
  label: string;
}) {
  return (
    <button type="button" aria-pressed={current === value} onClick={() => onClick(value)}>
      {label}
    </button>
  );
}
