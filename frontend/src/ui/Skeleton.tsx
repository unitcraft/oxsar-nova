import './Skeleton.css';

interface SkeletonProps {
  count?: number;
  height?: number | string;
  width?: number | string;
  circle?: boolean;
  className?: string;
}

export function Skeleton({
  count = 1,
  height = 16,
  width = '100%',
  circle = false,
  className,
}: SkeletonProps) {
  return (
    <>
      {Array.from({ length: count }).map((_, i) => (
        <div
          key={i}
          className={`skeleton ${circle ? 'skeleton-circle' : ''} ${className ?? ''}`}
          style={{
            height: typeof height === 'number' ? `${height}px` : height,
            width: typeof width === 'number' ? `${width}px` : width,
            borderRadius: circle ? '50%' : '6px',
          }}
        />
      ))}
    </>
  );
}

export function ScreenSkeleton() {
  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 24, padding: '16px' }}>
      <Skeleton height={48} width="60%" />
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(150px, 1fr))', gap: 16 }}>
        {Array.from({ length: 3 }).map((_, i) => (
          <div key={i} style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
            <Skeleton height={20} width="80%" />
            <Skeleton height={32} />
            <Skeleton height={16} width="60%" />
          </div>
        ))}
      </div>
      <Skeleton height={200} />
    </div>
  );
}

export function TableSkeleton({ rows = 5 }: { rows?: number }) {
  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(5, 1fr)', gap: 12 }}>
        {Array.from({ length: 5 }).map((_, i) => (
          <Skeleton key={i} height={24} width="80%" />
        ))}
      </div>
      {Array.from({ length: rows }).map((_, i) => (
        <div key={i} style={{ display: 'grid', gridTemplateColumns: 'repeat(5, 1fr)', gap: 12, opacity: 1 - i * 0.1 }}>
          {Array.from({ length: 5 }).map((_, j) => (
            <Skeleton key={j} height={20} />
          ))}
        </div>
      ))}
    </div>
  );
}

export function CardSkeleton() {
  return (
    <div style={{
      padding: '16px',
      border: '1px solid var(--ox-border)',
      borderRadius: '8px',
      display: 'flex',
      flexDirection: 'column',
      gap: 12,
    }}>
      <Skeleton height={24} width="60%" />
      <Skeleton height={20} />
      <Skeleton height={20} width="80%" />
      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12 }}>
        <Skeleton height={32} />
        <Skeleton height={32} />
      </div>
    </div>
  );
}

export function ListSkeleton({ items = 3 }: { items?: number }) {
  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
      {Array.from({ length: items }).map((_, i) => (
        <div key={i} style={{ display: 'flex', gap: 12, alignItems: 'center', opacity: 1 - i * 0.08 }}>
          <Skeleton height={40} width={40} circle />
          <div style={{ flex: 1, display: 'flex', flexDirection: 'column', gap: 6 }}>
            <Skeleton height={18} width="70%" />
            <Skeleton height={14} width="50%" />
          </div>
        </div>
      ))}
    </div>
  );
}

export function ResourceCardSkeleton() {
  return (
    <div style={{
      padding: '12px',
      borderRadius: '8px',
      display: 'flex',
      flexDirection: 'column',
      gap: 8,
      background: 'rgba(56, 189, 248, 0.05)',
      border: '1px solid var(--ox-border)',
    }}>
      <Skeleton height={14} width="60%" />
      <Skeleton height={24} />
      <Skeleton height={12} width="50%" />
    </div>
  );
}

export function ResourceScreenSkeleton() {
  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 24, padding: '16px' }}>
      <Skeleton height={32} width="50%" />
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 12 }}>
        {Array.from({ length: 3 }).map((_, i) => (
          <ResourceCardSkeleton key={i} />
        ))}
      </div>
      <div style={{ padding: '16px', borderRadius: '8px', background: 'rgba(56, 189, 248, 0.05)', border: '1px solid var(--ox-border)', display: 'flex', flexDirection: 'column', gap: 12 }}>
        <Skeleton height={20} width="40%" />
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 16 }}>
          {Array.from({ length: 3 }).map((_, i) => (
            <div key={i} style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
              <Skeleton height={14} width="80%" />
              <Skeleton height={24} />
            </div>
          ))}
        </div>
      </div>
      <TableSkeleton rows={4} />
    </div>
  );
}
