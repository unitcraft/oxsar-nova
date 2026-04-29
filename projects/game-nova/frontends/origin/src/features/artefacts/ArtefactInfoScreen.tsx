// S-014 ArtefactInfo (план 72 Ф.4 Spring 3).
//
// Pixel-perfect зеркало legacy `templates/standard/artefactinfo.tpl` —
// статическая страница каталог-описания одного артефакта.
//
// Endpoint:
//   GET /api/artefacts/catalog/{type}  → ArtefactCatalogEntry
//
// Дополнительно: GET /api/artefacts (мои) для блока «Местоположение»
// (показываем сколько копий этого типа у игрока + их state/expire).

import { useState } from 'react';
import { useParams, Link } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { fetchArtefactCatalog } from '@/api/catalog';
import { fetchArtefacts } from '@/api/artefacts';
import { QK } from '@/api/query-keys';
import type { ArtefactState } from '@/api/types';
import { useTranslation } from '@/i18n/i18n';
import { formatDuration } from '@/lib/format';
import {
  ARTEFACT_FALLBACK_IMAGE,
  artefactImageUrl,
  artefactImageUrlFallback,
} from '../common/artefact-catalog';

function stateLabelKey(state: ArtefactState): string {
  switch (state) {
    case 'held':
      return 'stateHeld';
    case 'active':
      return 'stateActive';
    case 'delayed':
      return 'stateDelayed';
    case 'expired':
      return 'stateExpired';
    case 'consumed':
      return 'stateConsumed';
  }
}

export function ArtefactInfoScreen() {
  const params = useParams<{ id?: string }>();
  const type = params.id ?? '';
  const { t } = useTranslation();

  const catQ = useQuery({
    queryKey: QK.artefactCatalog(type),
    queryFn: () => fetchArtefactCatalog(type),
    enabled: type.length > 0,
    staleTime: 60 * 60 * 1000,
  });

  const myQ = useQuery({
    queryKey: QK.artefacts(),
    queryFn: fetchArtefacts,
    staleTime: 30_000,
  });

  if (catQ.isLoading) return <div className="idiv">…</div>;
  if (catQ.isError || !catQ.data) {
    return (
      <table className="ntable">
        <tbody>
          <tr>
            <td className="center">
              <i>{t('alliance', 'nothing')}</i>
              {' · '}
              <Link to="/artefacts">{t('artefacts', 'title')}</Link>
            </td>
          </tr>
        </tbody>
      </table>
    );
  }

  const entry = catQ.data;
  const nameKey = entry.key.replace(/_([a-z])/g, (_, c: string) => c.toUpperCase());
  const name = t('info', nameKey);
  const hasI18n = name !== `[info.${nameKey}]`;
  const displayName = hasI18n ? name : entry.name || `${t('artefacts', 'toastArtefact')} #${entry.id}`;
  const fullDescKey = `${nameKey}FullDesc`;
  const descKey = `${nameKey}Desc`;
  const fullDesc = t('info', fullDescKey);
  const desc = t('info', descKey);
  const hasFull = fullDesc !== `[info.${fullDescKey}]`;
  const hasDesc = desc !== `[info.${descKey}]`;

  const myCopies = (myQ.data?.artefacts ?? []).filter(
    (a) => a.unit_id === entry.id,
  );

  return (
    <table className="ntable">
      <thead>
        <tr>
          <th colSpan={2}>{displayName}</th>
        </tr>
      </thead>
      <tbody>
        <tr>
          <td className="center" style={{ width: '120px' }}>
            <ArtefactInfoImage unitId={entry.id} alt={displayName} />
          </td>
          <td>
            {hasFull ? <span>{fullDesc}</span> : hasDesc ? <span>{desc}</span> : <i>—</i>}
            <br />
            {entry.lifetime_seconds > 0 && (
              <i>
                {t('artefacts', 'expires')} {formatDuration(entry.lifetime_seconds)}
              </i>
            )}
          </td>
        </tr>
        {myCopies.length > 0 && (
          <>
            <tr>
              <th colSpan={2}>{t('artefacts', 'groupHeld')}</th>
            </tr>
            {myCopies.map((a) => (
              <tr key={a.id}>
                <td>
                  {t('artefacts', stateLabelKey(a.state))}
                  {a.expire_at && (
                    <>
                      {' · '}
                      {t('artefacts', 'expires')}{' '}
                      {new Date(a.expire_at).toLocaleString('ru-RU')}
                    </>
                  )}
                </td>
                <td className="center" style={{ width: '20%' }}>
                  <Link to="/artefacts">{t('artefacts', 'title')}</Link>
                </td>
              </tr>
            ))}
          </>
        )}
      </tbody>
    </table>
  );
}

// Крупная картинка артефакта в S-014 (план 72.1 ч.17). Логика
// fallback'а та же что в S-013 ArtefactsScreen — gif → png →
// usable_artefact.gif.
function ArtefactInfoImage({ unitId, alt }: { unitId: number; alt: string }) {
  const primary = artefactImageUrl(unitId) ?? ARTEFACT_FALLBACK_IMAGE;
  const png = artefactImageUrlFallback(unitId);
  const [src, setSrc] = useState(primary);
  return (
    <img
      src={src}
      alt={alt}
      width={96}
      height={96}
      style={{ verticalAlign: 'middle' }}
      onError={() => {
        if (png && src !== png) {
          setSrc(png);
        } else if (src !== ARTEFACT_FALLBACK_IMAGE) {
          setSrc(ARTEFACT_FALLBACK_IMAGE);
        }
      }}
    />
  );
}
