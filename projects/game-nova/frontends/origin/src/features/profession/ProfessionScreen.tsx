// S-041 Profession — выбор профессии (план 72 Ф.5 Spring 4 ч.2).
//
// Pixel-perfect зеркало legacy `templates/standard/profession.tpl`:
//   <form><table class="ntable">
//     <thead><tr><th colspan=2>Профессия</th></tr></thead>
//     <tfoot><tr><td colspan=2><input type=submit value=COMMIT/></td></tr></tfoot>
//     <tr><td colspan=2 class="false2">{change_info}</td></tr>
//     {foreach professions}
//       <tr>
//         <td><input type=radio> <label><b class="true|">{name}</b></label></td>
//         <td>{description} <table>{tech_special}</table></td>
//       </tr>
//
// Backend (internal/profession/) держит каталог в configs/professions.yml,
// возвращает {key, label, bonus, malus}. Профиль игрока в /me даёт
// текущую профессию + cooldown до следующей смены (14 дней).
//
// R9 Idempotency-Key — на смене (api/profession.ts).
// R12 i18n — ключи группы 'profession' существуют в configs/i18n/{ru,en}.yml.

import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  changeProfession,
  fetchProfessionMe,
  fetchProfessions,
} from '@/api/profession';
import { QK } from '@/api/query-keys';
import { useTranslation } from '@/i18n/i18n';
import type { ApiError } from '@/api/client';
import type { Profession } from '@/api/types';

// Бонусы/штрафы из profession.yml имеют технические ключи зданий и
// технологий. i18n-ключи 'profession.bonus*' существуют только для
// выборочных техн.ключей (план 67); для остальных показываем raw
// технический ключ как fallback (отображается без перевода) — это
// ожидаемо для редких профессий.
const BONUS_LABEL: Record<string, string> = {
  metalmine: 'bonusProductionMetal',
  silicon_lab: 'bonusProductionSilicon',
  solar_plant: 'bonusProductionHydrogen',
  shipyard: 'bonusBuildSpeed',
  gun: 'bonusShipAttack',
  shield_weapon: 'bonusShipShield',
  shell_weapon: 'bonusShipHull',
  ballistics: 'bonusFleetSpeed',
  masking: 'bonusEspionage',
  defense_factory: 'bonusBuildSpeed',
  rocket_station: 'bonusBuildSpeed',
  computer_tech: 'bonusResearchSpeed',
  gravi: 'bonusResearchSpeed',
  combustion_drive: 'bonusFleetSpeed',
  impulse_drive: 'bonusFleetSpeed',
  hyperspace_drive: 'bonusFleetSpeed',
};

function fmtDelta(v: number): string {
  return v > 0 ? `+${v}` : String(v);
}

function timeUntil(iso: string): string | null {
  const ms = new Date(iso).getTime() - Date.now();
  if (ms <= 0) return null;
  const d = Math.floor(ms / 86_400_000);
  const h = Math.floor((ms % 86_400_000) / 3_600_000);
  if (d > 0) return `${d}д ${h}ч`;
  const m = Math.floor((ms % 3_600_000) / 60_000);
  return `${h}ч ${m}м`;
}

export function ProfessionScreen() {
  const { t } = useTranslation();
  const qc = useQueryClient();
  const [selected, setSelected] = useState<string | null>(null);
  const [errMsg, setErrMsg] = useState<string | null>(null);

  const listQ = useQuery({
    queryKey: QK.professions(),
    queryFn: fetchProfessions,
    staleTime: 5 * 60_000,
  });

  const meQ = useQuery({
    queryKey: QK.professionMe(),
    queryFn: fetchProfessionMe,
    refetchInterval: 30_000,
  });

  const change = useMutation({
    mutationFn: changeProfession,
    onSuccess: () => {
      setErrMsg(null);
      void qc.invalidateQueries({ queryKey: QK.professionMe() });
    },
    onError: (err) => setErrMsg((err as ApiError).message),
  });

  if (listQ.isLoading || meQ.isLoading) {
    return <div className="idiv">…</div>;
  }
  const professions: Profession[] = listQ.data?.professions ?? [];
  const me = meQ.data;
  const currentKey = me?.profession ?? 'none';
  const blocked = me?.next_change_allowed
    ? timeUntil(me.next_change_allowed)
    : null;

  const radioValue = selected ?? currentKey;

  function submit(e: React.FormEvent) {
    e.preventDefault();
    if (!radioValue || radioValue === currentKey || blocked) return;
    change.mutate(radioValue);
  }

  // change_info зеркалит legacy: если cooldown активен — показать
  // оставшееся время; если можно сменить — выводим стоимость 1000 cr +
  // интервал 14 дней.
  const changeInfo = blocked
    ? t('profession', 'confirmChoose', { name: '—', days: '14' }) +
      ' ' +
      `(⏳ ${blocked})`
    : t('profession', 'confirmChoose', { name: '…', days: '14' });

  return (
    <form method="post" action="#" onSubmit={submit} data-testid="profession-form">
      <table className="ntable">
        <thead>
          <tr>
            <th colSpan={2}>{t('profession', 'title')}</th>
          </tr>
        </thead>
        <tfoot>
          <tr>
            <td className="center" colSpan={2}>
              <input
                type="submit"
                name="save"
                value={
                  change.isPending
                    ? '…'
                    : t('profession', 'confirmBtn')
                }
                className="button"
                disabled={
                  change.isPending ||
                  !!blocked ||
                  !radioValue ||
                  radioValue === currentKey
                }
              />
              {errMsg && (
                <div>
                  <span className="false">{errMsg}</span>
                </div>
              )}
            </td>
          </tr>
        </tfoot>
        <tbody>
          <tr>
            <td colSpan={2} className="false2">
              {changeInfo}
            </td>
          </tr>
          {professions.map((p) => {
            const isCurrent = p.key === currentKey;
            const bonusEntries = p.bonus
              ? Object.entries(p.bonus).filter(([, v]) => v !== 0)
              : [];
            const malusEntries = p.malus
              ? Object.entries(p.malus).filter(([, v]) => v !== 0)
              : [];
            const specs: Array<[string, number]> = [
              ...bonusEntries,
              ...malusEntries,
            ];
            return (
              <tr key={p.key}>
                <td style={{ whiteSpace: 'nowrap' }}>
                  <input
                    type="radio"
                    name="profession"
                    id={`profession_${p.key}`}
                    value={p.key}
                    checked={radioValue === p.key}
                    onChange={() => setSelected(p.key)}
                    disabled={!!blocked}
                  />{' '}
                  <label htmlFor={`profession_${p.key}`}>
                    <b className={isCurrent ? 'true' : undefined}>{p.label}</b>
                    {isCurrent && (
                      <>
                        {' '}
                        <span className="true">
                          ({t('profession', 'active')})
                        </span>
                      </>
                    )}
                  </label>
                </td>
                <td>
                  {p.description && (
                    <div style={{ marginBottom: 6 }}>{p.description}</div>
                  )}
                  {specs.length > 0 && (
                    <table
                      className="table_no_background"
                      cellSpacing={0}
                      cellPadding={0}
                      border={0}
                    >
                      <thead>
                        <tr>
                          <th colSpan={2}>
                            {t('profession', 'currentLabel')}
                          </th>
                        </tr>
                      </thead>
                      <tbody>
                        {specs.map(([k, v]) => {
                          const labelKey = BONUS_LABEL[k];
                          const label = labelKey
                            ? t('profession', labelKey)
                            : k;
                          return (
                            <tr key={k}>
                              <td>{label}</td>
                              <td className={v > 0 ? 'true' : 'false'}>
                                &nbsp;{fmtDelta(v)}
                              </td>
                            </tr>
                          );
                        })}
                      </tbody>
                    </table>
                  )}
                </td>
              </tr>
            );
          })}
        </tbody>
      </table>
    </form>
  );
}
