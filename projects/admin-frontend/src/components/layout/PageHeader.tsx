// PageHeader: H1 + опциональное description + правая action-зона.
// Используется во всех content-страницах.
import { cn } from '@/lib/utils';

interface PageHeaderProps {
  title: string;
  description?: React.ReactNode;
  action?: React.ReactNode;
  className?: string;
}

export function PageHeader({
  title,
  description,
  action,
  className,
}: PageHeaderProps): React.ReactElement {
  return (
    <div className={cn('flex flex-wrap items-start justify-between gap-3 pb-4', className)}>
      <div className="space-y-0.5">
        <h1 className="text-base font-semibold text-foreground">{title}</h1>
        {description && <p className="text-xs text-muted-foreground">{description}</p>}
      </div>
      {action && <div className="flex items-center gap-2">{action}</div>}
    </div>
  );
}
