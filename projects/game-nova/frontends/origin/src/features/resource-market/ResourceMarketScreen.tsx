// S-020 Resource market — placeholder, реализация в Spring 2 ч.2.
//
// Заглушка нужна, чтобы router.tsx из ч.1 типизировался и собирался.
// В ч.2 будет полная реализация resource exchange (legacy `resource.tpl`
// + 4 sub-template'а market_*.tpl).

import { useTranslation } from '@/i18n/i18n';

export function ResourceMarketScreen() {
  const { t } = useTranslation();
  return <div className="idiv">{t('market', 'title')}</div>;
}
