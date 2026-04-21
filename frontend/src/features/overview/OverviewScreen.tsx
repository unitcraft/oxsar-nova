import { useQuery } from '@tanstack/react-query';
import { api } from '@/api/client';
import { useTranslation } from '@/i18n/i18n';
import type { Planet } from '@/api/types';

export function OverviewScreen() {
  const { t } = useTranslation();
  const { data, isLoading, error } = useQuery({
    queryKey: ['planets'],
    queryFn: () => api.get<{ planets: Planet[] }>('/api/planets'),
  });

  if (isLoading) return <p>…</p>;
  if (error)
    return (
      <p className="ox-error">
        {t('global', 'ERROR')}: {error instanceof Error ? error.message : ''}
      </p>
    );
  const planets = data?.planets ?? [];

  if (planets.length === 0) {
    return <p>{t('global', 'ERROR')}: no starter planet</p>;
  }

  return (
    <div>
      <table className="ox-table">
        <thead>
          <tr>
            <th>{t('Main', 'NEW_PLANET_NAME')}</th>
            <th>{t('Main', 'POSITION')}</th>
            <th>{t('global', 'METAL')}</th>
            <th>{t('global', 'SILICON')}</th>
            <th>{t('global', 'HYDROGEN')}</th>
          </tr>
        </thead>
        <tbody>
          {planets.map((p) => (
            <tr key={p.id}>
              <td>{p.name}</td>
              <td>
                [{p.galaxy}:{p.system}:{p.position}
                {p.is_moon ? ' 🌑' : ''}]
              </td>
              <td className="num">{formatNum(p.metal)}</td>
              <td className="num">{formatNum(p.silicon)}</td>
              <td className="num">{formatNum(p.hydrogen)}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function formatNum(v: number): string {
  return Math.floor(v).toLocaleString('ru-RU');
}
