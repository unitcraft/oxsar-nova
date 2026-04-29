// План 72.1 ч.17: тесты планет-маппинга.
//
// Проверяем:
//   - детерминизм (один и тот же seed → один и тот же вариант),
//   - маппинг попадает в `01..NN` диапазон каждого типа,
//   - неизвестный тип → unformed,
//   - луна — фиксированный URL (mond.jpg, не зависит от seed).

import { describe, it, expect } from 'vitest';
import {
  planetImageUrl,
  planetImageSmallUrl,
  moonImageUrl,
  moonImageSmallUrl,
} from './planet-image';

const TYPES_AND_LIMITS: Array<[string, number]> = [
  ['dschjungelplanet', 10],
  ['eisplanet', 10],
  ['gasplanet', 8],
  ['normaltempplanet', 7],
  ['trockenplanet', 10],
  ['wasserplanet', 9],
  ['wuestenplanet', 4],
];

describe('planetImageUrl', () => {
  it('детерминирован по seed', () => {
    const a = planetImageUrl('wasserplanet', 'planet-id-001');
    const b = planetImageUrl('wasserplanet', 'planet-id-001');
    expect(a).toBe(b);
  });

  it('каждый planet_type попадает в свой диапазон 01..NN', () => {
    for (const [type, max] of TYPES_AND_LIMITS) {
      // Прогоняем 100 разных seed'ов чтобы накрыть все варианты.
      for (let i = 0; i < 100; i++) {
        const url = planetImageUrl(type, `seed-${i}`);
        const m = url.match(
          /\/assets\/origin\/images\/planets\/([a-z]+)(\d{2})\.jpg$/,
        );
        expect(m, `bad URL: ${url}`).toBeTruthy();
        expect(m![1]).toBe(type);
        const n = Number(m![2]);
        expect(n).toBeGreaterThanOrEqual(1);
        expect(n).toBeLessThanOrEqual(max);
      }
    }
  });

  it('неизвестный тип → unformed.jpg', () => {
    expect(planetImageUrl('unknown', 'x')).toBe(
      '/assets/origin/images/planets/unformed.jpg',
    );
    expect(planetImageUrl(null, 'x')).toBe(
      '/assets/origin/images/planets/unformed.jpg',
    );
    expect(planetImageUrl(undefined, 'x')).toBe(
      '/assets/origin/images/planets/unformed.jpg',
    );
  });
});

describe('planetImageSmallUrl', () => {
  it('детерминирован и в подкаталоге small/', () => {
    const url = planetImageSmallUrl('eisplanet', 'p-42');
    expect(url).toMatch(
      /^\/assets\/origin\/images\/planets\/small\/s_eisplanet\d{2}\.jpg$/,
    );
    expect(url).toBe(planetImageSmallUrl('eisplanet', 'p-42'));
  });
});

describe('moonImageUrl', () => {
  it('mond.jpg для большой', () => {
    expect(moonImageUrl()).toBe('/assets/origin/images/planets/mond.jpg');
  });
  it('s_mond.jpg для small', () => {
    expect(moonImageSmallUrl()).toBe(
      '/assets/origin/images/planets/small/s_mond.jpg',
    );
  });
});
