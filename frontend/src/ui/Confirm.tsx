import { useTranslation } from '@/i18n/i18n';
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
  title,
  message,
  confirmLabel,
  cancelLabel,
  danger = false,
  onConfirm,
  onCancel,
}: ConfirmProps) {
  const { t } = useTranslation('confirm');
  return (
    <Modal
      title={title ?? t('defaultTitle')}
      onClose={onCancel}
      actions={
        <>
          <button type="button" className="btn-ghost" onClick={onCancel}>{cancelLabel ?? t('defaultCancel')}</button>
          <button
            type="button"
            className={danger ? 'btn-danger' : ''}
            onClick={onConfirm}
          >
            {confirmLabel ?? t('defaultConfirm')}
          </button>
        </>
      }
    >
      <p style={{ color: 'var(--ox-fg-dim)', lineHeight: 1.6 }}>{message}</p>
    </Modal>
  );
}
