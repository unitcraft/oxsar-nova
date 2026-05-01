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

import { useEffect, useState } from 'react';
import { useParams, Link } from 'react-router-dom';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { fetchArtefactCatalog } from '@/api/catalog';
import {
  fetchArtefacts,
  activateArtefact,
  deactivateArtefact,
} from '@/api/artefacts';
import { QK } from '@/api/query-keys';
import type { ApiError } from '@/api/client';
import type { ArtefactState } from '@/api/types';
import { useTranslation } from '@/i18n/i18n';
import { formatDuration } from '@/lib/format';
import {
  ARTEFACT_FALLBACK_IMAGE,
  artefactImageUrl,
  artefactImageUrlFallback,
} from '../common/artefact-catalog';

// План 72.1.38: live countdown для expire/delay (legacy artefactinfo.tpl
// jQuery countdown). Обновляется каждую секунду пока компонент жив.
function useCountdown(targetIso: string | null | undefined): string {
  const [now, setNow] = useState(() => Date.now());
  useEffect(() => {
    if (!targetIso) return;
    const id = setInterval(() => setNow(Date.now()), 1000);
    return () => clearInterval(id);
  }, [targetIso]);
  if (!targetIso) return '';
  const ms = new Date(targetIso).getTime() - now;
  if (ms <= 0) return '00:00:00';
  const sec = Math.floor(ms / 1000) % 60;
  const min = Math.floor(ms / 60_000) % 60;
  const hrs = Math.floor(ms / 3_600_000) % 24;
  const days = Math.floor(ms / 86_400_000);
  return days > 0
    ? `${days}d ${pad2(hrs)}:${pad2(min)}:${pad2(sec)}`
    : `${pad2(hrs)}:${pad2(min)}:${pad2(sec)}`;
}
function pad2(n: number): string {
  return n < 10 ? `0${n}` : String(n);
}

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
  const qc = useQueryClient();
  const [actErr, setActErr] = useState<string | null>(null);

  // План 72.1.38: activate/deactivate прямо со страницы инфо
  // (legacy artefactinfo.tpl L.234-257).
  const activateMut = useMutation({
    mutationFn: (artId: string) => activateArtefact(artId),
    onSuccess: () => {
      setActErr(null);
      void qc.invalidateQueries({ queryKey: QK.artefacts() });
    },
    onError: (e) => setActErr((e as ApiError).message),
  });
  const deactivateMut = useMutation({
    mutationFn: (artId: string) => deactivateArtefact(artId),
    onSuccess: () => {
      setActErr(null);
      void qc.invalidateQueries({ queryKey: QK.artefacts() });
    },
    onError: (e) => setActErr((e as ApiError).message),
  });

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
        {/* План 72.1.38: показ боевых множителей (legacy
            artefactinfo.tpl L.125-136 effect_type flags). */}
        {entry.effect && (entry.effect.battle_attack ||
          entry.effect.battle_shield || entry.effect.battle_shell) && (
          <tr>
            <th colSpan={2}>
              {t('artefacts', 'battleEffect') || 'Боевой эффект'}
            </th>
          </tr>
        )}
        {entry.effect?.battle_attack && (
          <tr>
            <td>{t('assaultReport', 'gunPower') || 'Атака'}</td>
            <td>×{entry.effect.battle_attack}</td>
          </tr>
        )}
        {entry.effect?.battle_shield && (
          <tr>
            <td>{t('assaultReport', 'shieldPower') || 'Щит'}</td>
            <td>×{entry.effect.battle_shield}</td>
          </tr>
        )}
        {entry.effect?.battle_shell && (
          <tr>
            <td>{t('assaultReport', 'armoring') || 'Броня'}</td>
            <td>×{entry.effect.battle_shell}</td>
          </tr>
        )}

        {myCopies.length > 0 && (
          <>
            <tr>
              <th colSpan={2}>{t('artefacts', 'groupHeld')}</th>
            </tr>
            {myCopies.map((a) => (
              <CopyRow
                key={a.id}
                artId={a.id}
                state={a.state}
                {...(a.expire_at ? { expireAt: a.expire_at } : {})}
                onActivate={() => activateMut.mutate(a.id)}
                onDeactivate={() => deactivateMut.mutate(a.id)}
                pending={activateMut.isPending || deactivateMut.isPending}
                t={t}
              />
            ))}
            {actErr && (
              <tr>
                <td colSpan={2}>
                  <span className="false">{actErr}</span>
                </td>
              </tr>
            )}
          </>
        )}
      </tbody>
    </table>
  );
}

type TFunc = (group: string, key: string, vars?: Record<string, string>) => string;

// План 72.1.38: row для одной копии с countdown-таймером и кнопками
// activate/deactivate.
function CopyRow({
  artId,
  state,
  expireAt,
  onActivate,
  onDeactivate,
  pending,
  t,
}: {
  artId: string;
  state: ArtefactState;
  expireAt?: string;
  onActivate: () => void;
  onDeactivate: () => void;
  pending: boolean;
  t: TFunc;
}) {
  const countdown = useCountdown(state === 'active' ? expireAt : undefined);
  return (
    <tr key={artId}>
      <td>
        {t('artefacts', stateLabelKey(state))}
        {countdown && (
          <>
            {' · '}
            <code>{countdown}</code>
          </>
        )}
        {expireAt && !countdown && (
          <>
            {' · '}
            {t('artefacts', 'expires')}{' '}
            {new Date(expireAt).toLocaleString('ru-RU')}
          </>
        )}
      </td>
      <td className="center" style={{ width: '30%' }}>
        {state === 'held' && (
          <button
            type="button"
            className="button"
            disabled={pending}
            onClick={onActivate}
          >
            {t('artefacts', 'activate') || 'Активировать'}
          </button>
        )}
        {state === 'active' && (
          <button
            type="button"
            className="button"
            disabled={pending}
            onClick={onDeactivate}
          >
            {t('artefacts', 'deactivate') || 'Отключить'}
          </button>
        )}
        {(state === 'delayed' || state === 'expired' || state === 'consumed') && (
          <Link to="/artefacts">{t('artefacts', 'title')}</Link>
        )}
      </td>
    </tr>
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
