// S-024 BattleReport — публичный анонимный просмотр боевого отчёта
// (план 72.1 ч.20.11).
//
// URL: /battle-report/{uuid}. Доступен без авторизации — любой
// пользователь по ссылке может посмотреть бой или симуляцию.
// Отчёты идентифицируются непредсказуемым UUID v7.
//
// План 72.1 ч.20.11.11: показываем «только сам бой» — без
// header'а «Результат — победитель …» и без кнопок навигации
// «К списку боёв / Симулятор боя» (last сейчас могут вести на
// закрытые роуты для гостя). Это совпадает с legacy 8092 страницей
// «Отчёт о сражении».

import { useParams } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { fetchBattleReport } from '@/api/battles';
import { QK } from '@/api/query-keys';
import { BattleReportView } from '@/features/common/BattleReportView';
import { useTranslation } from '@/i18n/i18n';

export function BattleReportScreen() {
  const { id = '' } = useParams<{ id?: string }>();
  const { t } = useTranslation();

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
              <i>{t('alliance', 'nothing') ?? 'Отчёт не найден'}</i>
            </td>
          </tr>
        </tbody>
      </table>
    );
  }

  return (
    <BattleReportView
      report={q.data.report}
      startedAt={q.data.started_at}
      compact
    />
  );
}
