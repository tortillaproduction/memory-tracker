import { stashPrefill } from '../lib/sitePrefill';
import { API_BASE_URL, apiFetch } from './client';

export type CurrentUser = {
  id: string;
  email: string;
  name: string;
  pictureUrl: string;
  planType: 'free' | 'premium';
  maxSites: number;
};

// ログイン/ログアウトはページ全体がリロードされるフローなので、SPAのstateでは
// 完了をトースト表示できない。sessionStorageにフラグを残し、次回マウント時に
// App側で検知してトースト表示する。
export const LOGIN_TOAST_FLAG_KEY = 'mt_show_login_toast';
export const LOGOUT_TOAST_FLAG_KEY = 'mt_show_logout_toast';

// ログインはリダイレクトフローのため、fetchではなくブラウザ遷移させる。
export function redirectToGoogleLogin() {
  // ブックマークレット経由で開かれた場合、ログイン後にモーダルへ引き継ぐ
  stashPrefill();
  try {
    sessionStorage.setItem(LOGIN_TOAST_FLAG_KEY, '1');
  } catch {
    // プライベートブラウジング等でsessionStorageが使えなくてもログイン自体は継続する
  }
  window.location.href = `${API_BASE_URL}/api/auth/google/login`;
}

export function fetchCurrentUser(): Promise<CurrentUser> {
  return apiFetch<CurrentUser>('/api/auth/me');
}

export async function logout(): Promise<void> {
  await apiFetch<void>('/api/auth/logout', { method: 'POST' });
  try {
    sessionStorage.setItem(LOGOUT_TOAST_FLAG_KEY, '1');
  } catch {
    // プライベートブラウジング等でsessionStorageが使えなくてもログアウト自体は継続する
  }
}
