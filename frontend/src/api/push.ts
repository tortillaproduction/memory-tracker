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

export function unsubscribePush(endpoint: string): Promise<void> {
  return apiFetch<void>('/api/push/unsubscribe', {
    method: 'POST',
    body: JSON.stringify({ endpoint }),
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
