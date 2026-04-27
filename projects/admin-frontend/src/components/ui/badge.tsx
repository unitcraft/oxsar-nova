// shadcn/ui Badge — status indicators.
import * as React from 'react';
import { cva, type VariantProps } from 'class-variance-authority';
import { cn } from '@/lib/utils';

const badgeVariants = cva(
  'inline-flex items-center rounded-md border px-2 py-0.5 text-2xs font-medium transition-colors',
  {
    variants: {
      variant: {
        default: 'border-transparent bg-primary/15 text-primary',
        secondary: 'border-transparent bg-muted text-muted-foreground',
        destructive: 'border-transparent bg-destructive/15 text-destructive',
        success: 'border-transparent bg-success/15 text-success',
        warning: 'border-transparent bg-warning/15 text-warning',
        outline: 'text-foreground',
      },
    },
    defaultVariants: {
      variant: 'default',
    },
  },
);

export interface BadgeProps
  extends React.HTMLAttributes<HTMLSpanElement>,
    VariantProps<typeof badgeVariants> {}

export function Badge({ className, variant, ...props }: BadgeProps): React.ReactElement {
  return <span className={cn(badgeVariants({ variant }), className)} {...props} />;
}

export { badgeVariants };
