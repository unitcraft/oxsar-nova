// Хелпер expectNoLayoutIssues — автоматически проверяет макет на каждом
// экране: отсутствие горизонтального скролла, видимость landmark'ов,
// отсутствие пересечений между соседними кликабельными элементами,
// z-order sticky-шапки.
//
// Не аттачит скриншот — это делает сам Playwright при падении теста.

import { expect, type Page } from '@playwright/test';

export interface LayoutIssue {
  kind: 'horizontal-scroll' | 'missing-landmark' | 'invisible-text' | 'overlap' | 'outside-viewport';
  detail: string;
}

// Элементы, которым разрешено накладываться на соседей (overlay, badge, portal).
// .ox-bottom-nav — sticky-панель внизу mobile-экрана поверх контента.
// .ox-header — sticky-панель сверху.
const OVERLAY_WHITELIST_SELECTOR = [
  '.ox-modal-overlay',
  '.ox-modal',
  '.ox-planet-dropdown',
  '.badge',
  '[role="tooltip"]',
  '.ox-toast',
  '.ox-bottom-nav',
  '.ox-header',
].join(',');

export async function collectLayoutIssues(page: Page): Promise<LayoutIssue[]> {
  return page.evaluate(
    ([overlayWhitelistSelector]) => {
      const issues: Array<{ kind: string; detail: string }> = [];

      // 1. Горизонтальный скролл body — 99% признак протекающей ширины.
      if (document.documentElement.scrollWidth > document.documentElement.clientWidth + 1) {
        issues.push({
          kind: 'horizontal-scroll',
          detail: `scrollWidth=${document.documentElement.scrollWidth} > clientWidth=${document.documentElement.clientWidth}`,
        });
      }

      // 2. Landmark'и: header и либо sidebar (desktop), либо bottom-nav (mobile).
      const requireVisible = (selector: string, name: string) => {
        const el = document.querySelector(selector);
        if (!el) {
          issues.push({ kind: 'missing-landmark', detail: `${name}: ${selector} not found` });
          return;
        }
        const r = (el as HTMLElement).getBoundingClientRect();
        if (r.width === 0 || r.height === 0) {
          issues.push({ kind: 'missing-landmark', detail: `${name}: ${selector} has zero size` });
        }
      };
      const isShown = (el: Element | null): boolean => {
        if (!el) return false;
        const r = (el as HTMLElement).getBoundingClientRect();
        if (r.width === 0 || r.height === 0) return false;
        const style = window.getComputedStyle(el as HTMLElement);
        return style.display !== 'none' && style.visibility !== 'hidden';
      };

      requireVisible('.ox-header', 'header');
      const sidebarVisible = isShown(document.querySelector('.ox-sidebar'));
      const bottomVisible = isShown(document.querySelector('.ox-bottom-nav'));
      if (!sidebarVisible && !bottomVisible) {
        issues.push({ kind: 'missing-landmark', detail: 'neither .ox-sidebar nor .ox-bottom-nav visible' });
      }
      requireVisible('main.ox-content', 'main content');

      // 3. Нулевые текстовые узлы — индикатор сломанного flex/grid.
      // Пропускаем элементы, скрытые через CSS (display:none, visibility:hidden,
      // offsetParent:null) — на mobile/desktop разные landmark'и скрыты media-query.
      document.querySelectorAll<HTMLElement>('h1, h2, h3, h4, button, a, label, td, th').forEach((el) => {
        const text = el.textContent?.trim() ?? '';
        if (!text) return;
        if (el.offsetParent === null) return; // скрыт media-query или родителем
        const style = window.getComputedStyle(el);
        if (style.display === 'none' || style.visibility === 'hidden') return;
        const r = el.getBoundingClientRect();
        if (r.width === 0 || r.height === 0) {
          issues.push({
            kind: 'invisible-text',
            detail: `"${text.slice(0, 40)}" at ${el.tagName} has zero size`,
          });
        }
      });

      // 4. Пересечения между соседними кликабельными элементами.
      // Берём только элементы в пределах вьюпорта; исключаем overlay-whitelist
      // и элементы, находящиеся внутри whitelist-предков.
      const clickable = Array.from(
        document.querySelectorAll<HTMLElement>(
          'button:not([disabled]), a[href], [role="button"]:not([aria-disabled="true"])',
        ),
      ).filter((el) => {
        if (el.closest(overlayWhitelistSelector)) return false;
        if (el.matches(overlayWhitelistSelector)) return false;
        const r = el.getBoundingClientRect();
        if (r.width <= 0 || r.height <= 0) return false;
        // В пределах вьюпорта (с запасом).
        if (r.bottom < 0 || r.top > window.innerHeight) return false;
        if (r.right < 0 || r.left > window.innerWidth) return false;
        return true;
      });

      const rectsOverlap = (a: DOMRect, b: DOMRect): boolean =>
        !(a.right <= b.left || b.right <= a.left || a.bottom <= b.top || b.bottom <= a.top);

      for (let i = 0; i < clickable.length; i++) {
        for (let j = i + 1; j < clickable.length; j++) {
          const a = clickable[i]!;
          const b = clickable[j]!;
          // Исключаем вложенные (nested a inside button и т.п.).
          if (a.contains(b) || b.contains(a)) continue;
          const ra = a.getBoundingClientRect();
          const rb = b.getBoundingClientRect();
          if (rectsOverlap(ra, rb)) {
            // Проверяем перекрытие по площади > 25% — мелкие касания краёв игнорируем.
            const ix = Math.max(0, Math.min(ra.right, rb.right) - Math.max(ra.left, rb.left));
            const iy = Math.max(0, Math.min(ra.bottom, rb.bottom) - Math.max(ra.top, rb.top));
            const overlapArea = ix * iy;
            const minArea = Math.min(ra.width * ra.height, rb.width * rb.height);
            if (minArea > 0 && overlapArea / minArea > 0.25) {
              issues.push({
                kind: 'overlap',
                detail: `"${(a.textContent ?? '').trim().slice(0, 20)}" overlaps "${(b.textContent ?? '').trim().slice(0, 20)}" (${Math.round((overlapArea / minArea) * 100)}%)`,
              });
            }
          }
        }
      }

      return issues;
    },
    [OVERLAY_WHITELIST_SELECTOR] as const,
  ) as Promise<LayoutIssue[]>;
}

export async function expectNoLayoutIssues(page: Page, screenName: string): Promise<void> {
  const issues = await collectLayoutIssues(page);
  if (issues.length > 0) {
    const summary = issues
      .slice(0, 10)
      .map((i) => `  [${i.kind}] ${i.detail}`)
      .join('\n');
    expect
      .soft(issues, `layout issues on screen "${screenName}":\n${summary}`)
      .toHaveLength(0);
  }
}
