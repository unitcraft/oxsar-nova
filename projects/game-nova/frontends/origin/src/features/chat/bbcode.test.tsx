// План 72.1.19: тесты BBCode-парсера.
//
// Используем react-renderer для inspect (vitest + react-dom/server
// уже подключены как dev-deps origin фронта).

import { describe, expect, it } from 'vitest';
import { renderToStaticMarkup } from 'react-dom/server';
import { renderBBCode } from './bbcode';

function render(input: string): string {
  return renderToStaticMarkup(<>{renderBBCode(input)}</>);
}

describe('renderBBCode', () => {
  it('plain text passes through', () => {
    expect(render('hello world')).toBe('hello world');
  });

  it('[b] → <b>', () => {
    expect(render('text [b]bold[/b] more')).toBe('text <b>bold</b> more');
  });

  it('[i][u][s] mapped to <i><u><s>', () => {
    expect(render('[i]a[/i] [u]b[/u] [s]c[/s]')).toBe(
      '<i>a</i> <u>b</u> <s>c</s>',
    );
  });

  it('[color=red]...[/color] → <span style="color:red">', () => {
    expect(render('[color=red]hi[/color]')).toBe(
      '<span style="color:red">hi</span>',
    );
  });

  it('rejects unsafe color (script-like)', () => {
    const out = render('[color=red;background:url(x)]hi[/color]');
    // Парсер не нашёл валидный color → оригинал как текст
    expect(out).toContain('[color=');
    expect(out).not.toContain('<span style');
  });

  it('[url]http://x[/url] → safe <a>', () => {
    const out = render('[url]http://example.com[/url]');
    expect(out).toContain('href="http://example.com"');
    expect(out).toContain('rel="noopener noreferrer"');
    expect(out).toContain('target="_blank"');
  });

  it('rejects javascript: URL', () => {
    const out = render('[url]javascript:alert(1)[/url]');
    expect(out).not.toContain('href="javascript:');
    expect(out).toContain('[url]');
  });

  it('[url=http://x]label[/url] uses label', () => {
    const out = render('[url=http://example.com]click[/url]');
    expect(out).toContain('href="http://example.com"');
    expect(out).toContain('>click</a>');
  });

  it('emoji [:smile:] → <img>', () => {
    const out = render('hi [:smile:]');
    expect(out).toContain('<img');
    expect(out).toContain('src="/assets/origin/emo/smile.gif"');
  });

  it('unbalanced [b] without close stays as text', () => {
    expect(render('text [b] no close')).toBe('text [b] no close');
  });

  it('nested unsupported — outer parsed, inner literal', () => {
    // Парсер не делает рекурсию для других теговых выражений,
    // только для [:emoji:]. [b][i]x[/i][/b] — нужна иерархия.
    // Текущая реализация: первый regex.exec возьмёт самый внешний
    // matched тег только если inner закрыт нашим же tag. Тестируем
    // что хотя бы [b][:smile:][/b] работает.
    const out = render('[b][:smile:][/b]');
    expect(out).toContain('<b>');
    expect(out).toContain('<img');
  });
});
