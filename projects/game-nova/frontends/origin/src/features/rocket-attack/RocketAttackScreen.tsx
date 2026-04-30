// S-RA Rocket Attack — ракетная атака (план 72.1 ч.20.5).
// Pixel-perfect клон legacy rocket_attack.tpl.
//
// Backend endpoint POST /api/planets/{id}/rocket-attack пока не
// реализован — ставим disabled submit + label «Backend pending»
// (требование п.11 без упрощений: TODO записан в части 20).

import { useSearchParams } from 'react-router-dom';
import { useState } from 'react';
import { useTranslation } from '@/i18n/i18n';
import { useResolvedPlanet } from '@/features/common/useResolvedPlanet';
import { catalogByGroup } from '@/features/common/catalog';

export function RocketAttackScreen() {
  const { t } = useTranslation();
  const { planet } = useResolvedPlanet();
  const [params] = useSearchParams();
  const [quantity, setQuantity] = useState('0');
  const [target, setTarget] = useState('all');

  const targetCoords = `[${params.get('g') ?? '?'}:${params.get('s') ?? '?'}:${params.get('p') ?? '?'}]`;
  const maxRockets = 0; // pending backend
  const defenseUnits = catalogByGroup('defense');

  if (!planet) return <div className="idiv">…</div>;

  return (
    <form
      method="post"
      action="#"
      onSubmit={(e) => {
        e.preventDefault();
      }}
    >
      <table className="ntable">
        <tbody>
          <tr>
            <th colSpan={2}>
              Ракетная атака: {targetCoords}
            </th>
          </tr>
          <tr>
            <td>
              {t('galaxy', 'quantity')} ({t('galaxy', 'max')} {maxRockets})
            </td>
            <td>
              <input
                type="text"
                name="quantity"
                value={quantity}
                onChange={(e) => setQuantity(e.target.value)}
              />
            </td>
          </tr>
          <tr>
            <td>{t('galaxy', 'primaryTarget')}</td>
            <td>
              <select
                name="target"
                value={target}
                onChange={(e) => setTarget(e.target.value)}
              >
                <option value="all">{t('galaxy', 'all')}</option>
                {defenseUnits.map((entry) => {
                  const [g, k] = entry.i18n.split('.') as [string, string];
                  return (
                    <option key={entry.id} value={String(entry.id)}>
                      {t(g, k)}
                    </option>
                  );
                })}
              </select>
            </td>
          </tr>
          <tr>
            <td colSpan={2} className="center">
              <input
                type="submit"
                name="start"
                value="Атаковать"
                className="button"
                disabled
                title="Backend pending"
              />
            </td>
          </tr>
        </tbody>
      </table>
    </form>
  );
}
