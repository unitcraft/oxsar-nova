// X-009: helptip с расширенным содержимым при наведении.
// Origin (resource.tpl): класс `helptip` с onmouseover Tip()
// показывал описание + требования + бонусы. У nova в проекте на
// `<span title=...>` (нативный browser-tooltip) — мы оставляем его
// для простых случаев, а HelpTip даёт богатый контент (multi-line,
// форматированный) на hover/focus. Чистый CSS-подход без JS-либы.

import type { ReactNode } from 'react';

interface HelpTipProps {
  // children — то, что игрок видит и наводит курсор.
  children: ReactNode;
  // content — содержимое подсказки. Может быть текстом или JSX
  // (для требований, бонусов, формул).
  content: ReactNode;
  // placement — куда показывать. По умолчанию 'top'.
  placement?: 'top' | 'bottom';
}

export function HelpTip({ children, content, placement = 'top' }: HelpTipProps) {
  return (
    <span style={{ position: 'relative', display: 'inline-block' }} className="ox-helptip">
      <span style={{ borderBottom: '1px dotted var(--ox-fg-muted)', cursor: 'help' }}>
        {children}
      </span>
      <span
        role="tooltip"
        className="ox-helptip-content"
        style={{
          position: 'absolute',
          [placement]: '100%',
          left: '50%',
          transform: 'translateX(-50%)',
          marginTop: placement === 'bottom' ? 6 : 0,
          marginBottom: placement === 'top' ? 6 : 0,
          padding: '8px 12px',
          background: 'var(--ox-bg-panel-2, rgba(15,20,30,0.96))',
          color: 'var(--ox-fg)',
          border: '1px solid var(--ox-border)',
          borderRadius: 6,
          fontSize: 13,
          lineHeight: 1.4,
          whiteSpace: 'normal',
          minWidth: 200,
          maxWidth: 320,
          zIndex: 1000,
          opacity: 0,
          visibility: 'hidden',
          transition: 'opacity 120ms ease, visibility 120ms',
          pointerEvents: 'none',
          textAlign: 'left',
        }}
      >
        {content}
      </span>
    </span>
  );
}
