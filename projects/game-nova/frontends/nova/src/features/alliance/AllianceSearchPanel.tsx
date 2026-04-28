// План 67 Ф.5 ч.2 — фильтры списка альянсов (U-012).
//
// Контролируемый компонент: владельцем state'а является родитель
// (AllianceScreen, view='list'). Здесь — только UI + обработка ввода.
// Дебаунс реализован в родителе, чтобы query-key для TanStack Query
// строился один раз при изменении effective-фильтров, а не при каждом
// keystroke (иначе кеш TanStack Query разрастётся в свалку).

import { type AllianceSearchFilters, EMPTY_FILTERS, hasActiveFilters } from './search-filters';
import { useTranslation } from '@/i18n/i18n';

export function AllianceSearchPanel({
  value,
  onChange,
}: {
  value: AllianceSearchFilters;
  onChange: (next: AllianceSearchFilters) => void;
}) {
  const { t } = useTranslation('alliance');
  const showReset = hasActiveFilters(value);

  return (
    <div
      className="ox-panel"
      style={{
        padding: '10px 14px',
        display: 'flex',
        gap: 10,
        flexWrap: 'wrap',
        alignItems: 'flex-end',
      }}
    >
      <label style={{ display: 'flex', flexDirection: 'column', gap: 2, flex: '1 1 220px' }}>
        <span style={{ fontSize: 13, color: 'var(--ox-fg-dim)' }}>{t('search.q')}</span>
        <input
          value={value.q}
          onChange={(e) => onChange({ ...value, q: e.target.value })}
          placeholder={t('search.qPlaceholder')}
          maxLength={64}
        />
      </label>

      <label style={{ display: 'flex', flexDirection: 'column', gap: 2, flex: '0 0 140px' }}>
        <span style={{ fontSize: 13, color: 'var(--ox-fg-dim)' }}>{t('search.openness')}</span>
        <select
          value={value.isOpen}
          onChange={(e) =>
            onChange({ ...value, isOpen: e.target.value as AllianceSearchFilters['isOpen'] })
          }
        >
          <option value="all">{t('search.opennessAll')}</option>
          <option value="open">{t('search.opennessOpen')}</option>
          <option value="closed">{t('search.opennessClosed')}</option>
        </select>
      </label>

      <label style={{ display: 'flex', flexDirection: 'column', gap: 2, flex: '0 0 100px' }}>
        <span style={{ fontSize: 13, color: 'var(--ox-fg-dim)' }}>{t('search.minMembers')}</span>
        <input
          type="number"
          min={0}
          inputMode="numeric"
          value={value.minMembers}
          onChange={(e) => onChange({ ...value, minMembers: e.target.value })}
          placeholder="0"
        />
      </label>

      <label style={{ display: 'flex', flexDirection: 'column', gap: 2, flex: '0 0 100px' }}>
        <span style={{ fontSize: 13, color: 'var(--ox-fg-dim)' }}>{t('search.maxMembers')}</span>
        <input
          type="number"
          min={0}
          inputMode="numeric"
          value={value.maxMembers}
          onChange={(e) => onChange({ ...value, maxMembers: e.target.value })}
          placeholder="∞"
        />
      </label>

      {showReset && (
        <button
          type="button"
          className="btn-ghost btn-sm"
          onClick={() => onChange(EMPTY_FILTERS)}
        >
          ✕ {t('search.reset')}
        </button>
      )}
    </div>
  );
}
