// План 72.1.19 — компактный BBCode-парсер для чата (legacy паритет с
// `Chat.class.php::bbcode`).
//
// Поддерживаемые теги (в порядке legacy regex):
//   [b]...[/b]    [i]...[/i]    [u]...[/u]    [s]...[/s]
//   [color=red]...[/color]
//   [:emojiName:]                   → <img src=/assets/origin/emo/<name>.gif>
//   [img]http://...[/img]           → <a target=_blank>http://...</a>
//   [url]http://...[/url]           → <a>http://...</a>
//   [url=http://...]label[/url]     → <a>label</a>
//
// SAFE: только whitelist-теги, без dangerouslySetInnerHTML.
//
// Парсер однопроходный: ищет первый matched тег, рекурсивно рендерит
// inner и продолжает с остатка. Это медленно для длинных сообщений, но
// чат-сообщения короткие (≤500 символов), так что O(n × tags) ок.

import type { ReactNode } from 'react';
import { Fragment } from 'react';

// safeColor — проверка, что цвет — это `red`/`#abc`/`#aabbcc`.
function safeColor(c: string): string | null {
  if (/^[a-zA-Z]+$/.test(c) && c.length <= 24) return c;
  if (/^#[0-9a-fA-F]{3}$/.test(c)) return c;
  if (/^#[0-9a-fA-F]{6}$/.test(c)) return c;
  return null;
}

// safeUrl — только http/https.
function safeUrl(u: string): string | null {
  const trimmed = u.trim();
  if (/^https?:\/\//i.test(trimmed)) return trimmed;
  return null;
}

// Кандидат-теги — каждый описан regex (с захватом inner / args) и
// функцией-рендером. Порядок важен: проверяем в order, берём с самой
// ранней позицией (по index).
interface TagCandidate {
  re: RegExp;
  render: (m: RegExpExecArray, key: string, parsedInner: ReactNode) => ReactNode;
  // Какой индекс группы содержит inner (для рекурсивного парсинга).
  innerGroup: number;
  // Параметры тега, которые надо проверить ДО рендера; если null —
  // тег невалиден и сегмент остаётся как plain text.
  validate?: (m: RegExpExecArray) => boolean;
}

const TAGS: TagCandidate[] = [
  // [b]...[/b]
  {
    re: /\[b\]([\s\S]*?)\[\/b\]/,
    innerGroup: 1,
    render: (_m, key, inner) => <b key={key}>{inner}</b>,
  },
  {
    re: /\[i\]([\s\S]*?)\[\/i\]/,
    innerGroup: 1,
    render: (_m, key, inner) => <i key={key}>{inner}</i>,
  },
  {
    re: /\[u\]([\s\S]*?)\[\/u\]/,
    innerGroup: 1,
    render: (_m, key, inner) => <u key={key}>{inner}</u>,
  },
  {
    re: /\[s\]([\s\S]*?)\[\/s\]/,
    innerGroup: 1,
    render: (_m, key, inner) => <s key={key}>{inner}</s>,
  },
  // [color=red]...[/color]
  {
    re: /\[color=([^\]]+)\]([\s\S]*?)\[\/color\]/,
    innerGroup: 2,
    validate: (m) => safeColor(m[1] ?? '') !== null,
    render: (m, key, inner) => (
      <span key={key} style={{ color: safeColor(m[1] ?? '') as string }}>
        {inner}
      </span>
    ),
  },
  // [url=http://x]label[/url]
  {
    re: /\[url=([^\]]+)\]([\s\S]*?)\[\/url\]/,
    innerGroup: 2,
    validate: (m) => safeUrl(m[1] ?? '') !== null,
    render: (m, key, inner) => (
      <a
        key={key}
        href={safeUrl(m[1] ?? '') as string}
        target="_blank"
        rel="noopener noreferrer"
        className="external"
      >
        {inner}
      </a>
    ),
  },
  // [url]http://x[/url]
  {
    re: /\[url\]([\s\S]*?)\[\/url\]/,
    innerGroup: 1,
    validate: (m) => safeUrl(m[1] ?? '') !== null,
    render: (m, key, inner) => (
      <a
        key={key}
        href={safeUrl(m[1] ?? '') as string}
        target="_blank"
        rel="noopener noreferrer"
        className="external"
      >
        {inner}
      </a>
    ),
  },
  // [img]http://x[/img] → ссылка (legacy refdir.php)
  {
    re: /\[img\]([\s\S]*?)\[\/img\]/,
    innerGroup: 1,
    validate: (m) => safeUrl(m[1] ?? '') !== null,
    render: (m, key) => {
      const href = safeUrl(m[1] ?? '') as string;
      return (
        <a
          key={key}
          href={href}
          target="_blank"
          rel="noopener noreferrer"
          className="external"
        >
          {href}
        </a>
      );
    },
  },
];

// Эмодзи [:name:] — отдельный regex; рендерится в text-only сегментах.
const EMOJI_RE = /\[:([a-zA-Z0-9_-]{1,32}):\]/g;

function renderEmojis(text: string, keyBase: string): ReactNode[] {
  const out: ReactNode[] = [];
  let lastEnd = 0;
  let m: RegExpExecArray | null;
  EMOJI_RE.lastIndex = 0;
  let i = 0;
  while ((m = EMOJI_RE.exec(text)) !== null) {
    if (m.index > lastEnd) {
      out.push(text.slice(lastEnd, m.index));
    }
    const name = m[1];
    if (name) {
      out.push(
        <img
          key={`${keyBase}-emo-${i}`}
          src={`/assets/origin/emo/${name}.gif`}
          alt={`:${name}:`}
          width={20}
          height={20}
          style={{ verticalAlign: 'middle' }}
        />,
      );
    }
    lastEnd = m.index + m[0].length;
    i++;
  }
  if (lastEnd < text.length) {
    out.push(text.slice(lastEnd));
  }
  return out.length === 0 ? [text] : out;
}

// findFirstTag — ищет тег с самой ранней позицией среди всех кандидатов.
function findFirstTag(input: string): { tag: TagCandidate; m: RegExpExecArray } | null {
  let best: { tag: TagCandidate; m: RegExpExecArray } | null = null;
  for (const tag of TAGS) {
    const re = new RegExp(tag.re.source, '');
    const m = re.exec(input);
    if (m === null) continue;
    if (tag.validate && !tag.validate(m)) continue;
    if (best === null || m.index < best.m.index) {
      best = { tag, m };
    }
  }
  return best;
}

// renderBBCode — рекурсивный рендер. Возвращает Fragment с массивом
// нод. keyPrefix используется для уникальных React-ключей при
// рекурсии.
export function renderBBCode(input: string, keyPrefix = 'r0'): ReactNode {
  if (!input) return null;
  const out: ReactNode[] = [];
  let rest = input;
  let i = 0;
  while (rest.length > 0) {
    const found = findFirstTag(rest);
    if (!found) {
      out.push(...renderEmojis(rest, `${keyPrefix}-tail-${i}`));
      break;
    }
    if (found.m.index > 0) {
      out.push(...renderEmojis(rest.slice(0, found.m.index), `${keyPrefix}-pre-${i}`));
    }
    const innerText = found.m[found.tag.innerGroup] ?? '';
    const parsedInner = renderBBCode(innerText, `${keyPrefix}-${i}-in`);
    out.push(found.tag.render(found.m, `${keyPrefix}-${i}`, parsedInner));
    rest = rest.slice(found.m.index + found.m[0].length);
    i++;
  }
  return <Fragment>{out}</Fragment>;
}
