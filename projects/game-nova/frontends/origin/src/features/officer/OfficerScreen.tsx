// S-040 Officer — наём офицеров за credits (план 72 Ф.5 Spring 4 ч.2).
//
// Pixel-perfect зеркало legacy `templates/standard/officer.tpl`:
//   <form><table class="ntable">
//     <thead><tr><th colspan=3>Офицеры</th></tr></thead>
//     <tr><td><img/></td><td>описание + цена + срок</td>
//         <td>{кнопка нанять}</td></tr>
//     ...4 фиксированных officer-типов (торговый/шахтёр/энергетик/
//        кладовщик в legacy).
//
// Backend (internal/officer/) хранит officer-каталог в БД (таблица
// officer_def + INSERT в миграции) и возвращает Entry с полным
// набором полей (title/description/duration_days/cost_credit/effect/
// activated_at/expires_at). Поэтому фронт **не** дублирует каталог
// hardcoded'ом, а рендерит то, что отдал backend.
//
// R9 Idempotency-Key — внутри activateOfficer (api/officer.ts).
// R12 i18n — ключи группы 'officers' уже существуют в configs/i18n/
// {ru,en}.yml (план 67 nova).

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { activateOfficer, fetchOfficers } from '@/api/officer';
import { QK } from '@/api/query-keys';
import { useTranslation } from '@/i18n/i18n';
import type { ApiError } from '@/api/client';
import type { Officer } from '@/api/types';

const RELATIVE_URL = '/legacy-assets/';

function effectLabel(t: ReturnType<typeof useTranslation>['t'], key: string): string {
  switch (key) {
    case 'produce_factor':
      return t('officers', 'effectProduce');
    case 'build_factor':
      return t('officers', 'effectBuild');
    case 'research_factor':
      return t('officers', 'effectResearch');
    case 'energy_factor':
      return t('officers', 'effectEnergy');
    case 'storage_factor':
      return t('officers', 'effectStorage');
    default:
      return key;
  }
}

function formatEffect(
  t: ReturnType<typeof useTranslation>['t'],
  effect: Record<string, number> | null | undefined,
): string {
  if (!effect) return '';
  return Object.entries(effect)
    .filter(([, v]) => v !== 1)
    .map(([k, v]) => {
      const pct = Math.round((v - 1) * 100);
      return `${effectLabel(t, k)} ${pct > 0 ? '+' : ''}${pct}%`;
    })
    .join(', ');
}

function imageIndex(key: string): string {
  // Legacy officer.tpl нумерует картинки 01..04 в порядке отображения.
  // Для современных ключей (например 'trader', 'miner', 'energy',
  // 'storage') — мапим на ту же нумерацию.
  const order: Record<string, number> = {
    trader: 1,
    miner: 2,
    energy: 3,
    storage: 4,
  };
  const idx = order[key] ?? 1;
  return idx.toString().padStart(2, '0');
}

export function OfficerScreen() {
  const { t } = useTranslation();
  const qc = useQueryClient();

  const officersQ = useQuery({
    queryKey: QK.officers(),
    queryFn: fetchOfficers,
    refetchInterval: 30_000,
  });

  const activate = useMutation({
    mutationFn: (key: string) => activateOfficer(key, { auto_renew: false }),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: QK.officers() });
      void qc.invalidateQueries({ queryKey: QK.planets() });
    },
  });

  if (officersQ.isLoading) {
    return <div className="idiv">…</div>;
  }
  const list: Officer[] = officersQ.data?.officers ?? [];
  if (list.length === 0) {
    return (
      <table className="ntable">
        <thead>
          <tr>
            <th>{t('officers', 'title')}</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td className="center">{t('officers', 'empty')}</td>
          </tr>
        </tbody>
      </table>
    );
  }

  return (
    <form
      method="post"
      action="#"
      onSubmit={(e) => e.preventDefault()}
      data-testid="officer-form"
    >
      <table className="ntable">
        <thead>
          <tr>
            <th colSpan={3}>{t('officers', 'title')}</th>
          </tr>
        </thead>
        <tbody>
          {list.map((o) => {
            const active = !!o.expires_at;
            const expiresLabel = active && o.expires_at
              ? new Date(o.expires_at).toLocaleString('ru-RU')
              : '—';
            const eff = formatEffect(t, o.effect);
            return (
              <tr key={o.key}>
                <td width="120">
                  <img
                    src={`${RELATIVE_URL}images/officer/${imageIndex(o.key)}.jpg`}
                    alt={o.title}
                  />
                </td>
                <td>
                  <b>{o.title}</b>
                  <br />
                  {o.description}
                  {eff && (
                    <>
                      <p />
                      <span className="true">{eff}</span>
                    </>
                  )}
                  <p />
                  {t('officers', 'daysAndCost', {
                    days: String(o.duration_days),
                    cost: String(o.cost_credit),
                  })}
                  <p />
                  {active ? (
                    <>
                      {t('officers', 'expires')} {expiresLabel}
                    </>
                  ) : null}
                </td>
                <td width="50" className="center">
                  {active ? (
                    <span className="true">{t('officers', 'active')}</span>
                  ) : (
                    <button
                      type="button"
                      className="button"
                      disabled={activate.isPending}
                      onClick={() => activate.mutate(o.key)}
                    >
                      {t('officers', 'activateBtn', {
                        cost: String(o.cost_credit),
                      })}
                    </button>
                  )}
                </td>
              </tr>
            );
          })}
          {activate.isError && (
            <tr>
              <td colSpan={3} className="center">
                <span className="false">
                  {(activate.error as ApiError)?.message ??
                    t('officers', 'toastError')}
                </span>
              </td>
            </tr>
          )}
        </tbody>
      </table>
    </form>
  );
}
