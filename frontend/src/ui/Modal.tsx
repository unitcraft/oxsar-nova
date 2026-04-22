import { useEffect, useRef } from 'react';

interface ModalProps {
  title: string;
  onClose: () => void;
  children: React.ReactNode;
  actions?: React.ReactNode;
  maxWidth?: number;
}

export function Modal({ title, onClose, children, actions, maxWidth = 480 }: ModalProps) {
  const overlayRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => { if (e.key === 'Escape') onClose(); };
    document.addEventListener('keydown', onKey);
    return () => document.removeEventListener('keydown', onKey);
  }, [onClose]);

  return (
    <div
      className="ox-modal-overlay"
      ref={overlayRef}
      onClick={(e) => { if (e.target === overlayRef.current) onClose(); }}
    >
      <div className="ox-modal" style={{ maxWidth }}>
        <div className="ox-modal-title">{title}</div>
        <button className="ox-modal-close" onClick={onClose} type="button" aria-label="Закрыть">✕</button>
        <div>{children}</div>
        {actions && <div className="ox-modal-actions">{actions}</div>}
      </div>
    </div>
  );
}
