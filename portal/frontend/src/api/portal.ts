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
    register: (username: string, email: string, password: string) =>
      api.post<{ user: AuthUser; tokens: Tokens }>('/auth/register', {
        username,
        email,
        password,
      }),
    login: (login: string, password: string) =>
      api.post<{ user: AuthUser; tokens: Tokens }>('/auth/login', { login, password }),
    refresh: (refresh: string) =>
      api.post<{ tokens: Tokens }>('/auth/refresh', { refresh }),
    me: () => api.get<AuthUser>('/auth/me'),
    creditBalance: () => api.get<{ balance: number }>('/auth/credits/balance'),
  },
};
