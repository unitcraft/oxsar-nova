// S-R02 Defense — оборона (план 72.1 ч.20.3).
// Pixel-perfect клон legacy shipyard.tpl при ?go=Defense (unit_type=defense).

import { BuildPanel } from '@/features/common/BuildPanel';
import { useTranslation } from '@/i18n/i18n';

export function DefenseScreen() {
  const { t } = useTranslation();
  return (
    <BuildPanel
      group="defense"
      title={t('buildings', 'defense') ?? 'Оборона'}
    />
  );
}
