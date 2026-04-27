import { useEffect } from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';
import { fetchMe } from '@/lib/auth/flow';
import { Login } from '@/routes/Login';
import { Dashboard } from '@/routes/Dashboard';
import { Protected } from '@/routes/Protected';

export function App(): React.ReactElement {
  useEffect(() => {
    void fetchMe();
  }, []);

  return (
    <Routes>
      <Route path="/login" element={<Login />} />
      <Route
        path="/"
        element={
          <Protected>
            <Dashboard />
          </Protected>
        }
      />
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  );
}
