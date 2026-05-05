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

import { useEffect, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  changeProfession,
  fetchProfessionMe,
  fetchProfessions,
} from '@/api/profession';
import { QK } from '@/api/query-keys';
import { useTranslation } from '@/i18n/i18n';
import { secondsUntil } from '@/lib/format';
import type { ApiError } from '@/api/client';
import type { Profession } from '@/api/types';

// План 72.1.58: имена эффектов 1:1 с legacy. В legacy
// `Profession.class.php:104` имя эффекта = имя соответствующего
// здания/исследования из catalog. Маппинг effectKey → info.<key>
// ниже зеркалит configs/professions.yml ↔ configs/i18n/info.*.
//
// Если ключ отсутствует — fallback на 'profession.bonus*' (старая
// группа), затем raw key.
const EFFECT_INFO_KEY: Record<string, string> = {
  metalmine: 'metalmine',
  silicon_lab: 'siliconLab',
  solar_plant: 'solarPlant',
  shipyard: 'shipyard',
  defense_factory: 'defenseFactory',
  rocket_station: 'rocketStation',
  // Боевые техи — отдельные исследования.
  gun: 'gunTech',
  shield_weapon: 'shieldTech',
  shell_weapon: 'shellTech',
  ballistics: 'ballisticsTech',
  masking: 'maskingTech',
  computer_tech: 'computerTech',
  // План 72.1.59: legacy UNIT_IGN (id=26) — Межгалактическая
  // исследовательская сеть. Используется в malus у Атакера.
  ign: 'ign',
  gravi: 'gravi',
  combustion_drive: 'combustionEngine',
  impulse_drive: 'impulseEngine',
  hyperspace_drive: 'hyperspaceEngine',
};

// План 72.1.59: порядок эффектов задаётся backend через ordered
// `effects` массив в DTO (1:1 с legacy `consts.php:$GLOBALS["PROFESSIONS"]`
// tech_special insertion-order). Frontend рендерит as-is без сортировки.

// Sentinel error codes от backend (план 72.1.58, см.
// internal/profession/handler.go::errCode*).
const ERROR_CODE_TO_I18N: Record<string, string> = {
  profession_unknown: 'errUnknownProfession',
  profession_not_enough_credit: 'errNotEnoughCredit',
  profession_in_vacation: 'errInVacation',
};

function fmtDelta(v: number): string {
  return v > 0 ? `+${v}` : String(v);
}

export function ProfessionScreen() {
  const { t } = useTranslation();
  const qc = useQueryClient();
  const [selected, setSelected] = useState<string | null>(null);
  const [errMsg, setErrMsg] = useState<string | null>(null);

  const listQ = useQuery({
    queryKey: QK.professions(),
    queryFn: fetchProfessions,
    // План 72.1.59: каталог редко меняется (только при деплое), но
    // 5 минут staleTime затрудняет fresh-данные после правки YAML.
    // 30 сек — компромисс.
    staleTime: 30_000,
  });

  const meQ = useQuery({
    queryKey: QK.professionMe(),
    queryFn: fetchProfessionMe,
    refetchInterval: 30_000,
  });

  // План 72.1.59: countdown «N дней HH:MM:SS» вместо округлённых дней.
  // Тикаем каждую секунду пока есть cooldown (next_change_allowed в
  // будущем). Хук должен быть до early-return — правило React Hooks.
  const nextChangeAllowed = meQ.data?.next_change_allowed ?? null;
  const [tick, setTick] = useState(0);
  useEffect(() => {
    if (!nextChangeAllowed) return;
    const remain = secondsUntil(nextChangeAllowed);
    if (remain <= 0) return;
    const id = setInterval(() => setTick((n) => n + 1), 1000);
    return () => clearInterval(id);
  }, [nextChangeAllowed]);
  // Использование tick для force re-render (React уже знает что state
  // изменился — countdownText пересчитается автоматически).
  void tick;

  const change = useMutation({
    mutationFn: changeProfession,
    onSuccess: () => {
      setErrMsg(null);
      void qc.invalidateQueries({ queryKey: QK.professionMe() });
      // План 72.1.59: backend списывает 1000 cr при cooldown. После
      // успешной смены инвалидируем /api/me — шапка обновит
      // отображаемый кредит без ручного refresh.
      void qc.invalidateQueries({ queryKey: QK.me() });
      // Скролл в начало страницы (legacy `Profession.class.php`
      // делает doHeaderRedirection — браузер начинает сверху).
      window.scrollTo({ top: 0, behavior: 'smooth' });
    },
    onError: (err) => {
      // План 72.1.58: backend возвращает sentinel-коды
      // (profession_unknown / not_enough_credit / in_vacation), которые
      // мы переводим через i18n. Generic-ошибки fallback на errGeneric.
      const apiErr = err as ApiError;
      const i18nKey = ERROR_CODE_TO_I18N[apiErr.code];
      if (i18nKey) {
        // Передаём cost в not_enough_credit (info из meQ).
        const cost = meQ.data?.change_cost ?? 0;
        setErrMsg(t('profession', i18nKey, { cost: String(cost) }));
      } else {
        setErrMsg(
          t('profession', 'errGeneric', { msg: apiErr.message }) ||
            apiErr.message,
        );
      }
    },
  });

  if (listQ.isLoading || meQ.isLoading) {
    return <div className="idiv">…</div>;
  }
  const professions: Profession[] = listQ.data?.professions ?? [];
  const me = meQ.data;
  const currentKey = me?.profession ?? 'none';
  // План 72.1.47: legacy позволяет менять профессию в любое время —
  // в cooldown стоимость = 1000 cr, после = 0. UI показывает динамику
  // через me.change_cost / me.days_remain (backend NS::getProfessionChangeCost).
  const changeCost = me?.change_cost ?? 0;
  const daysRemain = me?.days_remain ?? 0;

  const radioValue = selected ?? currentKey;

  function submit(e: React.FormEvent) {
    e.preventDefault();
    if (!radioValue || radioValue === currentKey) return;
    change.mutate(radioValue);
  }

  // План 72.1.58: legacy PROFESSION_CHANGE_*_INFO передают и days, и
  // cost в обоих ветках (бесплатной и платной). При cooldown=0 —
  // показываем дефолтные 14 дней / 1000 cr как info о следующем
  // ограничении (legacy NS::PROFESSION_CHANGE_MIN_DAYS=14, COST=1000).
  const PROFESSION_CHANGE_MIN_DAYS = 14;
  const PROFESSION_CHANGE_COST = 1000;
  // План 72.1.59: countdown-строка «D дней HH:MM:SS» вместо
  // округлённых дней. i18n-шаблон содержит {{days}} и {{time}}
  // отдельно, чтобы между ними поместить слово «дней» (ru) /
  // «days» (en). Источник — next_change_allowed (точный момент
  // конца cooldown'а).
  const remainSec = nextChangeAllowed ? secondsUntil(nextChangeAllowed) : 0;
  const remainDays = Math.floor(remainSec / 86400);
  const remainHours = Math.floor((remainSec % 86400) / 3600);
  const remainMinutes = Math.floor((remainSec % 3600) / 60);
  const remainSeconds = remainSec % 60;
  const pad = (n: number) => (n < 10 ? `0${n}` : String(n));
  const countdownTime = `${pad(remainHours)}:${pad(remainMinutes)}:${pad(remainSeconds)}`;
  const changeInfo = daysRemain > 0
    ? t('profession', 'changeCostInfo', {
        cost: String(changeCost),
        days: String(remainDays),
        time: countdownTime,
      })
    : t('profession', 'changeFreeInfo', {
        days: String(PROFESSION_CHANGE_MIN_DAYS),
        cost: String(PROFESSION_CHANGE_COST),
      });

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
            // План 72.1.59: backend отдаёт effects как ordered slice
            // в legacy-порядке (`consts.php:$GLOBALS["PROFESSIONS"]`).
            // Рендерим as-is, фильтруем нули.
            const specs = (p.effects ?? []).filter((e) => e.value !== 0);
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
                  />{' '}
                  <label htmlFor={`profession_${p.key}`}>
                    {/* План 72.1.58: legacy `profession.tpl` показывает
                        активную профессию ТОЛЬКО зелёным цветом
                        (<b class="true">), без подписи «(активна)».
                        План 72.1.59: имя берётся из i18n
                        (configs/i18n/{ru,en}.yml ключ profession.<key>Label). */}
                    <b className={isCurrent ? 'true' : undefined}>
                      {t('profession', `${p.key}Label`)}
                    </b>
                  </label>
                </td>
                <td className="profession-cell">
                  {/* План 72.1.59: описание из i18n profession.<key>Desc
                      (раньше приходило в DTO p.description из YAML).
                      DOM 1:1 с legacy `profession.tpl:22` —
                      `<p><label for=profession_id>{desc}</label></p>`.
                      Класс profession-cell даёт text-align: left для
                      содержимого этой ячейки (в legacy наследуется
                      от .main_content, в origin родитель #content
                      центрирован, нужен точечный override). */}
                  <p>
                    <label
                      htmlFor={`profession_${p.key}`}
                      style={{ cursor: 'pointer' }}
                    >
                      {t('profession', `${p.key}Desc`)}
                    </label>
                  </p>
                  {specs.length > 0 && (
                    <table
                      className="table_no_background"
                      cellSpacing={0}
                      cellPadding={0}
                      border={0}
                    >
                      <thead>
                        <tr>
                          {/* План 72.1.58: legacy PROFESSION_SPECIALISATION
                              «Специализация» вместо «Текущая профессия». */}
                          <th colSpan={2}>
                            {t('profession', 'specialization')}
                          </th>
                        </tr>
                      </thead>
                      <tbody>
                        {specs.map((e) => {
                          // План 72.1.58: имена эффектов 1:1 с legacy —
                          // имя соответствующего здания/исследования
                          // (info.<key>). Fallback на старую группу
                          // profession.bonus*, затем на raw key.
                          const infoKey = EFFECT_INFO_KEY[e.key];
                          let label: string;
                          if (infoKey) {
                            label = t('info', infoKey);
                            if (label.startsWith('[')) {
                              // не нашлось в info — fallback
                              label = e.key;
                            }
                          } else {
                            label = e.key;
                          }
                          return (
                            <tr key={e.key}>
                              <td>{label}</td>
                              <td className={e.value > 0 ? 'true' : 'false'}>
                                &nbsp;{fmtDelta(e.value)}
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
