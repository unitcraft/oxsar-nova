import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { portalApi } from '@/api/portal';
import { useAuthStore } from '@/store/auth';
import { Link } from '@/components/Link';
import styles from './FeedbackPage.module.css';

export function FeedbackListPage() {
  const { data, isLoading } = useQuery({
    queryKey: ['feedback', 'approved'],
    queryFn: () => portalApi.feedback.list('approved', 50),
  });

  if (isLoading) return <div className={styles.loading}>Загрузка…</div>;

  return (
    <div className={styles.page}>
      <div className={styles.header}>
        <h1 className={styles.title}>Предложения игроков</h1>
        <Link href="/feedback/new" className={styles.newBtn}>+ Предложить</Link>
      </div>
      <p className={styles.hint}>Поддержите понравившееся предложение голосом (100 кредитов).</p>
      <div className={styles.list}>
        {data?.posts.map((p) => (
          <Link key={p.id} href={`/feedback/${p.id}`} className={styles.card}>
            <div className={styles.votes}>▲ {p.vote_count}</div>
            <div className={styles.cardBody}>
              <div className={styles.cardTitle}>{p.title}</div>
              <div className={styles.cardMeta}>
                {p.author_name} · {new Date(p.created_at).toLocaleDateString('ru-RU')}
              </div>
            </div>
            {p.status === 'implemented' && <span className={styles.done}>✓ Реализовано</span>}
          </Link>
        ))}
        {data?.posts.length === 0 && (
          <p className={styles.empty}>Одобренных предложений пока нет.</p>
        )}
      </div>
    </div>
  );
}

export function FeedbackNewPage() {
  const { user } = useAuthStore();
  const qc = useQueryClient();
  const [title, setTitle] = useState('');
  const [body, setBody] = useState('');
  const [error, setError] = useState<string | null>(null);

  const mut = useMutation({
    mutationFn: () => portalApi.feedback.create(title, body),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['feedback'] });
      setTitle(''); setBody('');
      window.history.back();
    },
    onError: (e: Error) => setError(e.message),
  });

  if (!user) {
    return (
      <div className={styles.page}>
        <p className={styles.authPrompt}>
          <Link href="/login">Войдите</Link>, чтобы оставить предложение.
        </p>
      </div>
    );
  }

  return (
    <div className={styles.page}>
      <h1 className={styles.title}>Новое предложение</h1>
      <form
        className={styles.form}
        onSubmit={(e) => { e.preventDefault(); mut.mutate(); }}
      >
        <label className={styles.label}>
          Заголовок
          <input
            className={styles.input}
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            maxLength={200}
            required
          />
        </label>
        <label className={styles.label}>
          Описание
          <textarea
            className={styles.textarea}
            value={body}
            onChange={(e) => setBody(e.target.value)}
            rows={6}
            required
          />
        </label>
        {error && <div className={styles.error}>{error}</div>}
        <button className={styles.submit} type="submit" disabled={mut.isPending}>
          {mut.isPending ? 'Отправка…' : 'Отправить на рассмотрение'}
        </button>
      </form>
    </div>
  );
}

export function FeedbackDetailPage({ id }: { id: string }) {
  const { user } = useAuthStore();
  const qc = useQueryClient();
  const [comment, setComment] = useState('');

  const { data: post, isLoading } = useQuery({
    queryKey: ['feedback', id],
    queryFn: () => portalApi.feedback.get(id),
  });
  const { data: commentsData } = useQuery({
    queryKey: ['feedback', id, 'comments'],
    queryFn: () => portalApi.comments.list(id),
  });

  const voteMut = useMutation({
    mutationFn: () => portalApi.feedback.vote(id),
    onSuccess: () => void qc.invalidateQueries({ queryKey: ['feedback', id] }),
  });

  const commentMut = useMutation({
    mutationFn: () => portalApi.comments.add(id, comment),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['feedback', id, 'comments'] });
      setComment('');
    },
  });

  if (isLoading) return <div className={styles.loading}>Загрузка…</div>;
  if (!post) return <div className={styles.error}>Предложение не найдено.</div>;

  return (
    <div className={styles.page}>
      <Link href="/feedback" className={styles.back}>← Назад</Link>
      <div className={styles.postCard}>
        <div className={styles.voteBlock}>
          <button
            className={styles.voteBtn}
            onClick={() => voteMut.mutate()}
            disabled={voteMut.isPending || !user}
            title={user ? '100 кредитов за голос' : 'Войдите, чтобы проголосовать'}
          >
            ▲
          </button>
          <span className={styles.voteCount}>{post.vote_count}</span>
        </div>
        <div className={styles.postBody}>
          <h1 className={styles.postTitle}>{post.title}</h1>
          <div className={styles.postMeta}>
            {post.author_name} · {new Date(post.created_at).toLocaleDateString('ru-RU')}
            {post.status === 'implemented' && <span className={styles.done}> ✓ Реализовано</span>}
          </div>
          <p className={styles.postText}>{post.body}</p>
        </div>
      </div>

      <div className={styles.comments}>
        <h2 className={styles.commentsTitle}>Обсуждение</h2>
        {commentsData?.comments.map((c) => (
          <div key={c.id} className={styles.comment} style={{ marginLeft: c.parent_id ? '2rem' : 0 }}>
            <div className={styles.commentMeta}>
              <strong>{c.author_name}</strong> · {new Date(c.created_at).toLocaleDateString('ru-RU')}
            </div>
            <div className={styles.commentBody}>{c.body}</div>
          </div>
        ))}

        {user && (
          <form
            className={styles.commentForm}
            onSubmit={(e) => { e.preventDefault(); commentMut.mutate(); }}
          >
            <textarea
              className={styles.textarea}
              value={comment}
              onChange={(e) => setComment(e.target.value)}
              rows={3}
              placeholder="Ваш комментарий…"
              required
            />
            <button className={styles.submit} type="submit" disabled={commentMut.isPending}>
              Отправить
            </button>
          </form>
        )}
        {!user && (
          <p className={styles.authPrompt}>
            <Link href="/login">Войдите</Link>, чтобы оставить комментарий.
          </p>
        )}
      </div>
    </div>
  );
}
