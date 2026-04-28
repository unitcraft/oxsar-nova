// S-047 AdvTechCalculator — калькулятор специализации (план 72 Ф.5
// Spring 4 ч.2).
//
// Pixel-perfect зеркало legacy `templates/standard/adv_tech_calc.tpl`:
//   таблица "Атакующий" / "Обороняющийся", по 3 технологии (ЛА/ИО/ПЛ),
//   расчёт распределения атаки и эффективности щитов.
//
// Pure client-side утилита — backend не нужен. Все формулы воспроизводят
// JS-логику adv_tech_calc.tpl 1:1 (это балансная формула, R0 не
// разрешает менять без ADR).
//
// Матрица взаимодействия (cfg_tech_*) и масштабы технологий
// (cfg_tech_scale_*) hardcoded'ом совпадают с дефолтом legacy.
// В будущем (если нужно) — выносим в catalog endpoint и читаем
// через TanStack Query.
//
// 3-канальная боевая система описана в memory: лазер/ион/плазма vs
// физ/магн/сил щит, ножницы; в MVP nova упрощено до scalar
// (см. memory `3-канальная боевая система — идея для расширения`).
// origin-калькулятор моделирует **legacy**-формулу как есть, потому
// что инструмент сравнения спецификаций нужен именно для legacy-
// совместимого расчёта.

import { useMemo, useState } from 'react';
import { useTranslation } from '@/i18n/i18n';
import {
  computeSide,
  EMPTY_SIDE,
  TECH_MATRIX,
  TECH_SCALE,
  type SideInputs,
  type SideResult,
} from './formula';

const TECH_NAMES = ['ЛА', 'ИО', 'ПЛ'] as const;

function SideForm({
  prefix,
  side,
  onChange,
  result,
  oppShield,
}: {
  prefix: string;
  side: SideInputs;
  onChange: (next: SideInputs) => void;
  result: SideResult;
  oppShield: SideResult;
}) {
  // shield_after = oppShield - mySide_attack * shots
  // shell_after = sum of negative shield_after.
  const shieldAfter: [number, number, number] = [0, 0, 0];
  let shellAfter = 0;
  for (let i = 0; i < 3; i++) {
    const v = (oppShield.shield[i] ?? 0) - (result.attack[i] ?? 0) * side.shots;
    shieldAfter[i] = v;
    if (v < 0) shellAfter += v;
  }

  function setTech(i: 0 | 1 | 2, v: number) {
    const tech: [number, number, number] = [...side.tech] as typeof side.tech;
    tech[i] = Math.max(0, Math.floor(v) || 0);
    onChange({ ...side, tech });
  }

  return (
    <td valign="top">
      <table
        cellSpacing={0}
        cellPadding={0}
        border={0}
        className="table_no_background"
      >
        <tbody>
          <tr>
            <td align="right">&nbsp;</td>
            <td colSpan={3}>{prefix === 'a' ? 'Атакующий' : 'Обороняющийся'}</td>
          </tr>
          <tr>
            <td align="right">Базовая атака</td>
            <td colSpan={3}>
              <input
                type="number"
                size={4}
                maxLength={5}
                value={side.baseAttack}
                onChange={(e) =>
                  onChange({ ...side, baseAttack: Number(e.target.value) || 0 })
                }
              />
            </td>
          </tr>
          <tr>
            <td align="right">Базовый щит</td>
            <td colSpan={3}>
              <input
                type="number"
                size={4}
                maxLength={5}
                value={side.baseShield}
                onChange={(e) =>
                  onChange({ ...side, baseShield: Number(e.target.value) || 0 })
                }
              />
            </td>
          </tr>
          <tr>
            <td align="right">Выстрелов</td>
            <td colSpan={3}>
              <input
                type="number"
                size={2}
                maxLength={2}
                value={side.shots}
                onChange={(e) =>
                  onChange({
                    ...side,
                    shots: Math.max(1, Number(e.target.value) || 1),
                  })
                }
              />
            </td>
          </tr>
          <tr>
            <td align="right">&nbsp;</td>
            {TECH_NAMES.map((n) => (
              <td key={n}>{n}</td>
            ))}
          </tr>
          <tr>
            <td align="right">Технология</td>
            {[0, 1, 2].map((i) => (
              <td key={i}>
                <input
                  type="number"
                  size={2}
                  maxLength={2}
                  value={side.tech[i as 0 | 1 | 2]}
                  onChange={(e) =>
                    setTech(i as 0 | 1 | 2, Number(e.target.value))
                  }
                />
              </td>
            ))}
          </tr>
          <tr>
            <td align="right">Атака %</td>
            {result.proc.map((v, i) => (
              <td key={i}>
                <input readOnly type="text" size={4} value={`${v}%`} />
              </td>
            ))}
          </tr>
          <tr>
            <td align="right">Атака</td>
            {result.attack.map((v, i) => (
              <td key={i}>
                <input readOnly type="text" size={4} value={v} />
              </td>
            ))}
          </tr>
          <tr>
            <td align="right">Щиты %</td>
            {result.shieldProc.map((v, i) => (
              <td key={i}>
                <input readOnly type="text" size={4} value={`${v}%`} />
              </td>
            ))}
          </tr>
          <tr>
            <td align="right">Щиты</td>
            {result.shield.map((v, i) => (
              <td key={i}>
                <input readOnly type="text" size={4} value={v} />
              </td>
            ))}
          </tr>
          <tr>
            <td align="right">Щитов после боя</td>
            {shieldAfter.map((v, i) => (
              <td key={i}>
                <input readOnly type="text" size={4} value={v} />
              </td>
            ))}
          </tr>
          <tr>
            <td align="right">Повреждение брони</td>
            <td colSpan={3}>
              <input readOnly type="text" size={4} value={shellAfter} />
            </td>
          </tr>
        </tbody>
      </table>
    </td>
  );
}

export function AdvTechCalculatorScreen() {
  useTranslation();
  const [att, setAtt] = useState<SideInputs>(EMPTY_SIDE);
  const [def, setDef] = useState<SideInputs>(EMPTY_SIDE);

  const attResult = useMemo(() => computeSide(att), [att]);
  const defResult = useMemo(() => computeSide(def), [def]);

  return (
    <table className="ntable">
      <thead>
        <tr>
          <th colSpan={2}>Калькулятор специализации</th>
        </tr>
      </thead>
      <tfoot>
        <tr>
          <td colSpan={2}>
            <table
              cellSpacing={0}
              cellPadding={0}
              border={0}
              className="table_no_background"
            >
              <tbody>
                <tr>
                  <td colSpan={4}>Влияние технологий</td>
                </tr>
                <tr>
                  <td>&nbsp;</td>
                  {TECH_NAMES.map((n) => (
                    <td key={n}>{n}</td>
                  ))}
                </tr>
                <tr>
                  <td>Коеф.</td>
                  {TECH_SCALE.map((v, i) => (
                    <td key={i}>{v}</td>
                  ))}
                </tr>
              </tbody>
            </table>
          </td>
        </tr>
        <tr>
          <td colSpan={2}>
            <table
              cellSpacing={0}
              cellPadding={0}
              border={0}
              className="table_no_background"
            >
              <tbody>
                <tr>
                  <td colSpan={4}>Эффективность щитов (атака × щит)</td>
                </tr>
                <tr>
                  <td>&nbsp;</td>
                  {TECH_NAMES.map((n) => (
                    <td key={`th-${n}`}>
                      {n}&nbsp;<sub>щиты</sub>
                    </td>
                  ))}
                </tr>
                {TECH_MATRIX.map((row, i) => (
                  <tr key={`row-${i}`}>
                    <td align="right">
                      {TECH_NAMES[i]}&nbsp;<sub>атака</sub>
                    </td>
                    {row.map((c, j) => (
                      <td key={`cell-${i}-${j}`}>{Math.round(c * 100)}%</td>
                    ))}
                  </tr>
                ))}
              </tbody>
            </table>
          </td>
        </tr>
        <tr>
          <td colSpan={2}>
            <table
              cellSpacing={0}
              cellPadding={0}
              border={0}
              className="table_no_background"
            >
              <tbody>
                <tr>
                  <td style={{ textAlign: 'left' }}>ЛА — лазерная технология</td>
                </tr>
                <tr>
                  <td style={{ textAlign: 'left' }}>ИО — ионная технология</td>
                </tr>
                <tr>
                  <td style={{ textAlign: 'left' }}>ПЛ — плазменная технология</td>
                </tr>
              </tbody>
            </table>
          </td>
        </tr>
      </tfoot>
      <tbody>
        <tr>
          <SideForm
            prefix="a"
            side={att}
            onChange={setAtt}
            result={attResult}
            oppShield={defResult}
          />
          <SideForm
            prefix="d"
            side={def}
            onChange={setDef}
            result={defResult}
            oppShield={attResult}
          />
        </tr>
      </tbody>
    </table>
  );
}
