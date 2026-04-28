export interface Universe {
  id: string;
  name: string;
  description: string;
  subdomain: string;
  status: 'active' | 'maintenance' | 'upcoming' | 'retired';
  speed: number;
  deathmatch: boolean;
  online_players?: number;
  total_players?: number;
  launched_at: string;
}

export interface NewsItem {
  id: string;
  title: string;
  body: string;
  author_id: string;
  published: boolean;
  pinned: boolean;
  created_at: string;
  updated_at: string;
}

export interface FeedbackPost {
  id: string;
  author_id: string;
  author_name: string;
  title: string;
  body: string;
  status: 'pending' | 'approved' | 'rejected' | 'implemented';
  vote_count: number;
  created_at: string;
  updated_at: string;
}

export interface FeedbackComment {
  id: string;
  post_id: string;
  parent_id?: string;
  author_id: string;
  author_name: string;
  body: string;
  created_at: string;
  edited_at?: string;
}

export interface AuthUser {
  id: string;
  username: string;
  email: string;
  roles: string[];
}

// План 63: формат ответа login/refresh по RFC 6749 §5.1.
// access_token + token_type:"Bearer" + expires_in (TTL в секундах) + refresh_token.
// Поле user — non-RFC, но widespread practice для SPA login endpoint'ов.
export interface TokenResponse {
  access_token: string;
  token_type: string;
  expires_in: number;
  refresh_token: string;
  user?: AuthUser;
}
