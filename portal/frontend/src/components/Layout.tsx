import { useAuthStore } from '@/store/auth';
import { Link } from './Link';
import styles from './Layout.module.css';

export function Layout({ children }: { children: React.ReactNode }) {
  const { user, clearAuth } = useAuthStore();

  return (
    <div className={styles.root}>
      <header className={styles.header}>
        <div className={styles.headerInner}>
          <Link href="/" className={styles.logo}>Oxsar Nova</Link>
          <nav className={styles.nav}>
            <Link href="/news" className={styles.navLink}>Новости</Link>
            <Link href="/feedback" className={styles.navLink}>Предложения</Link>
          </nav>
          <div className={styles.actions}>
            {user ? (
              <>
                <Link href="/profile" className={styles.navLink}>{user.username}</Link>
                <button
                  className={styles.logoutBtn}
                  onClick={() => { clearAuth(); window.location.href = '/'; }}
                >
                  Выйти
                </button>
              </>
            ) : (
              <>
                <Link href="/login" className={styles.navLink}>Войти</Link>
                <Link href="/register" className={styles.registerBtn}>Регистрация</Link>
              </>
            )}
          </div>
        </div>
      </header>
      <main className={styles.main}>{children}</main>
      <footer className={styles.footer}>
        <span>© 2026 Oxsar Nova</span>
      </footer>
    </div>
  );
}
