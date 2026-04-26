import { useQuery } from '@tanstack/react-query';
import { portalApi } from '@/api/portal';
import { Link } from '@/components/Link';
import styles from './NewsPage.module.css';

export function NewsListPage() {
  const { data, isLoading } = useQuery({
    queryKey: ['news'],
    queryFn: () => portalApi.news.list(50),
  });

  if (isLoading) return <div className={styles.loading}>Загрузка…</div>;

  return (
    <div className={styles.page}>
      <h1 className={styles.title}>Новости</h1>
      <div className={styles.list}>
        {data?.news.map((n) => (
          <Link key={n.id} href={`/news/${n.id}`} className={styles.card}>
            {n.pinned && <span className={styles.pin}>📌</span>}
            <div className={styles.cardTitle}>{n.title}</div>
            <div className={styles.cardDate}>
              {new Date(n.created_at).toLocaleDateString('ru-RU')}
            </div>
          </Link>
        ))}
        {data?.news.length === 0 && (
          <p className={styles.empty}>Новостей пока нет.</p>
        )}
      </div>
    </div>
  );
}

export function NewsDetailPage({ id }: { id: string }) {
  const { data, isLoading, isError } = useQuery({
    queryKey: ['news', id],
    queryFn: () => portalApi.news.get(id),
  });

  if (isLoading) return <div className={styles.loading}>Загрузка…</div>;
  if (isError || !data) return <div className={styles.error}>Новость не найдена.</div>;

  return (
    <div className={styles.page}>
      <Link href="/news" className={styles.back}>← Назад</Link>
      <article className={styles.article}>
        <h1 className={styles.articleTitle}>{data.title}</h1>
        <div className={styles.articleDate}>
          {new Date(data.created_at).toLocaleDateString('ru-RU', {
            year: 'numeric', month: 'long', day: 'numeric',
          })}
        </div>
        <div className={styles.articleBody}>{data.body}</div>
      </article>
    </div>
  );
}
