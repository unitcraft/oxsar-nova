import { Modal } from './Modal';

interface ConfirmProps {
  title?: string;
  message: string;
  confirmLabel?: string;
  cancelLabel?: string;
  danger?: boolean;
  onConfirm: () => void;
  onCancel: () => void;
}

export function Confirm({
  title = 'Подтверждение',
  message,
  confirmLabel = 'Подтвердить',
  cancelLabel = 'Отмена',
  danger = false,
  onConfirm,
  onCancel,
}: ConfirmProps) {
  return (
    <Modal
      title={title}
      onClose={onCancel}
      actions={
        <>
          <button type="button" className="btn-ghost" onClick={onCancel}>{cancelLabel}</button>
          <button
            type="button"
            className={danger ? 'btn-danger' : ''}
            onClick={onConfirm}
          >
            {confirmLabel}
          </button>
        </>
      }
    >
      <p style={{ color: 'var(--ox-fg-dim)', lineHeight: 1.6 }}>{message}</p>
    </Modal>
  );
}
