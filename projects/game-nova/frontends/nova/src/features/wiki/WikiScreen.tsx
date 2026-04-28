import { useQuery } from '@tanstack/react-query';
import { useEffect, useMemo, useRef, useState } from 'react';
import { api } from '@/api/client';
import { categoryOfId, imageOfId, keyOfId, nameOf } from '@/api/catalog';
import { renderMarkdown } from './md';
import { useTranslation } from '@/i18n/i18n';
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
// parseWikiHash: «#wiki/buildings/exchange» → { cat: 'buildings', slug: 'exchange' }.
// Возвращает пустые строки, если хеш не относится к wiki.
function parseWikiHash(): { cat: string; slug: string } {
  const hash = window.location.hash.replace(/^#/, '');
  const parts = hash.split('/');
  if (parts[0] !== 'wiki') return { cat: '', slug: '' };
  return { cat: parts[1] ?? '', slug: parts[2] ?? 'index' };
}

export function WikiScreen() {
  const { t } = useTranslation('wiki');
  const { t: ti } = useTranslation('info');
  const initial = parseWikiHash();
  const [activeCat, setActiveCat] = useState<string>(initial.cat);
  const [activeSlug, setActiveSlug] = useState<string>(initial.slug || 'index');
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

  // navigate: меняем активную страницу и записываем в history, чтобы
  // браузерные «назад/вперёд» возвращали к предыдущим статьям вики.
  const navigate = (cat: string, slug: string) => {
    setActiveCat(cat);
    setActiveSlug(slug);
    const url = `#wiki/${cat}/${slug || 'index'}`;
    if (window.location.hash !== url) {
      history.pushState(null, '', url);
    }
  };

  // Синхронизируем url с авто-выбранной категорией (без push в историю).
  useEffect(() => {
    if (activeCat && !parseWikiHash().cat) {
      history.replaceState(null, '', `#wiki/${activeCat}/${activeSlug || 'index'}`);
    }
  }, [activeCat, activeSlug]);

  // popstate: пользователь нажал «назад» — восстанавливаем страницу из url.
  useEffect(() => {
    const onPop = () => {
      const parsed = parseWikiHash();
      if (!parsed.cat) return; // вышли за пределы wiki — пусть App.tsx разруливает.
      setActiveCat(parsed.cat);
      setActiveSlug(parsed.slug || 'index');
    };
    window.addEventListener('popstate', onPop);
    return () => window.removeEventListener('popstate', onPop);
  }, []);

  // resolveUnit: id → { name, image } для md-рендерера.
  const resolveUnit = useMemo(
    () => (id: number) => {
      const name = nameOf(id, ti);
      const image = imageOfId(id);
      return { name, image };
    },
    [ti]
  );

  // Делегированный click на ссылках [[unit:N]].
  // При клике — узнаём из id wiki-категорию и slug, переключаем активную страницу.
  useEffect(() => {
    const root = contentRef.current;
    if (!root) return;
    const onClick = (e: MouseEvent) => {
      let el = e.target as HTMLElement | null;
      while (el && !el.classList?.contains('wiki-unit-link') && !el.classList?.contains('wiki-page-link')) {
        el = el.parentElement;
      }
      if (!el) return;
      // [[wiki:cat/slug]] — прямая навигация по категории и slug.
      if (el.classList.contains('wiki-page-link')) {
        e.preventDefault();
        const cat = el.getAttribute('data-wiki-cat') ?? '';
        const slug = el.getAttribute('data-wiki-slug') ?? 'index';
        if (!cat) return;
        navigate(cat, slug);
        return;
      }
      // [[unit:N]] — ищем категорию и slug по числовому id юнита.
      e.preventDefault();
      const idStr = el.getAttribute('data-unit-id');
      if (!idStr) return;
      const id = Number(idStr);
      const cat = categoryOfId(id);
      const slug = keyOfId(id);
      if (!cat || !slug) return;
      navigate(cat, slug);
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
      const name = nameOf(unitID, ti);
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
    // Если headerNode уже отрисует заголовок (по unit_id или title из
    // frontmatter), убираем первый `# ...` из markdown, иначе он дублирует.
    if (fm.unit_id || fm.title) {
      md = md.replace(/^\s*#\s+.+\n+/, '');
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
        <h3 style={{ margin: '4px 8px', fontSize: 14 }}>{t('sidebarTitle')}</h3>
        {cats.isLoading && <div style={{ padding: 8 }}>{t('loading')}</div>}
        {(cats.data?.categories ?? []).map((c) => (
          <div key={c.key} style={{ marginBottom: 4 }}>
            <button
              onClick={() => navigate(c.key, 'index')}
              className={`wiki-cat-btn${activeCat === c.key ? ' active' : ''}`}
            >
              {c.title}
            </button>
            {activeCat === c.key && pages.data && (() => {
              // Скрываем список если единственная страница — index с тем же именем что категория.
              const pageList = pages.data.pages.filter((p) => {
                const slug = p.path.split('/')[1] ?? 'index';
                if (slug !== 'index') return true;
                const fm = p.frontmatter ?? {};
                const unitID = fm.unit_id ? Number(fm.unit_id) : 0;
                const title = unitID > 0 ? nameOf(unitID, ti) : (fm.title ?? slug);
                return title !== c.title;
              });
              if (pageList.length === 0) return null;
              return (
                <div style={{ marginLeft: 12, marginTop: 4, display: 'flex', flexDirection: 'column', gap: 2 }}>
                  {pageList.map((p) => {
                    const slug = p.path.split('/')[1] ?? 'index';
                    const fm = p.frontmatter ?? {};
                    const unitID = fm.unit_id ? Number(fm.unit_id) : 0;
                    const title = unitID > 0 ? nameOf(unitID, ti) : (fm.title ?? slug);
                    const img = unitID > 0 ? imageOfId(unitID) : '';
                    return (
                      <button
                        key={slug}
                        onClick={() => navigate(c.key, slug)}
                        className={`wiki-page-btn${activeSlug === slug ? ' active' : ''}`}
                      >
                        {img && <img src={img} alt="" className="wiki-page-icon" />}
                        <span>{title}</span>
                      </button>
                    );
                  })}
                </div>
              );
            })()}
          </div>
        ))}
      </aside>

      {/* Content */}
      <main
        ref={contentRef}
        className="ox-panel"
        style={{ flex: 1, overflowY: 'auto', padding: '20px 28px' }}
      >
        {page.isLoading && <div>{t('articleLoading')}</div>}
        {page.isError && <div>{t('articleError')}</div>}
        {page.data && (
          <article className="wiki-article">
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
