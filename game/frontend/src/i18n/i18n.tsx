// Клиентский i18n. Повторяет семантику backend/internal/i18n (см. §10.3 ТЗ).
// Источник данных — /api/i18n/{lang}, который возвращает полный
// словарь локали (~1500 ключей, ~80kb JSON).
//
// Кеш двухуровневый:
//   1) TanStack Query — staleTime=1ч, переживает смену компонентов
//      внутри сессии.
//   2) localStorage — переживает F5/перезапуск dev-сервера. Важно
//      для HMR: при правках Vite триггерит full reload, и без
//      localStorage первый рендер ждал бы сеть заново. С кешом ~80kb
//      JSON парсится из строки за <10 мс.
// TanStack Query прогревается из localStorage через initialData
// (запрос считается fresh и не делает сразу fetch), но в фоне всё
// равно делает revalidate при следующем использовании — свежие
// переводы подтянутся без явного управления.
//
// Ключ storage — версия + язык. Поднимая LOCALE_VERSION, инвалидируем
// кеш у всех пользователей (например, после смены формата группы).
//
// Правила совпадают с бэком:
//   1) locales[lang][group][key] — прямой хит
//   2) locales[fallback][group][key] — fallback на ru
//   3) "[group.key]" — маркер отсутствия ключа (виден в UI, чтобы
//      разработчик сразу заметил и добавил перевод)
//
// Плейсхолдеры — только именованные {{name}}, подставляются через vars.

import { createContext, useContext, useMemo, type ReactNode } from 'react';
import { useQuery } from '@tanstack/react-query';
import { api } from '@/api/client';

const LOCALE_VERSION = 2;
const LOCALE_STORAGE_PREFIX = 'oxsar.locale.v';

function storageKey(lang: string): string {
  return `${LOCALE_STORAGE_PREFIX}${LOCALE_VERSION}.${lang}`;
}

function readCachedLocale(lang: string): LocaleDict | null {
  if (typeof window === 'undefined') return null;
  try {
    const raw = window.localStorage.getItem(storageKey(lang));
    if (!raw) return null;
    return JSON.parse(raw) as LocaleDict;
  } catch {
    return null;
  }
}

function writeCachedLocale(lang: string, dict: LocaleDict): void {
  if (typeof window === 'undefined') return;
  try {
    window.localStorage.setItem(storageKey(lang), JSON.stringify(dict));
  } catch {
    // QuotaExceeded/Private Mode — молча игнорируем, кеш это не
    // критичный путь, просто не ускорит следующий старт.
  }
}

export type Lang = 'ru' | 'en';
export const FALLBACK_LANG: Lang = 'ru';

export type LocaleDict = Record<string, Record<string, string>>;

interface I18nContextValue {
  lang: Lang;
  dict: LocaleDict | null;
  fallback: LocaleDict | null;
  loading: boolean;
}

const I18nContext = createContext<I18nContextValue | null>(null);

// I18nProvider — корневой провайдер. Грузит словарь выбранного языка
// и fallback (если они различны). Использует TanStack Query +
// localStorage (initialData): после первого визита словарь живёт
// между F5 и HMR-reload'ами. В фоне Query всё равно revalidate-ит,
// свежие ключи подхватятся без управления.
export function I18nProvider({ lang, children }: { lang: Lang; children: ReactNode }) {
  const dict = useQuery({
    queryKey: ['i18n', lang],
    queryFn: async () => {
      const data = await api.get<LocaleDict>(`/api/i18n/${lang}`);
      writeCachedLocale(lang, data);
      return data;
    },
    staleTime: 1000 * 60 * 60, // час
    retry: 1,
    initialData: () => readCachedLocale(lang) ?? undefined,
  });
  const fallback = useQuery({
    queryKey: ['i18n', FALLBACK_LANG],
    queryFn: async () => {
      const data = await api.get<LocaleDict>(`/api/i18n/${FALLBACK_LANG}`);
      writeCachedLocale(FALLBACK_LANG, data);
      return data;
    },
    staleTime: 1000 * 60 * 60,
    enabled: lang !== FALLBACK_LANG,
    initialData: () => readCachedLocale(FALLBACK_LANG) ?? undefined,
  });

  const value = useMemo<I18nContextValue>(
    () => ({
      lang,
      dict: dict.data ?? null,
      fallback: lang === FALLBACK_LANG ? (dict.data ?? null) : (fallback.data ?? null),
      loading: dict.isLoading || fallback.isLoading,
    }),
    [lang, dict.data, dict.isLoading, fallback.data, fallback.isLoading],
  );

  return <I18nContext.Provider value={value}>{children}</I18nContext.Provider>;
}

// useTranslation возвращает `t` и `tf`.
//
// t(group, key, vars?) — основная форма.
// Если группа известна заранее, удобнее передать её в параметре хука:
// useTranslation('auth') → t('myKey').
//
// tf(group, key, fallback, vars?) — форма с fallback. Используется,
// когда ключ может отсутствовать в словаре (новый текст ещё не добавлен).
// Вместо "[group.key]" возвращается fallback.
//
// vars — именованные плейсхолдеры {{name}} → значение.
export function useTranslation(defaultGroup?: string) {
  const ctx = useContext(I18nContext);
  if (!ctx) {
    throw new Error('useTranslation must be used inside <I18nProvider>');
  }

  function t(groupOrKey: string, keyOrVars?: string | Record<string, string>, vars?: Record<string, string>): string {
    let group: string;
    let key: string;
    let v: Record<string, string> | undefined;

    if (defaultGroup && typeof keyOrVars !== 'string') {
      group = defaultGroup;
      key = groupOrKey;
      v = keyOrVars;
    } else {
      group = groupOrKey;
      key = (keyOrVars as string) ?? '';
      v = vars;
    }

    const template = lookup(ctx.dict, group, key) ?? lookup(ctx.fallback, group, key);
    if (template === null) {
      return `[${group}.${key}]`;
    }
    return interpolate(template, v);
  }

  function tf(group: string, key: string, fallback: string, vars?: Record<string, string>): string {
    const template = lookup(ctx.dict, group, key) ?? lookup(ctx.fallback, group, key) ?? fallback;
    return interpolate(template, vars);
  }

  return { t, tf, lang: ctx.lang, loading: ctx.loading };
}

function lookup(d: LocaleDict | null, group: string, key: string): string | null {
  if (!d) return null;
  const g = d[group];
  if (!g) return null;
  const v = g[key];
  return v === undefined ? null : v;
}

function interpolate(tmpl: string, vars?: Record<string, string>): string {
  if (!vars) return tmpl;
  return tmpl.replace(/\{\{(\w+)\}\}/g, (m, name: string) => vars[name] ?? m);
}
