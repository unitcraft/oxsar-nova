import { useQuery } from '@tanstack/react-query';
import { useAuthStore } from '@/store/auth';
import { portalApi } from '@/api/portal';
import { Link } from '@/components/Link';
import styles from './ProfilePage.module.css';

export function ProfilePage() {
  const { user, clearAuth } = useAuthStore();
  const { data: balanceData } = useQuery({
    queryKey: ['credits-balance'],
    queryFn: () => portalApi.auth.creditBalance(),
    enabled: !!user,
  });

  if (!user) {
    return (
      <div className={styles.page}>
        <p className={styles.prompt}>
          <Link href="/login">Войдите</Link>, чтобы увидеть профиль.
        </p>
      </div>
    );
  }

  return (
    <div className={styles.page}>
      <h1 className={styles.title}>Профиль</h1>
      <div className={styles.card}>
        <div className={styles.row}>
          <span className={styles.label}>Имя игрока</span>
          <span className={styles.value}>{user.username}</span>
        </div>
        <div className={styles.row}>
          <span className={styles.label}>Email</span>
          <span className={styles.value}>{user.email}</span>
        </div>
        <div className={styles.row}>
          <span className={styles.label}>Глобальные кредиты</span>
          <span className={styles.value}>{balanceData?.balance ?? user.global_credits}</span>
        </div>
        <div className={styles.row}>
          <span className={styles.label}>Роли</span>
          <span className={styles.value}>{user.roles.join(', ')}</span>
        </div>
      </div>
      <button
        className={styles.logout}
        onClick={() => { clearAuth(); window.location.href = '/'; }}
      >
        Выйти
      </button>
    </div>
  );
}
