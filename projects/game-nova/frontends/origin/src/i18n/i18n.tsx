// I18n origin-фронта (план 72 Ф.1).
//
// Полная переиспользуемость nova-bundle (R12 плана 72): origin-фронт
// тянет тот же словарь /api/i18n/{lang} что и nova-фронт. Отдельной
// шины ключей origin не имеет — это намеренно, чтобы:
//   1) не дублировать переводы;
//   2) при добавлении нового ключа разработчик origin сначала
//      grep'ал nova-bundle и максимально переиспользовал;
//   3) метрики плана 72 (переиспользовано/новых) считались
//      объективно.
//
// Семантика lookup идентична nova:
//   - locales[lang][group][key] — прямой хит
//   - locales[fallback][group][key] — fallback на ru
//   - "[group.key]" — маркер отсутствия ключа
//
// На Ф.1 реализуем минимальный API (Provider + useTranslation); кеш
// (localStorage + revalidate) можно поднять до nova-уровня в Ф.7
// при подключении первых экранов.

import { createContext, useContext, useMemo, type ReactNode } from 'react';
import { useQuery } from '@tanstack/react-query';
import { api } from '@/api/client';

export type Lang = 'ru' | 'en';
export const FALLBACK_LANG: Lang = 'ru';

type LocaleDict = Record<string, Record<string, string>>;

interface I18nContextValue {
  lang: Lang;
  dict: LocaleDict;
  fallbackDict: LocaleDict;
}

const I18nContext = createContext<I18nContextValue | null>(null);

function lookup(
  dict: LocaleDict,
  group: string,
  key: string,
): string | undefined {
  const groupDict = dict[group];
  if (!groupDict) return undefined;
  return groupDict[key];
}

function interpolate(
  template: string,
  vars?: Record<string, string | number>,
): string {
  if (!vars) return template;
  return template.replace(/\{\{(\w+)\}\}/g, (_, name: string) => {
    const v = vars[name];
    return v === undefined ? `{{${name}}}` : String(v);
  });
}

interface I18nProviderProps {
  lang: Lang;
  children: ReactNode;
}

export function I18nProvider({ lang, children }: I18nProviderProps) {
  const langQuery = useQuery({
    queryKey: ['i18n', lang],
    queryFn: () => api.get<LocaleDict>(`/api/i18n/${lang}`),
    staleTime: 60 * 60 * 1000,
  });

  const fallbackQuery = useQuery({
    queryKey: ['i18n', FALLBACK_LANG],
    queryFn: () => api.get<LocaleDict>(`/api/i18n/${FALLBACK_LANG}`),
    staleTime: 60 * 60 * 1000,
    enabled: lang !== FALLBACK_LANG,
  });

  const value = useMemo<I18nContextValue>(() => {
    const dict = langQuery.data ?? {};
    const fallbackDict =
      lang === FALLBACK_LANG ? dict : (fallbackQuery.data ?? {});
    return { lang, dict, fallbackDict };
  }, [lang, langQuery.data, fallbackQuery.data]);

  return <I18nContext.Provider value={value}>{children}</I18nContext.Provider>;
}

export function useTranslation() {
  const ctx = useContext(I18nContext);
  if (!ctx) throw new Error('useTranslation used outside I18nProvider');

  function t(
    group: string,
    key: string,
    vars?: Record<string, string | number>,
  ): string {
    const direct = lookup(ctx.dict, group, key);
    if (direct !== undefined) return interpolate(direct, vars);
    const fallback = lookup(ctx.fallbackDict, group, key);
    if (fallback !== undefined) return interpolate(fallback, vars);
    return `[${group}.${key}]`;
  }

  return { t, lang: ctx.lang };
}
