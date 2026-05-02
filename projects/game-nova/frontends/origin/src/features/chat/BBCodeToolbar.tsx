// План 72.1.55 Task F (P72.S4.BBCODE 1:1): toolbar для вставки BBCode
// тегов в чат-поле. Дополняет renderBBCode (он на read-side) — теперь
// пользователь может ввести [b]/[i]/[u]/[s]/[color]/[url]/[img]/смайл
// через кнопки, а не руками.
//
// Legacy `chat.tpl` использует jQuery-плагин «BBCode Toolbar»;
// origin-фронт реализует тот же набор операций нативной textarea
// selectionStart/End API (без TipTap, чтобы не тянуть зависимости).

import { useTranslation } from '@/i18n/i18n';

// Список emoji, доступных через picker. Должен совпадать с файлами в
// /assets/origin/emo/<name>.gif.
const EMOJIS = [
  'smile', 'sad', 'wink', 'big-grin', 'tongue', 'cool',
  'shocked', 'angry', 'love', 'sleepy', 'evil', 'angel',
] as const;

export interface BBCodeToolbarProps {
  // ref на textarea для вставки тегов вокруг selection.
  textareaRef: React.RefObject<HTMLTextAreaElement | null>;
  // Колбек вызывается после каждой вставки — фронт может обновить
  // controlled-state (если textarea controlled).
  onChange: (newValue: string) => void;
  disabled?: boolean;
}

// wrapSelection вставляет открывающий + закрывающий теги вокруг
// текущего selection в textarea. Если selection пуст — вставляет
// `[tag][/tag]` в позицию курсора и оставляет курсор внутри.
function wrapSelection(
  ta: HTMLTextAreaElement,
  open: string,
  close: string,
): string {
  const value = ta.value;
  const start = ta.selectionStart;
  const end = ta.selectionEnd;
  const before = value.slice(0, start);
  const inside = value.slice(start, end);
  const after = value.slice(end);
  const newValue = before + open + inside + close + after;
  // Восстановим selection после React re-render (через requestAnimationFrame).
  requestAnimationFrame(() => {
    ta.focus();
    const cursorAt = start + open.length + inside.length;
    ta.setSelectionRange(cursorAt, cursorAt);
  });
  return newValue;
}

// insertAtCursor вставляет text в позицию курсора (без selection-замены).
function insertAtCursor(ta: HTMLTextAreaElement, text: string): string {
  const value = ta.value;
  const start = ta.selectionStart;
  const after = value.slice(start);
  const newValue = value.slice(0, start) + text + after;
  requestAnimationFrame(() => {
    ta.focus();
    const cursorAt = start + text.length;
    ta.setSelectionRange(cursorAt, cursorAt);
  });
  return newValue;
}

export function BBCodeToolbar({
  textareaRef,
  onChange,
  disabled,
}: BBCodeToolbarProps): React.ReactElement {
  const { t } = useTranslation();

  function wrap(open: string, close: string) {
    const ta = textareaRef.current;
    if (!ta) return;
    onChange(wrapSelection(ta, open, close));
  }

  function insertEmoji(name: string) {
    const ta = textareaRef.current;
    if (!ta) return;
    onChange(insertAtCursor(ta, `[:${name}:]`));
  }

  function insertUrl() {
    const ta = textareaRef.current;
    if (!ta) return;
    const url = window.prompt(t('chat', 'urlPrompt') || 'URL:', 'https://');
    if (!url) return;
    // Если есть selection — сделать [url=..]label[/url]; иначе [url]url[/url].
    const inside = ta.value.slice(ta.selectionStart, ta.selectionEnd);
    if (inside) {
      onChange(wrapSelection(ta, `[url=${url}]`, '[/url]'));
    } else {
      onChange(insertAtCursor(ta, `[url]${url}[/url]`));
    }
  }

  function insertImg() {
    const ta = textareaRef.current;
    if (!ta) return;
    const url = window.prompt(t('chat', 'imgPrompt') || 'Image URL:', 'https://');
    if (!url) return;
    onChange(insertAtCursor(ta, `[img]${url}[/img]`));
  }

  function insertColor() {
    const ta = textareaRef.current;
    if (!ta) return;
    const color = window.prompt(t('chat', 'colorPrompt') || 'Color (red, #f0a, #ff00aa):', 'red');
    if (!color) return;
    onChange(wrapSelection(ta, `[color=${color}]`, '[/color]'));
  }

  return (
    <div
      style={{
        display: 'flex',
        flexWrap: 'wrap',
        gap: 4,
        marginBottom: 4,
        padding: 4,
      }}
    >
      <button
        type="button"
        className="button"
        disabled={disabled}
        title={t('chat', 'bbBold') || 'Жирный'}
        onClick={() => wrap('[b]', '[/b]')}
        style={{ fontWeight: 'bold' }}
      >
        B
      </button>
      <button
        type="button"
        className="button"
        disabled={disabled}
        title={t('chat', 'bbItalic') || 'Курсив'}
        onClick={() => wrap('[i]', '[/i]')}
        style={{ fontStyle: 'italic' }}
      >
        I
      </button>
      <button
        type="button"
        className="button"
        disabled={disabled}
        title={t('chat', 'bbUnderline') || 'Подчёркнутый'}
        onClick={() => wrap('[u]', '[/u]')}
        style={{ textDecoration: 'underline' }}
      >
        U
      </button>
      <button
        type="button"
        className="button"
        disabled={disabled}
        title={t('chat', 'bbStrike') || 'Зачёркнутый'}
        onClick={() => wrap('[s]', '[/s]')}
        style={{ textDecoration: 'line-through' }}
      >
        S
      </button>
      <button
        type="button"
        className="button"
        disabled={disabled}
        title={t('chat', 'bbColor') || 'Цвет'}
        onClick={insertColor}
      >
        🎨
      </button>
      <button
        type="button"
        className="button"
        disabled={disabled}
        title={t('chat', 'bbUrl') || 'Ссылка'}
        onClick={insertUrl}
      >
        🔗
      </button>
      <button
        type="button"
        className="button"
        disabled={disabled}
        title={t('chat', 'bbImg') || 'Изображение'}
        onClick={insertImg}
      >
        🖼
      </button>
      <span
        style={{
          display: 'inline-flex',
          gap: 2,
          marginLeft: 8,
          alignItems: 'center',
        }}
      >
        {EMOJIS.map((name) => (
          <button
            key={name}
            type="button"
            className="button"
            disabled={disabled}
            title={name}
            onClick={() => insertEmoji(name)}
            style={{ padding: 2 }}
          >
            <img
              src={`/assets/origin/emo/${name}.gif`}
              alt={`:${name}:`}
              width={20}
              height={20}
              onError={(e) => {
                // Если smile.gif не задеплоен — скрыть кнопку.
                (e.currentTarget.parentElement as HTMLElement).style.display = 'none';
              }}
            />
          </button>
        ))}
      </span>
    </div>
  );
}
