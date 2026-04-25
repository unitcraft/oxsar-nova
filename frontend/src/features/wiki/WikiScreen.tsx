import { useQuery } from '@tanstack/react-query';
import { useEffect, useMemo, useRef, useState } from 'react';
import { api } from '@/api/client';
import { categoryOfId, imageOfId, keyOfId, nameOf } from '@/api/catalog';
import { renderMarkdown } from './md';
import './wiki.css';

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
 *
 * Внутри markdown поддерживается `[[unit:N]]` — рендерится как ссылка
 * с картинкой и красивым именем юнита; click переключает текущий
 * page внутри SPA-вкладки.
 */
export function WikiScreen() {
  const [activeCat, setActiveCat] = useState<string>('');
  const [activeSlug, setActiveSlug] = useState<string>('index');
  const contentRef = useRef<HTMLDivElement | null>(null);

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
  useEffect(() => {
    if (!activeCat && cats.data && cats.data.categories.length > 0) {
      const first = cats.data.categories[0];
      if (first) setActiveCat(first.key);
    }
  }, [cats.data, activeCat]);

  // resolveUnit: id → { name, image } для md-рендерера.
  const resolveUnit = useMemo(
    () => (id: number) => {
      const name = nameOf(id);
      const image = imageOfId(id);
      return { name, image };
    },
    []
  );

  // Делегированный click на ссылках [[unit:N]].
  // При клике — узнаём из id wiki-категорию и slug, переключаем активную страницу.
  useEffect(() => {
    const root = contentRef.current;
    if (!root) return;
    const onClick = (e: MouseEvent) => {
      let el = e.target as HTMLElement | null;
      while (el && !el.classList?.contains('wiki-unit-link')) {
        el = el.parentElement;
      }
      if (!el) return;
      e.preventDefault();
      const idStr = el.getAttribute('data-unit-id');
      if (!idStr) return;
      const id = Number(idStr);
      const cat = categoryOfId(id);
      const slug = keyOfId(id);
      if (!cat || !slug) return;
      setActiveCat(cat);
      setActiveSlug(slug);
    };
    root.addEventListener('click', onClick);
    return () => {
      root.removeEventListener('click', onClick);
    };
  }, [page.data]);

  // Заголовок страницы — иконка юнита + красивое имя (если frontmatter
  // содержит unit_id). Иначе title из frontmatter.
  const headerNode = useMemo(() => {
    const fm = page.data?.frontmatter ?? {};
    const unitID = fm.unit_id ? Number(fm.unit_id) : 0;
    if (unitID > 0) {
      const name = nameOf(unitID);
      const img = imageOfId(unitID);
      return (
        <div style={{ display: 'flex', alignItems: 'center', gap: 12, marginBottom: 12 }}>
          {img && <img src={img} alt="" style={{ width: 48, height: 48 }} />}
          <h1 style={{ margin: 0 }}>{name}</h1>
          <span style={{ color: 'var(--ox-fg-muted)', fontSize: 14 }}>id={unitID}</span>
        </div>
      );
    }
    if (fm.title) {
      return (
        <h1 style={{ margin: '0 0 12px' }}>{fm.title}</h1>
      );
    }
    return null;
  }, [page.data]);

  // Контент: убираем верхний `# heading` если показываем headerNode
  // (чтобы не дублировать с unit-id-заголовком).
  const bodyHtml = useMemo(() => {
    if (!page.data) return '';
    let md = page.data.markdown;
    const fm = page.data.frontmatter ?? {};
    if (fm.unit_id) {
      // Удаляем первый `# ...` (он совпадает с headerNode).
      md = md.replace(/^#\s+.+\n+/, '');
    }
    return renderMarkdown(md, resolveUnit);
  }, [page.data, resolveUnit]);

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
              className={`wiki-cat-btn${activeCat === c.key ? ' active' : ''}`}
            >
              {c.title}
            </button>
            {activeCat === c.key && pages.data && (
              <div style={{ marginLeft: 12, marginTop: 4, display: 'flex', flexDirection: 'column', gap: 2 }}>
                {pages.data.pages.map((p) => {
                  const slug = p.path.split('/')[1] ?? 'index';
                  const fm = p.frontmatter ?? {};
                  const unitID = fm.unit_id ? Number(fm.unit_id) : 0;
                  const title = unitID > 0 ? nameOf(unitID) : (fm.title ?? slug);
                  const img = unitID > 0 ? imageOfId(unitID) : '';
                  return (
                    <button
                      key={slug}
                      onClick={() => setActiveSlug(slug)}
                      className={`wiki-page-btn${activeSlug === slug ? ' active' : ''}`}
                    >
                      {img && <img src={img} alt="" className="wiki-page-icon" />}
                      <span>{title}</span>
                    </button>
                  );
                })}
              </div>
            )}
          </div>
        ))}
      </aside>

      {/* Content */}
      <main
        ref={contentRef}
        className="ox-panel"
        style={{ flex: 1, overflowY: 'auto', padding: 16 }}
      >
        {page.isLoading && <div>Загрузка статьи…</div>}
        {page.isError && <div>Не удалось загрузить статью.</div>}
        {page.data && (
          <article className="wiki-article" style={{ lineHeight: 1.6 }}>
            {headerNode}
            <div
              // Контент из docs/wiki/, контролируется командой проекта.
              // XSS не критичен, но html-инъекций мы избегаем через escapeHtml в md.ts.
              dangerouslySetInnerHTML={{ __html: bodyHtml }}
            />
          </article>
        )}
      </main>
    </div>
  );
}
