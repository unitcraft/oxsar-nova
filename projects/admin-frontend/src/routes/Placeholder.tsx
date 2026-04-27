// Шаблон placeholder для разделов, реализуемых в следующих фазах.
import { PageHeader } from '@/components/layout/PageHeader';
import { Card, CardContent } from '@/components/ui/card';

interface PlaceholderProps {
  title: string;
  phase: string;
  description?: string;
}

export function Placeholder({ title, phase, description }: PlaceholderProps): React.ReactElement {
  return (
    <>
      <PageHeader title={title} {...(description ? { description } : {})} />
      <Card className="max-w-xl">
        <CardContent className="pt-4">
          <p className="text-sm text-muted-foreground">
            Раздел появится в фазе{' '}
            <span className="font-mono-sm text-foreground">{phase}</span> плана 53.
          </p>
        </CardContent>
      </Card>
    </>
  );
}
