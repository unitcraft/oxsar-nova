import { useQuery } from '@tanstack/react-query';
import { Link } from '@/components/Link';
import { portalApi } from '@/api/portal';
import styles from './HomePage.module.css';

export function HomePage() {
  const { data: universesData } = useQuery({
    queryKey: ['universes'],
    queryFn: () => portalApi.universes.list(),
  });
  const { data: newsData } = useQuery({
    queryKey: ['news'],
    queryFn: () => portalApi.news.list(5),
  });

  return (
    <div className={styles.page}>
      <section className={styles.hero}>
        <h1 className={styles.heroTitle}>Oxsar Nova</h1>
        <p className={styles.heroSub}>Браузерная космическая стратегия</p>
      </section>

      <section className={styles.section}>
        <h2 className={styles.sectionTitle}>Вселенные</h2>
        <div className={styles.universeGrid}>
          {universesData?.universes.map((u) => (
            <div key={u.id} className={`${styles.universeCard} ${styles[`status_${u.status}`]}`}>
              <div className={styles.universeName}>{u.name}</div>
              <div className={styles.universeDesc}>{u.description}</div>
              <div className={styles.universeMeta}>
                <span>Скорость ×{u.speed}</span>
                {u.online_players !== undefined && (
                  <span>{u.online_players} онлайн</span>
                )}
              </div>
              {u.status === 'active' && (
                <a
                  href={`https://${u.subdomain}.oxsar-nova.ru`}
                  className={styles.universePlay}
                  rel="noopener noreferrer"
                >
                  Играть
                </a>
              )}
              {u.status === 'upcoming' && (
                <span className={styles.universeSoon}>Скоро</span>
              )}
            </div>
          ))}
        </div>
      </section>

      <section className={styles.section}>
        <div className={styles.sectionHeader}>
          <h2 className={styles.sectionTitle}>Новости</h2>
          <Link href="/news">Все новости →</Link>
        </div>
        <div className={styles.newsList}>
          {newsData?.news.map((n) => (
            <Link key={n.id} href={`/news/${n.id}`} className={styles.newsCard}>
              {n.pinned && <span className={styles.newsPin}>📌</span>}
              <div className={styles.newsTitle}>{n.title}</div>
              <div className={styles.newsDate}>
                {new Date(n.created_at).toLocaleDateString('ru-RU')}
              </div>
            </Link>
          ))}
        </div>
      </section>

      <section className={styles.section}>
        <div className={styles.sectionHeader}>
          <h2 className={styles.sectionTitle}>Предложения игроков</h2>
          <Link href="/feedback">Все предложения →</Link>
        </div>
        <TopFeedbackPreview />
      </section>
    </div>
  );
}

function TopFeedbackPreview() {
  const { data } = useQuery({
    queryKey: ['feedback', 'approved', 3],
    queryFn: () => portalApi.feedback.list('approved', 3),
  });
  return (
    <div className={styles.feedbackList}>
      {data?.posts.map((p) => (
        <Link key={p.id} href={`/feedback/${p.id}`} className={styles.feedbackCard}>
          <div className={styles.feedbackVotes}>▲ {p.vote_count}</div>
          <div className={styles.feedbackTitle}>{p.title}</div>
          <div className={styles.feedbackAuthor}>{p.author_name}</div>
        </Link>
      ))}
    </div>
  );
}
