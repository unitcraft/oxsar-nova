// S-023 Battlestats — placeholder, реализация в Spring 2 ч.2.
import { useTranslation } from '@/i18n/i18n';

export function BattleStatsScreen() {
  const { t } = useTranslation();
  return <div className="idiv">{t('battlestats', 'title')}</div>;
}
