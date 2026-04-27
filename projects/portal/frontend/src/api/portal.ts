import { api } from './client';
import type { Universe, NewsItem, FeedbackPost, FeedbackComment, AuthUser, Tokens } from './types';

export const portalApi = {
  universes: {
    list: () => api.get<{ universes: Universe[] }>('/api/universes'),
  },

  news: {
    list: (limit = 20, offset = 0) =>
      api.get<{ news: NewsItem[] }>(`/api/news?limit=${limit}&offset=${offset}`),
    get: (id: string) => api.get<NewsItem>(`/api/news/${id}`),
  },

  feedback: {
    list: (status = 'approved', limit = 20, offset = 0) =>
      api.get<{ posts: FeedbackPost[] }>(
        `/api/feedback?status=${status}&limit=${limit}&offset=${offset}`,
      ),
    get: (id: string) => api.get<FeedbackPost>(`/api/feedback/${id}`),
    create: (title: string, body: string) =>
      api.post<FeedbackPost>('/api/feedback', { title, body }),
    vote: (id: string) => api.post<{ vote_count: number }>(`/api/feedback/${id}/vote`, {}),
    moderate: (id: string, status: string) =>
      api.patch<void>(`/api/feedback/${id}/status`, { status }),
  },

  comments: {
    list: (postId: string) =>
      api.get<{ comments: FeedbackComment[] }>(`/api/feedback/${postId}/comments`),
    add: (postId: string, body: string, parentId?: string) =>
      api.post<FeedbackComment>(`/api/feedback/${postId}/comments`, {
        body,
        ...(parentId !== undefined ? { parent_id: parentId } : {}),
      }),
  },

  auth: {
    register: (username: string, email: string, password: string, consentAccepted: boolean) =>
      api.post<{ user: AuthUser; tokens: Tokens }>('/auth/register', {
        username,
        email,
        password,
        consent_accepted: consentAccepted,
      }),
    login: (login: string, password: string) =>
      api.post<{ user: AuthUser; tokens: Tokens }>('/auth/login', { login, password }),
    refresh: (refresh: string) =>
      api.post<{ tokens: Tokens }>('/auth/refresh', { refresh }),
    logout: (refresh: string) =>
      api.post<void>('/auth/logout', { refresh }),
    me: () => api.get<AuthUser>('/auth/me'),
    // План 44 (152-ФЗ ст. 14): право на удаление ПДн.
    deleteMe: () => api.delete<void>('/auth/users/me'),
  },

  // План 38 Ф.7: billing-service. Кошельки, история, заказы пакетов кредитов.
  billing: {
    balance: () =>
      api.get<{ balance: number; currency_code: string; frozen: boolean }>(
        '/billing/wallet/balance',
      ),
    history: (limit = 50) =>
      api.get<{ transactions: Array<Record<string, unknown>> }>(
        `/billing/wallet/history?limit=${limit}`,
      ),
    packages: () =>
      api.get<{ packages: Array<{ id: string; title: string; amount_kop: number; credits: number; bonus?: number }> }>(
        '/billing/packages',
      ),
    createOrder: (packageId: string) =>
      api.post<{ order: Record<string, unknown>; pay_url: string }>('/billing/orders', {
        package_id: packageId,
      }),
  },
};
