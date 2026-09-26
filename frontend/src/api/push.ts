import { apiFetch } from './client';

export function getVapidPublicKey(): Promise<{ publicKey: string }> {
  return apiFetch<{ publicKey: string }>('/api/push/vapid-public-key');
}

export function subscribePush(
  subscription: PushSubscriptionJSON,
): Promise<void> {
  return apiFetch<void>('/api/push/subscribe', {
    method: 'POST',
    body: JSON.stringify(subscription),
  });
}

// reason はブラウザ側の都合で自動解除する場合のみ指定する(サーバーの警告ログに残る)。
export function unsubscribePush(
  endpoint: string,
  reason?: string,
): Promise<void> {
  return apiFetch<void>('/api/push/unsubscribe', {
    method: 'POST',
    body: JSON.stringify({ endpoint, reason }),
  });
}

export type NotificationPreferences = {
  emailEnabled: boolean;
  pushEnabled: boolean;
  disableEmailWhenPushAvailable: boolean;
};

export function getNotificationPreferences(): Promise<NotificationPreferences> {
  return apiFetch<NotificationPreferences>('/api/notification-preferences');
}
