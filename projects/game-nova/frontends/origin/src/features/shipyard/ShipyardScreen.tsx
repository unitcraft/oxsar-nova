// S-004 Shipyard — верфь (план 72.1 ч.20.3).
// Pixel-perfect клон legacy shipyard.tpl (?go=Shipyard, type=fleet).

import { BuildPanel } from '@/features/common/BuildPanel';
import { useTranslation } from '@/i18n/i18n';

export function ShipyardScreen() {
  const { t } = useTranslation();
  return (
    <BuildPanel
      group="ship"
      title={t('buildings', 'shipConstruction') ?? 'Верфь'}
    />
  );
}
