import { useQuery } from '@tanstack/react-query';
import { useState } from 'react';
import { api } from '@/api/client';
import { useTranslation } from '@/i18n/i18n';

type Forecast = {
  hours: number;
  metal: number;
  silicon: number;
  hydrogen: number;
  capped: boolean;
};

const HOURS_OPTIONS = [1, 4, 12, 24];

function fmt(n: number): string {
  if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(1)}M`;
  if (n >= 1_000) return `${(n / 1_000).toFixed(1)}k`;
  return Math.round(n).toString();
}

/**
 * ForecastWidget — прогноз накопления ресурсов через N часов на планете
 * (план 17 G1). Показывает компактную таблицу.
 */
export function ForecastWidget({ planetID }: { planetID: string }) {
  const { t } = useTranslation('forecastUi');
  const [hours, setHours] = useState(4);

  const q = useQuery({
    queryKey: ['forecast', planetID, hours],
    queryFn: () => api.get<Forecast>(`/api/planets/${planetID}/forecast?hours=${hours}`),
    refetchInterval: 60_000,
  });

  const f = q.data;

  return (
    <div
      className="ox-panel"
      style={{
        padding: '10px 12px',
        margin: '8px 16px',
        display: 'flex',
        alignItems: 'center',
        gap: 12,
        flexWrap: 'wrap',
      }}
    >
      <span style={{ fontSize: 13, color: 'var(--ox-fg-muted)' }}>{t('forecastIn')}</span>
      {HOURS_OPTIONS.map((h) => (
        <button
          key={h}
          onClick={() => setHours(h)}
          className={hours === h ? 'ox-btn ox-btn-primary' : 'ox-btn'}
          style={{ padding: '2px 10px', fontSize: 13 }}
        >
          {h}{t('hourUnit')}
        </button>
      ))}
      {f && (
        <div style={{ display: 'flex', gap: 14, fontSize: 14 }}>
          <span>🟠 <strong>{fmt(f.metal)}</strong></span>
          <span>🔵 <strong>{fmt(f.silicon)}</strong></span>
          <span>🟢 <strong>{fmt(f.hydrogen)}</strong></span>
          {f.capped && (
            <span style={{ color: 'var(--ox-warning, #f5a623)' }} title={t('capTooltip')}>
              ⚠️ {t('capWarning')}
            </span>
          )}
        </div>
      )}
    </div>
  );
}
