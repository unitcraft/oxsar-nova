interface ProgressBarProps {
  pct: number; // 0–100
  variant?: 'default' | 'success' | 'warning' | 'danger';
  height?: number;
  showLabel?: boolean;
}

export function ProgressBar({ pct, variant = 'default', height = 6, showLabel = false }: ProgressBarProps) {
  const clamped = Math.min(100, Math.max(0, pct));
  const cls = variant === 'default' ? '' : variant;
  return (
    <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
      <div className="ox-progress" style={{ height, flex: 1 }}>
        <div className={`ox-progress-bar ${cls}`} style={{ width: `${clamped}%` }} />
      </div>
      {showLabel && (
        <span style={{ fontFamily: 'var(--ox-mono)', fontSize: 11, color: 'var(--ox-fg-dim)', flexShrink: 0 }}>
          {Math.round(clamped)}%
        </span>
      )}
    </div>
  );
}
