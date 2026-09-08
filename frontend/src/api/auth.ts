import { API_BASE_URL, apiFetch } from './client';

export type CurrentUser = {
  id: string;
  email: string;
  name: string;
  pictureUrl: string;
  planType: 'free' | 'premium';
  maxSites: number;
};

// ログインはリダイレクトフローのため、fetchではなくブラウザ遷移させる。
export function redirectToGoogleLogin() {
  window.location.href = `${API_BASE_URL}/api/auth/google/login`;
}

export function fetchCurrentUser(): Promise<CurrentUser> {
  return apiFetch<CurrentUser>('/api/auth/me');
}

export async function logout(): Promise<void> {
  await apiFetch<void>('/api/auth/logout', { method: 'POST' });
}
