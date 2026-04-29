import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Link } from '@/components/Link';
import { AgeRating } from '@/components/AgeRating';
import { portalApi } from '@/api/portal';
import { useAuthStore } from '@/store/auth';
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

  const user = useAuthStore((s) => s.user);
  const [pendingId, setPendingId] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  // План 72.2: handoff-flow вместо прямой ссылки.
  //   1. POST /api/universes/{id}/session → identity issues code (TTL 30s).
  //   2. window.location.assign(redirect_url) → game-фронт обменивает code.
  //   3. Если юзер не залогинен — Login-редирект.
  //   4. Ошибки: 401 → /login, 503 → toast.
  const handlePlay = async (universeID: string) => {
    if (!user) {
      window.location.assign('/login');
      return;
    }
    setPendingId(universeID);
    setError(null);
    try {
      const res = await portalApi.universes.createSession(universeID);
      window.location.assign(res.redirect_url);
    } catch (err) {
      const msg = err instanceof Error ? err.message : 'Не удалось перейти во вселенную';
      setError(msg);
      setPendingId(null);
    }
  };

  return (
    <div className={styles.page}>
      <section className={styles.hero}>
        <h1 className={styles.heroTitle}>Oxsar Nova</h1>
        <p className={styles.heroSub}>
          Браузерная космическая стратегия <AgeRating size="md" />
        </p>
      </section>

      <section className={styles.section}>
        <h2 className={styles.sectionTitle}>Вселенные</h2>
        {error && <div className={styles.universeError}>{error}</div>}
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
                <button
                  type="button"
                  className={styles.universePlay}
                  onClick={() => void handlePlay(u.id)}
                  disabled={pendingId === u.id}
                  title={user ? '' : 'Войдите чтобы играть'}
                >
                  {pendingId === u.id ? 'Открываем…' : user ? 'Играть' : 'Войти и играть'}
                </button>
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
          <Link href="/news" className={styles.sectionMore}>Все новости →</Link>
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
          <Link href="/feedback" className={styles.sectionMore}>Все предложения →</Link>
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
