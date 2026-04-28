// Биржа артефактов в nova-фронте (план 76).
// Использует backend плана 68 (/api/exchange/*). Не путать с legacy
// «artmarket» (отдельный endpoint /api/artefact-market) — это та же
// идея, но новая реализация на row-per-item + oxsarits.
//
// Один экран — три view'а через локальный state ('list' | 'detail' | 'create'),
// nova-style: внутренняя навигация без роутера (см. AllianceScreen
// как аналог). Это упрощает интеграцию: достаточно добавить tab
// 'exchange' в App.tsx.

import { useEffect, useState } from 'react';
import { exchangeApi, type ExchangeLot } from '@/api/exchange';
import { ExchangeListPage } from './ExchangeListPage';
import { ExchangeLotPage } from './ExchangeLotPage';
import { CreateLotPage } from './CreateLotPage';

type View =
  | { kind: 'list' }
  | { kind: 'detail'; id: string }
  | { kind: 'create' };

const HASH_PREFIX = '#exchange';

function parseSubHash(): View {
  if (typeof window === 'undefined') return { kind: 'list' };
  const hash = window.location.hash;
  if (!hash.startsWith(HASH_PREFIX)) return { kind: 'list' };
  const tail = hash.slice(HASH_PREFIX.length);
  if (tail === '/new') return { kind: 'create' };
  const m = tail.match(/^\/lots\/([0-9a-f-]+)$/i);
  if (m && m[1]) return { kind: 'detail', id: m[1] };
  return { kind: 'list' };
}

function buildHash(view: View): string {
  switch (view.kind) {
    case 'list':   return HASH_PREFIX;
    case 'create': return `${HASH_PREFIX}/new`;
    case 'detail': return `${HASH_PREFIX}/lots/${view.id}`;
  }
}

export function ExchangeScreen() {
  const [view, setView] = useState<View>(() => parseSubHash());

  // Реакция на back/forward — браузерная навигация уважается.
  useEffect(() => {
    const onPop = () => setView(parseSubHash());
    window.addEventListener('popstate', onPop);
    return () => window.removeEventListener('popstate', onPop);
  }, []);

  const navigate = (next: View) => {
    setView(next);
    history.pushState(null, '', buildHash(next));
  };

  if (view.kind === 'detail') {
    return (
      <ExchangeLotPage
        lotId={view.id}
        onBack={() => navigate({ kind: 'list' })}
      />
    );
  }
  if (view.kind === 'create') {
    return (
      <CreateLotPage
        onBack={() => navigate({ kind: 'list' })}
        onCreated={(lot: ExchangeLot) => navigate({ kind: 'detail', id: lot.id })}
      />
    );
  }
  return (
    <ExchangeListPage
      onOpenLot={(id) => navigate({ kind: 'detail', id })}
      onCreate={() => navigate({ kind: 'create' })}
    />
  );
}

// Reexport для удобства.
export { exchangeApi };
