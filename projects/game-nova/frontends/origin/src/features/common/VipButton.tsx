// VipButton — унифицированная кнопка VIP-старта для всех очередей.
//
// Используется в ConstructionsScreen, ResearchScreen, BuildPanel,
// RepairScreen, DisassembleScreen.
//
// Поведение:
// - Если до конца задачи осталось ≤ tooLateSec (по умолчанию 5) —
//   показывает информационный диалог «подождите», мутацию не вызывает.
// - Иначе — показывает confirm-диалог, и при подтверждении вызывает onVip.
// - При 409 от бекенда — тоже показывает информационный диалог.
// - Repair/Disassemble: confirm не нужен (showConfirm=false), кнопка
//   вызывает onVip сразу.

import { useRef } from 'react';
import { useConfirm, ConfirmDialog } from './ConfirmDialog';
import { useTranslation } from '@/i18n/i18n';

export interface VipButtonProps {
  taskId: string;
  endAt: string;
  onVip: (taskId: string) => void;
  isPending: boolean;
  /** Показывать ли confirm-диалог перед VIP (default: true). */
  showConfirm?: boolean;
  /** Порог в секундах — ниже которого показывается «слишком поздно» (default: 5). */
  tooLateSec?: number;
  /** Метка на кнопке (default: t('buildings','vipBtn')). */
  label?: React.ReactNode;
  /** title-атрибут кнопки (default: t('buildings','vipHint')). */
  title?: string;
  className?: string;
}

export function VipButton({
  taskId,
  endAt,
  onVip,
  isPending,
  showConfirm = true,
  tooLateSec = 5,
  label,
  title,
  className = 'button',
}: VipButtonProps) {
  const { t } = useTranslation();
  const { confirm, dialogProps } = useConfirm();
  const confirmRef = useRef(confirm);
  confirmRef.current = confirm;

  async function handleClick() {
    const remainSec = (new Date(endAt).getTime() - Date.now()) / 1000;
    if (remainSec <= tooLateSec) {
      await confirmRef.current({
        title: t('buildings', 'vipBtn'),
        message: t('buildings', 'vipTooLate'),
        infoOnly: true,
      });
      return;
    }
    if (!showConfirm) {
      onVip(taskId);
      return;
    }
    if (await confirmRef.current({
      title: t('buildings', 'vipBtn'),
      message: t('buildings', 'vipConfirm'),
    })) {
      onVip(taskId);
    }
  }

  return (
    <>
      <button
        type="button"
        className={className}
        disabled={isPending}
        title={title ?? t('buildings', 'vipHint')}
        onClick={() => { void handleClick(); }}
      >
        {label ?? t('buildings', 'vipBtn')}
      </button>
      <ConfirmDialog {...dialogProps} />
    </>
  );
}

