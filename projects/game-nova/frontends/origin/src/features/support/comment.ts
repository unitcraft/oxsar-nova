// Pure-функция сборки comment для S-045 Support (план 72 Ф.5 Spring 4
// ч.2). Вынесена в отдельный модуль чтобы тест не тащил
// SupportScreen → api/support → useAuthStore → localStorage.

export interface SupportFields {
  login: string;
  universe: string;
  page: string;
  browser: string;
  description: string;
  steps: string;
}

export const EMPTY_SUPPORT_FIELDS: SupportFields = {
  login: '',
  universe: 'Oxsar Classic',
  page: '',
  browser: '',
  description: '',
  steps: '',
};

export function buildSupportComment(fields: SupportFields): string {
  const lines: string[] = [];
  if (fields.login) lines.push(`Логин: ${fields.login}`);
  if (fields.universe) lines.push(`Вселенная: ${fields.universe}`);
  if (fields.page) lines.push(`Страница: ${fields.page}`);
  if (fields.browser) lines.push(`Браузер: ${fields.browser}`);
  if (fields.description) {
    lines.push('');
    lines.push('Описание:');
    lines.push(fields.description);
  }
  if (fields.steps) {
    lines.push('');
    lines.push('Шаги воспроизведения:');
    lines.push(fields.steps);
  }
  return lines.join('\n');
}
