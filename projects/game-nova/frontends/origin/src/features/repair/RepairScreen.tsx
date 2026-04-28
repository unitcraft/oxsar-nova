// S-022 Repair — placeholder, реализация в Spring 2 ч.2.
import { useTranslation } from '@/i18n/i18n';

export function RepairScreen() {
  const { t } = useTranslation();
  return <div className="idiv">{t('repair', 'title', { planetName: '' })}</div>;
}
