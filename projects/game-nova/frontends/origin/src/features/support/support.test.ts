import { describe, it, expect } from 'vitest';
import { buildSupportComment } from './comment';

describe('support', () => {
  it('собирает comment из полей с метками', () => {
    const c = buildSupportComment({
      login: 'alice',
      universe: 'Oxsar Classic',
      page: '/galaxy/4/120',
      browser: 'Firefox 130',
      description: 'Не работает кнопка отправки флота.',
      steps: '1. Открыть galaxy\n2. Нажать «Атаковать»',
    });
    expect(c).toContain('Логин: alice');
    expect(c).toContain('Вселенная: Oxsar Classic');
    expect(c).toContain('Страница: /galaxy/4/120');
    expect(c).toContain('Браузер: Firefox 130');
    expect(c).toContain('Описание:');
    expect(c).toContain('Не работает кнопка отправки флота.');
    expect(c).toContain('Шаги воспроизведения:');
  });

  it('пустые поля пропускаются', () => {
    const c = buildSupportComment({
      login: '',
      universe: '',
      page: '',
      browser: '',
      description: 'Только тело.',
      steps: '',
    });
    expect(c).not.toContain('Логин:');
    expect(c).not.toContain('Шаги воспроизведения:');
    expect(c).toContain('Только тело.');
  });
});
