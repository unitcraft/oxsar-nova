// Минимальный markdown→HTML рендерер для wiki (план 19 frontend MVP).
// Поддерживает: # заголовки, **bold**, *italic*, `code`, [text](url),
// списки -, нумерованные 1., таблицы | a | b |, blockquote >, hr ---.
//
// Не используем react-markdown/marked, чтобы не тащить зависимость.
// Если потребуется больше (mdx, плагины) — заменим целиком.

function escapeHtml(s: string): string {
  return s
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;');
}

function inline(s: string): string {
  // Сначала экранируем, потом восстанавливаем разметку.
  let out = escapeHtml(s);
  // Code (без вложенности).
  out = out.replace(/`([^`]+)`/g, '<code>$1</code>');
  // Bold.
  out = out.replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>');
  // Italic.
  out = out.replace(/(^|[^*])\*([^*]+)\*([^*]|$)/g, '$1<em>$2</em>$3');
  // Links [text](url).
  out = out.replace(/\[([^\]]+)\]\(([^)]+)\)/g, (_m, text: string, url: string) => {
    const safeUrl = url.replace(/"/g, '%22');
    return `<a href="${safeUrl}">${text}</a>`;
  });
  return out;
}

export function renderMarkdown(md: string): string {
  if (!md) return '';
  const lines = md.split(/\r?\n/);
  const html: string[] = [];
  let i = 0;

  const flushParagraph = (buf: string[]) => {
    if (buf.length === 0) return;
    html.push(`<p>${inline(buf.join(' '))}</p>`);
    buf.length = 0;
  };

  let para: string[] = [];

  while (i < lines.length) {
    const line = lines[i] ?? '';

    // Пустая.
    if (/^\s*$/.test(line)) {
      flushParagraph(para);
      i++;
      continue;
    }

    // hr.
    if (/^---+\s*$/.test(line)) {
      flushParagraph(para);
      html.push('<hr/>');
      i++;
      continue;
    }

    // Заголовки.
    const h = /^(#{1,6})\s+(.+)$/.exec(line);
    if (h) {
      flushParagraph(para);
      const lvl = (h[1] ?? '').length;
      html.push(`<h${lvl}>${inline(h[2] ?? '')}</h${lvl}>`);
      i++;
      continue;
    }

    // Таблица: строка содержит |…|, следующая — разделитель |---|---|.
    const nextLine = lines[i + 1] ?? '';
    if (/^\|.+\|\s*$/.test(line) && i + 1 < lines.length && /^\|[-:\s|]+\|\s*$/.test(nextLine)) {
      flushParagraph(para);
      const headers = line.split('|').slice(1, -1).map((c) => c.trim());
      i += 2; // пропускаем разделитель
      const rows: string[][] = [];
      let curRow = lines[i] ?? '';
      while (i < lines.length && /^\|.+\|\s*$/.test(curRow)) {
        rows.push(curRow.split('|').slice(1, -1).map((c) => c.trim()));
        i++;
        curRow = lines[i] ?? '';
      }
      let t = '<table><thead><tr>';
      for (const h of headers) t += `<th>${inline(h)}</th>`;
      t += '</tr></thead><tbody>';
      for (const r of rows) {
        t += '<tr>';
        for (const c of r) t += `<td>${inline(c)}</td>`;
        t += '</tr>';
      }
      t += '</tbody></table>';
      html.push(t);
      continue;
    }

    // Список — нумерованный или маркированный.
    const ul = /^\s*[-*]\s+(.+)$/.exec(line);
    const ol = /^\s*\d+\.\s+(.+)$/.exec(line);
    if (ul || ol) {
      flushParagraph(para);
      const tag = ul ? 'ul' : 'ol';
      const items: string[] = [];
      while (i < lines.length) {
        const cur = lines[i] ?? '';
        const u = /^\s*[-*]\s+(.+)$/.exec(cur);
        const o = /^\s*\d+\.\s+(.+)$/.exec(cur);
        if (!u && !o) break;
        const match = (u ?? o)!;
        items.push(`<li>${inline(match[1] ?? '')}</li>`);
        i++;
      }
      html.push(`<${tag}>${items.join('')}</${tag}>`);
      continue;
    }

    // Blockquote.
    if (/^\s*>\s+/.test(line)) {
      flushParagraph(para);
      const q: string[] = [];
      while (i < lines.length) {
        const cur = lines[i] ?? '';
        if (!/^\s*>\s+/.test(cur)) break;
        q.push(cur.replace(/^\s*>\s+/, ''));
        i++;
      }
      html.push(`<blockquote>${inline(q.join(' '))}</blockquote>`);
      continue;
    }

    // Иначе — параграф.
    para.push(line);
    i++;
  }
  flushParagraph(para);
  return html.join('\n');
}
