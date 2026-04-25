import { useQuery } from '@tanstack/react-query';
import { useState } from 'react';
import { api } from '@/api/client';
import { renderMarkdown } from './md';

type Category = { key: string; title: string; order: number };
type Page = {
  path: string;
  frontmatter: Record<string, string>;
  markdown: string;
};
type CategoryPages = { category: string; pages: Page[] };

/**
 * WikiScreen — план 19 frontend MVP.
 * Левый сайдбар: категории + страницы. Правая часть: содержимое.
 */
export function WikiScreen() {
  const [activeCat, setActiveCat] = useState<string>('');
  const [activeSlug, setActiveSlug] = useState<string>('index');

  const cats = useQuery({
    queryKey: ['wiki-categories'],
    queryFn: () => api.get<{ categories: Category[] }>('/api/wiki'),
    staleTime: 5 * 60_000,
  });

  const pages = useQuery({
    queryKey: ['wiki-pages', activeCat],
    queryFn: () => api.get<CategoryPages>(`/api/wiki/${activeCat}`),
    enabled: !!activeCat,
    staleTime: 5 * 60_000,
  });

  const page = useQuery({
    queryKey: ['wiki-page', activeCat, activeSlug],
    queryFn: () => api.get<Page>(`/api/wiki/${activeCat || 'index'}/${activeSlug || 'index'}`),
    enabled: !!activeCat,
    staleTime: 5 * 60_000,
  });

  // Auto-pick first category once loaded.
  if (!activeCat && cats.data && cats.data.categories.length > 0) {
    const first = cats.data.categories[0];
    if (first) {
      setTimeout(() => setActiveCat(first.key), 0);
    }
  }

  return (
    <div style={{ display: 'flex', gap: 12, padding: 12, height: 'calc(100vh - 80px)' }}>
      {/* Sidebar */}
      <aside
        className="ox-panel"
        style={{ width: 240, flexShrink: 0, overflowY: 'auto', padding: 8 }}
      >
        <h3 style={{ margin: '4px 8px', fontSize: 14 }}>Вики</h3>
        {cats.isLoading && <div style={{ padding: 8 }}>Загрузка…</div>}
        {(cats.data?.categories ?? []).map((c) => (
          <div key={c.key} style={{ marginBottom: 4 }}>
            <button
              onClick={() => {
                setActiveCat(c.key);
                setActiveSlug('index');
              }}
              className={activeCat === c.key ? 'ox-btn ox-btn-primary' : 'ox-btn'}
              style={{
                width: '100%',
                justifyContent: 'flex-start',
                padding: '4px 8px',
                fontSize: 13,
              }}
            >
              {c.title}
            </button>
            {activeCat === c.key && pages.data && (
              <div style={{ marginLeft: 12, marginTop: 4, display: 'flex', flexDirection: 'column', gap: 2 }}>
                {pages.data.pages.map((p) => {
                  const slug = p.path.split('/')[1] ?? 'index';
                  const title = p.frontmatter.title ?? slug;
                  return (
                    <button
                      key={slug}
                      onClick={() => setActiveSlug(slug)}
                      className={activeSlug === slug ? 'ox-btn ox-btn-secondary' : 'ox-btn'}
                      style={{ padding: '2px 6px', fontSize: 12, justifyContent: 'flex-start' }}
                    >
                      {title}
                    </button>
                  );
                })}
              </div>
            )}
          </div>
        ))}
      </aside>

      {/* Content */}
      <main className="ox-panel" style={{ flex: 1, overflowY: 'auto', padding: 16 }}>
        {page.isLoading && <div>Загрузка статьи…</div>}
        {page.isError && <div>Не удалось загрузить статью.</div>}
        {page.data && (
          <article
            className="wiki-article"
            // Контент из docs/wiki/, контролируется командой проекта.
            // XSS не критичен, но html-инъекций мы избегаем через escapeHtml в md.ts.
            dangerouslySetInnerHTML={{ __html: renderMarkdown(page.data.markdown) }}
            style={{ lineHeight: 1.6 }}
          />
        )}
      </main>
    </div>
  );
}
