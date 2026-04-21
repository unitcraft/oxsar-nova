import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { api } from '@/api/client';
import { useTranslation } from '@/i18n/i18n';
import type { Planet, SystemView } from '@/api/types';

// Экран галактики. Показывает 15 позиций в системе G:S.
// По умолчанию открывает координаты текущей планеты игрока.
//
// Layout reference: oxsar2/www/templates/standard/galaxy.tpl.
// В легаси этот экран — большая таблица, плюс ссылки «шпионить /
// атаковать / транспорт». В v1 делаем чистое чтение; fleet-действия
// подключаем следующей итерацией (M3-fleet).
export function GalaxyScreen({ homePlanet }: { homePlanet: Planet }) {
  const { t } = useTranslation();
  const [g, setG] = useState(homePlanet.galaxy);
  const [s, setS] = useState(homePlanet.system);

  const sys = useQuery({
    queryKey: ['galaxy', g, s],
    queryFn: () => api.get<SystemView>(`/api/galaxy/${g}/${s}`),
    refetchInterval: 10_000,
  });

  return (
    <section>
      <h2>
        {t('global', 'MENU_GALAXY')} [{g}:{s}]
      </h2>

      <div className="ox-form" style={{ display: 'flex', gap: 8, marginBottom: 12 }}>
        <label>
          G&nbsp;
          <input
            type="number"
            min={1}
            max={16}
            value={g}
            onChange={(e) => setG(clamp(Number(e.target.value), 1, 16))}
            style={{ width: 70 }}
          />
        </label>
        <label>
          S&nbsp;
          <input
            type="number"
            min={1}
            max={999}
            value={s}
            onChange={(e) => setS(clamp(Number(e.target.value), 1, 999))}
            style={{ width: 90 }}
          />
        </label>
        <button type="button" onClick={() => setG((v) => Math.max(1, v - 1))}>
          ←
        </button>
        <button type="button" onClick={() => setG((v) => Math.min(16, v + 1))}>
          →
        </button>
        <button type="button" onClick={() => setS((v) => Math.max(1, v - 1))}>
          ↑
        </button>
        <button type="button" onClick={() => setS((v) => Math.min(999, v + 1))}>
          ↓
        </button>
      </div>

      {sys.isLoading && <p>…</p>}
      {sys.error && (
        <p className="ox-error">
          {t('global', 'ERROR')}: {sys.error instanceof Error ? sys.error.message : ''}
        </p>
      )}

      {sys.data && (
        <table className="ox-table">
          <thead>
            <tr>
              <th>{t('Main', 'POSITION')}</th>
              <th>{t('Main', 'NEW_PLANET_NAME')}</th>
              <th>{t('Ranking', 'USERNAME', 'Игрок')}</th>
              <th>{t('Main', 'DEBRIS')}</th>
            </tr>
          </thead>
          <tbody>
            {sys.data.cells.map((c) => (
              <tr key={c.position}>
                <td className="num">{c.position}</td>
                <td>
                  {c.has_planet ? c.planet_name : '—'}
                  {c.has_moon ? ' 🌑' : ''}
                </td>
                <td>{c.owner_username ?? '—'}</td>
                <td className="num">
                  {c.debris_metal > 0 || c.debris_silicon > 0
                    ? `${formatNum(c.debris_metal)} / ${formatNum(c.debris_silicon)}`
                    : '—'}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </section>
  );
}

function clamp(v: number, lo: number, hi: number): number {
  if (Number.isNaN(v)) return lo;
  return Math.max(lo, Math.min(hi, v));
}

function formatNum(v: number): string {
  return Math.floor(v).toLocaleString('ru-RU');
}
