// X-003: показ требований с конкретными именами и уровнями при
// can_build = false. Origin: если нет требований выполнить, вместо
// кнопки строительства выводится список требований («не хватает
// Энергетической технологии lvl 5»).
//
// В nova у нас уже есть unmet[] от backend (UnmetRequirement) или
// статичные requires[] от каталога. Этот компонент рендерит их в
// едином стиле, переиспользуется во всех экранах с can_build.

import type { UnmetRequirement } from '@/api/types';

// Одна несбывшаяся зависимость → строка вида:
// «🔒 Энергетическая технология ур.5 (у вас: 3)»
// Имя берём из переводов info-группы (как в существующих экранах).
interface UnmetRequirementListProps {
  unmet: UnmetRequirement[];
  // tInfo — функция перевода в группе 'info' (для имён построек/исследований).
  tInfo: (key: string) => string;
  // tBuildings — функция перевода для метки «у вас» / «ур.».
  tBuildings: (key: string) => string;
}

// keyToTKey — преобразует snake_case ключ из backend (`energy_tech`)
// в camelCase tKey (`energyTech`), используемый в i18n info-группе.
function keyToTKey(key: string): string {
  return key.replace(/_([a-z])/g, (_, c: string) => c.toUpperCase());
}

export function UnmetRequirementList({ unmet, tInfo, tBuildings }: UnmetRequirementListProps) {
  if (unmet.length === 0) return null;
  return (
    <div style={{ fontSize: 13, color: 'var(--ox-danger)', marginBottom: 4 }}>
      {unmet.map((r) => {
        const name = tInfo(keyToTKey(r.key));
        // Если ключа нет в info — t возвращает [info.xxx]; в этом
        // случае показываем raw-ключ, чтобы пользователь хотя бы
        // понял что блокирует.
        const display = name.startsWith('[info.') ? r.key : name;
        return (
          <div key={`${r.kind}-${r.key}`} style={{ fontFamily: 'var(--ox-mono)' }}>
            🔒 {display} {tBuildings('levelAbbr')}{r.required} ({tBuildings('youHave')} {r.current})
          </div>
        );
      })}
    </div>
  );
}
