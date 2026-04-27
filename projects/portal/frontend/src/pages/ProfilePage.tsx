import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { useAuthStore } from '@/store/auth';
import { portalApi } from '@/api/portal';
import { Link } from '@/components/Link';
import styles from './ProfilePage.module.css';

export function ProfilePage() {
  const { user, clearAuth } = useAuthStore();
  const [dangerOpen, setDangerOpen] = useState(false);
  const [confirmText, setConfirmText] = useState('');
  const [deleting, setDeleting] = useState(false);
  const [deleteError, setDeleteError] = useState<string | null>(null);
  // План 38 Ф.7: баланс берётся из billing-service, не auth-service.
  const { data: balanceData } = useQuery({
    queryKey: ['billing-balance'],
    queryFn: () => portalApi.billing.balance(),
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
          <span className={styles.value}>{balanceData?.balance ?? 0}</span>
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

      <section className={styles.danger}>
        <h2 className={styles.dangerTitle}>Опасная зона</h2>
        {!dangerOpen ? (
          <button
            type="button"
            className={styles.dangerBtn}
            onClick={() => setDangerOpen(true)}
          >
            Удалить аккаунт
          </button>
        ) : (
          <>
            <p className={styles.dangerDesc}>
              Удаление аккаунта необратимо. Email и имя игрока будут
              анонимизированы, вход по старым данным станет невозможен.
              Игровые объекты, важные для других участников (история планет,
              рейтинги, переписка), сохранятся с пометкой «удалённый
              пользователь».
            </p>
            <p className={styles.dangerDesc}>
              Чтобы подтвердить, введите слово <code>DELETE</code> в поле ниже.
            </p>
            <input
              type="text"
              className={styles.dangerInput}
              value={confirmText}
              onChange={(e) => setConfirmText(e.target.value)}
              placeholder="DELETE"
              autoComplete="off"
            />
            {deleteError && <div className={styles.dangerError}>{deleteError}</div>}
            <div className={styles.dangerActions}>
              <button
                type="button"
                className={styles.dangerConfirm}
                disabled={confirmText !== 'DELETE' || deleting}
                onClick={() => {
                  setDeleting(true);
                  setDeleteError(null);
                  portalApi.auth
                    .deleteMe()
                    .then(() => {
                      clearAuth();
                      window.location.href = '/';
                    })
                    .catch((err) => {
                      setDeleteError(err instanceof Error ? err.message : 'Ошибка удаления');
                      setDeleting(false);
                    });
                }}
              >
                {deleting ? 'Удаляем…' : 'Подтвердить удаление'}
              </button>
              <button
                type="button"
                className={styles.dangerCancel}
                onClick={() => {
                  setDangerOpen(false);
                  setConfirmText('');
                  setDeleteError(null);
                }}
              >
                Отмена
              </button>
            </div>
          </>
        )}
      </section>
    </div>
  );
}
