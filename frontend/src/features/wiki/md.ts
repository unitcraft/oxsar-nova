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

// Минимальная подсветка псевдокода — токенизатор без regex по HTML.
// Принимает сырую строку, возвращает safe HTML.
function highlightCode(raw: string): string {
  const commentIdx = raw.indexOf('//');
  const code = commentIdx >= 0 ? raw.slice(0, commentIdx) : raw;
  const comment = commentIdx >= 0 ? raw.slice(commentIdx) : '';

  // Токенизируем код по токенам: идентификатор, число, оператор, прочее.
  const tokenRe = /([A-Za-zА-Яа-яЁё_][\wА-Яа-яЁё]*)(\s*\()?|(\d+(?:\.\d+)?%?)|([()[\]])|([=+\-*/×÷≤≥≠])|([^A-Za-zА-Яа-яЁё_\d=+\-*/×÷≤≥≠()[\]]+)/g;
  let out = '';
  let m: RegExpExecArray | null;
  while ((m = tokenRe.exec(code)) !== null) {
    if (m[1] !== undefined) {
      // Идентификатор — функция если за ним «(», иначе обычный текст.
      const name = escapeHtml(m[1]);
      const paren = m[2] ?? '';
      if (paren.trimStart().startsWith('(')) {
        out += `<span class="ch-fn">${name}</span>${escapeHtml(paren)}`;
      } else {
        out += name + escapeHtml(paren);
      }
    } else if (m[3] !== undefined) {
      out += `<span class="ch-num">${escapeHtml(m[3])}</span>`;
    } else if (m[4] !== undefined) {
      out += `<span class="ch-paren">${escapeHtml(m[4])}</span>`;
    } else if (m[5] !== undefined) {
      out += `<span class="ch-op">${escapeHtml(m[5])}</span>`;
    } else {
      out += escapeHtml(m[6] ?? m[0] ?? '');
    }
  }

  if (comment) {
    out += `<span class="ch-comment">${escapeHtml(comment)}</span>`;
  }
  return out;
}

// resolveUnit — маппит unit_id в { name, image }. Передаётся снаружи
// (зависит от каталога, который во frontend живёт отдельным модулем).
export type UnitResolver = (id: number) => { name: string; image: string } | null;

function inline(s: string, resolveUnit?: UnitResolver): string {
  // Сначала экранируем, потом восстанавливаем разметку.
  let out = escapeHtml(s);
  // Code (без вложенности).
  out = out.replace(/`([^`]+)`/g, '<code>$1</code>');
  // Bold.
  out = out.replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>');
  // Italic.
  out = out.replace(/(^|[^*])\*([^*]+)\*([^*]|$)/g, '$1<em>$2</em>$3');
  // [[unit:N]] — ссылка на страницу юнита с иконкой + именем.
  // Должно идти ДО общих [text](url), потому что после escape остаётся
  // как `[[unit:42]]`.
  out = out.replace(/\[\[unit:(\d+)\]\]/g, (_m, idStr: string) => {
    const id = parseInt(idStr, 10);
    if (resolveUnit) {
      const u = resolveUnit(id);
      if (u) {
        const img = u.image
          ? `<img src="${u.image}" alt="" class="wiki-unit-icon"/>`
          : '';
        return `<a class="wiki-unit-link" data-unit-id="${id}">${img}<span class="wiki-unit-name">${escapeHtml(u.name)}</span></a>`;
      }
    }
    return `<a class="wiki-unit-link" data-unit-id="${id}">unit ${id}</a>`;
  });
  // [[wiki:category/slug|подпись]] — ссылка на рукописную страницу
  // вики (например, [[wiki:resources/index]] или [[wiki:combat/index|боя]]).
  out = out.replace(/\[\[wiki:([a-z0-9_-]+)\/([a-z0-9_-]+)(?:\|([^\]]+))?\]\]/g,
    (_m, cat: string, slug: string, label: string | undefined) => {
      const text = escapeHtml(label ?? slug);
      return `<a class="wiki-page-link" data-wiki-cat="${cat}" data-wiki-slug="${slug}">${text}</a>`;
    });
  // Links [text](url).
  out = out.replace(/\[([^\]]+)\]\(([^)]+)\)/g, (_m, text: string, url: string) => {
    const safeUrl = url.replace(/"/g, '%22');
    return `<a href="${safeUrl}">${text}</a>`;
  });
  return out;
}

export function renderMarkdown(md: string, resolveUnit?: UnitResolver): string {
  if (!md) return '';
  const lines = md.split(/\r?\n/);
  const html: string[] = [];
  let i = 0;

  const flushParagraph = (buf: string[]) => {
    if (buf.length === 0) return;
    html.push(`<p>${inline(buf.join(' '), resolveUnit)}</p>`);
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

    // Fenced code block: ```...```
    // Внутри блока [[unit:N]] и markdown НЕ обрабатываются — это код.
    if (/^```/.test(line)) {
      flushParagraph(para);
      i++; // пропускаем открывающую ```
      const codeLines: string[] = [];
      while (i < lines.length && !/^```/.test(lines[i] ?? '')) {
        codeLines.push(lines[i] ?? '');
        i++;
      }
      i++; // пропускаем закрывающую ```
      const escaped = codeLines.map((l) => highlightCode(l)).join('\n');
      html.push(`<pre><code>${escaped}</code></pre>`);
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
      html.push(`<h${lvl}>${inline(h[2] ?? '', resolveUnit)}</h${lvl}>`);
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
      for (const h of headers) t += `<th>${inline(h, resolveUnit)}</th>`;
      t += '</tr></thead><tbody>';
      for (const r of rows) {
        t += '<tr>';
        for (const c of r) t += `<td>${inline(c, resolveUnit)}</td>`;
        t += '</tr>';
      }
      t += '</tbody></table>';
      html.push(t);
      continue;
    }

    // Список — нумерованный или маркированный. Поддерживает
    // продолжение строк с отступом (CommonMark «list item continuation»),
    // чтобы перенос строки внутри пункта не разрывал нумерацию.
    const ul = /^\s*[-*]\s+(.+)$/.exec(line);
    const ol = /^\s*\d+\.\s+(.+)$/.exec(line);
    if (ul || ol) {
      flushParagraph(para);
      const tag = ul ? 'ul' : 'ol';
      const items: string[] = [];
      let buf: string[] = [];
      const pushItem = () => {
        if (buf.length === 0) return;
        items.push(`<li>${inline(buf.join(' '), resolveUnit)}</li>`);
        buf = [];
      };
      while (i < lines.length) {
        const cur = lines[i] ?? '';
        const u = /^\s*[-*]\s+(.+)$/.exec(cur);
        const o = /^\s*\d+\.\s+(.+)$/.exec(cur);
        if (u || o) {
          pushItem();
          const match = (u ?? o)!;
          buf.push(match[1] ?? '');
          i++;
          continue;
        }
        // Продолжение пункта: отступ ≥ 2 пробел/табуляция и непустая строка.
        if (buf.length > 0 && /^\s+\S/.test(cur)) {
          buf.push(cur.trim());
          i++;
          continue;
        }
        break;
      }
      pushItem();
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
      html.push(`<blockquote>${inline(q.join(' '), resolveUnit)}</blockquote>`);
      continue;
    }

    // Иначе — параграф.
    para.push(line);
    i++;
  }
  flushParagraph(para);
  return html.join('\n');
}
