import { apiFetch } from './client';

export type NotificationMode = 'email' | 'gate';

export type NotificationSettings = {
  mode: NotificationMode;
};

// 平文トークンは含まない。発行済みかどうかと、アプリが最後にAPIを使った日時だけ。
export type GateTokenStatus = {
  exists: boolean;
  createdAt: string | null;
  lastUsedAt: string | null;
};

// qrPayload はAndroidアプリが読み取るJSON文字列(APIのベースURLとトークン)。
// 平文トークンを受け取れるのは発行直後のこのレスポンスだけ。
export type IssuedGateToken = {
  token: string;
  qrPayload: string;
};

export function getNotificationSettings(): Promise<NotificationSettings> {
  return apiFetch<NotificationSettings>('/api/notification-settings');
}

export function updateNotificationMode(
  mode: NotificationMode,
): Promise<NotificationSettings> {
  return apiFetch<NotificationSettings>('/api/notification-settings', {
    method: 'PATCH',
    body: JSON.stringify({ mode }),
  });
}

export function getGateTokenStatus(): Promise<GateTokenStatus> {
  return apiFetch<GateTokenStatus>('/api/gate/token');
}

export function issueGateToken(): Promise<IssuedGateToken> {
  return apiFetch<IssuedGateToken>('/api/gate/token', { method: 'POST' });
}
