// S-021 Artefact market — placeholder, реализация в Spring 2 ч.2.
import { useTranslation } from '@/i18n/i18n';

export function MarketScreen() {
  const { t } = useTranslation();
  return <div className="idiv">{t('artefacts', 'title')}</div>;
}
