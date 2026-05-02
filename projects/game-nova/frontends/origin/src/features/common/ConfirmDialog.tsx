// In-game confirm dialog (план 72.1.53).
//
// Заменяет `window.confirm()` стилизованным модальным окном в духе
// игры. Раньше каждый action (recall, demolish, vacation, deletion,
// kick member, чат-msg delete и т.п.) использовал нативный
// browser-dialog — он визуально не вписывается в game-frame и в
// некоторых browsers выглядит как фишинг-warning.
//
// Используем HTML5 `<dialog>` element:
//   - Native focus-trap, Esc closes, backdrop click handled.
//   - Семантика modal — accessible by default.
//   - Нет дополнительных deps (без headless-ui / radix).
//
// Не используем portal/createRoot — `<dialog>` сам позиционируется
// поверх остального контента.
//
// API через React-state hook — паттерн как `useDialogConfirm()`:
//   const confirmDialog = useConfirmDialog();
//   ...
//   if (await confirmDialog.confirm({ title: 'Отозвать флот?', ... })) {
//     recall.mutate(id);
//   }
//
// Альтернативный простой API — компонент рендерится conditionally:
//   <ConfirmDialog open={...} title={...} onConfirm={...} onCancel={...} />

import { useEffect, useRef, useState, useCallback } from 'react';

export interface ConfirmDialogProps {
  open: boolean;
  title?: string;
  message: string;
  confirmLabel?: string;
  cancelLabel?: string;
  // Когда задан — confirm-кнопка стилизуется как «опасная» (красный).
  // Используется для destructive actions (delete, recall, abandon).
  destructive?: boolean;
  onConfirm: () => void;
  onCancel: () => void;
}

export function ConfirmDialog(props: ConfirmDialogProps): React.ReactElement | null {
  const ref = useRef<HTMLDialogElement | null>(null);

  // Открытие/закрытие через native API.
  useEffect(() => {
    const dlg = ref.current;
    if (!dlg) return;
    if (props.open && !dlg.open) {
      dlg.showModal();
    } else if (!props.open && dlg.open) {
      dlg.close();
    }
  }, [props.open]);

  // Закрытие через Esc или backdrop click. Native dialog кидает
  // событие 'close' на любой close.
  useEffect(() => {
    const dlg = ref.current;
    if (!dlg) return;
    const handler = () => {
      // Если dialog закрылся по Esc/backdrop — трактуем как cancel.
      if (props.open) props.onCancel();
    };
    dlg.addEventListener('cancel', handler);
    return () => dlg.removeEventListener('cancel', handler);
  }, [props.open, props.onCancel]);

  if (!props.open) return null;

  return (
    <dialog
      ref={ref}
      className="game-confirm-dialog"
      onClick={(ev) => {
        // backdrop click — закрытие. Native dialog имеет
        // ::backdrop pseudo-element — определяем по target===dialog
        // (иначе клик внутри dialog тоже закроет).
        if (ev.target === ref.current) props.onCancel();
      }}
    >
      <table className="ntable" style={{ minWidth: 320, maxWidth: 600 }}>
        {props.title && (
          <thead>
            <tr><th>{props.title}</th></tr>
          </thead>
        )}
        <tbody>
          <tr>
            <td className="center" style={{ padding: '1em' }}>
              {props.message}
            </td>
          </tr>
          <tr>
            <td className="center">
              <button
                type="button"
                className="button"
                onClick={props.onCancel}
                autoFocus
                style={{ marginRight: 8 }}
              >
                {props.cancelLabel ?? 'Отмена'}
              </button>
              <button
                type="button"
                className="button"
                onClick={props.onConfirm}
                style={
                  props.destructive
                    ? { color: '#fff', background: '#a33', borderColor: '#a33' }
                    : undefined
                }
              >
                {props.confirmLabel ?? 'OK'}
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </dialog>
  );
}

// useConfirm — hook для императивного использования. Возвращает
// функцию `confirm()` которая возвращает Promise<boolean>.
//
// Пример:
//   const { confirm, dialogProps } = useConfirm();
//   ...
//   <ConfirmDialog {...dialogProps} />
//   ...
//   if (await confirm({ message: 'Отозвать флот?' })) recall.mutate();
export function useConfirm(): {
  confirm: (opts: Omit<ConfirmDialogProps, 'open' | 'onConfirm' | 'onCancel'>) => Promise<boolean>;
  dialogProps: ConfirmDialogProps;
} {
  const [state, setState] = useState<{
    open: boolean;
    opts: Omit<ConfirmDialogProps, 'open' | 'onConfirm' | 'onCancel'>;
    resolve: ((v: boolean) => void) | null;
  }>({ open: false, opts: { message: '' }, resolve: null });

  const confirm = useCallback(
    (opts: Omit<ConfirmDialogProps, 'open' | 'onConfirm' | 'onCancel'>) =>
      new Promise<boolean>((resolve) => {
        setState({ open: true, opts, resolve });
      }),
    [],
  );

  const close = useCallback((value: boolean) => {
    setState((prev) => {
      if (prev.resolve) prev.resolve(value);
      return { open: false, opts: prev.opts, resolve: null };
    });
  }, []);

  return {
    confirm,
    dialogProps: {
      ...state.opts,
      open: state.open,
      onConfirm: () => close(true),
      onCancel: () => close(false),
    },
  };
}
