// S-024 Fleet operations — placeholder, реализация в Spring 2 ч.2.
import { useTranslation } from '@/i18n/i18n';

export function FleetOperationsScreen() {
  const { t } = useTranslation();
  return <div className="idiv">{t('fleet', 'activeFleets', { count: 0 })}</div>;
}
