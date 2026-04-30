// S-024 BattleReport — детальный просмотр боя (план 72.1 ч.20.8).
//
// Использует общий компонент BattleReportView (из features/common),
// тот же который рендерит результат симулятора. Pixel-perfect клон
// legacy oxsar2-java/Assault.java HTML rendering.
//
// Спец-id `last-sim` — читает результат последней симуляции из
// localStorage (key 'oxsar-origin-last-sim'), не делает запрос к API.

import { useParams, Link } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { fetchBattleReport } from '@/api/battles';
import { QK } from '@/api/query-keys';
import { BattleReportView } from '@/features/common/BattleReportView';
import { useTranslation } from '@/i18n/i18n';
import type { SimReport } from '@/api/simulator';

const LAST_SIM_KEY = 'oxsar-origin-last-sim';

function readLastSim(): SimReport | null {
  try {
    const raw = localStorage.getItem(LAST_SIM_KEY);
    if (!raw) return null;
    return JSON.parse(raw) as SimReport;
  } catch {
    return null;
  }
}

export function BattleReportScreen() {
  const { id = '' } = useParams<{ id?: string }>();
  const { t } = useTranslation();

  // Спец-режим: показываем последнюю симуляцию из localStorage.
  if (id === 'last-sim') {
    const sim = readLastSim();
    if (!sim) {
      return (
        <table className="ntable">
          <tbody>
            <tr>
              <td className="center">
                <i>
                  {t('mission', 'noLastSim') ??
                    'Нет сохранённой симуляции. Запустите бой в симуляторе.'}
                </i>
                <br />
                <Link to="/simulator" className="button" style={{ marginTop: 8 }}>
                  → {t('mission', 'simulator') ?? 'Симулятор боя'}
                </Link>
              </td>
            </tr>
          </tbody>
        </table>
      );
    }
    return (
      <>
        <BattleReportView
          report={sim}
          title={t('mission', 'simulationResult') ?? 'Результат симуляции'}
        />
        <div style={{ marginTop: 12, textAlign: 'center' }}>
          <Link to="/simulator" className="button">
            ← {t('mission', 'simulator') ?? 'Симулятор боя'}
          </Link>
        </div>
      </>
    );
  }

  // Обычный режим: читаем настоящий бой по UUID.
  const q = useQuery({
    queryKey: QK.battleReport(id),
    queryFn: () => fetchBattleReport(id),
    enabled: id.length > 0,
    staleTime: 60 * 60 * 1000,
  });

  if (q.isLoading) return <div className="idiv">…</div>;
  if (q.isError || !q.data) {
    return (
      <table className="ntable">
        <tbody>
          <tr>
            <td className="center">
              <i>{t('alliance', 'nothing') ?? 'Нет доступа или отчёт не найден'}</i>
            </td>
          </tr>
        </tbody>
      </table>
    );
  }

  return (
    <>
      <BattleReportView
        report={q.data.report}
        title={t('battlestats', 'colResult') ?? 'Боевой отчёт'}
      />
      <div style={{ marginTop: 12, textAlign: 'center' }}>
        <Link to="/battlestats" className="button">
          ← {t('battlestats', 'title') ?? 'К списку боёв'}
        </Link>
      </div>
    </>
  );
}
