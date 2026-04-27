import { useState, useEffect } from 'react';
import { Layout } from '@/components/Layout';
import { HomePage } from '@/pages/HomePage';
import { NewsListPage, NewsDetailPage } from '@/pages/NewsPage';
import { FeedbackListPage, FeedbackNewPage, FeedbackDetailPage } from '@/pages/FeedbackPage';
import { LoginPage, RegisterPage } from '@/pages/AuthPage';
import { ProfilePage } from '@/pages/ProfilePage';
import { PrivacyPage } from '@/pages/PrivacyPage';

function usePathname() {
  const [path, setPath] = useState(window.location.pathname);
  useEffect(() => {
    const handler = () => setPath(window.location.pathname);
    window.addEventListener('popstate', handler);
    return () => window.removeEventListener('popstate', handler);
  }, []);
  return path;
}

export function App() {
  const path = usePathname();

  let page: React.ReactNode;

  if (path === '/') {
    page = <HomePage />;
  } else if (path === '/news') {
    page = <NewsListPage />;
  } else if (path.startsWith('/news/')) {
    page = <NewsDetailPage id={path.slice(6)} />;
  } else if (path === '/feedback') {
    page = <FeedbackListPage />;
  } else if (path === '/feedback/new') {
    page = <FeedbackNewPage />;
  } else if (path.startsWith('/feedback/')) {
    page = <FeedbackDetailPage id={path.slice(10)} />;
  } else if (path === '/login') {
    page = <LoginPage />;
  } else if (path === '/register') {
    page = <RegisterPage />;
  } else if (path === '/profile') {
    page = <ProfilePage />;
  } else if (path === '/privacy') {
    page = <PrivacyPage />;
  } else {
    page = <div style={{ padding: '3rem', textAlign: 'center', color: 'var(--color-muted)' }}>Страница не найдена</div>;
  }

  return <Layout>{page}</Layout>;
}
