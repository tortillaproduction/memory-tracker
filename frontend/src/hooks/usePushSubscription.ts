import { useCallback, useEffect, useState } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import {
  getNotificationPreferences,
  getVapidPublicKey,
  subscribePush,
  unsubscribePush,
  updateNotificationPreferences,
} from '../api/push';

// VAPID公開鍵(base64url)をpushManager.subscribeが要求するUint8Arrayに変換する。
function urlBase64ToUint8Array(base64String: string): Uint8Array<ArrayBuffer> {
  const padding = '='.repeat((4 - (base64String.length % 4)) % 4);
  const base64 = (base64String + padding).replace(/-/g, '+').replace(/_/g, '/');
  const rawData = window.atob(base64);
  const outputArray = new Uint8Array(rawData.length);
  for (let i = 0; i < rawData.length; i++) {
    outputArray[i] = rawData.charCodeAt(i);
  }
  return outputArray;
}

export function usePushSubscription(enabled: boolean) {
  const queryClient = useQueryClient();
  const [isSupported] = useState(
    () =>
      'serviceWorker' in navigator &&
      'PushManager' in window &&
      'Notification' in window,
  );
  const [subscription, setSubscription] = useState<PushSubscription | null>(
    null,
  );
  const [isBusy, setIsBusy] = useState(false);

  const preferencesQuery = useQuery({
    queryKey: ['notificationPreferences'],
    queryFn: getNotificationPreferences,
    enabled,
  });

  useEffect(() => {
    if (!isSupported || !enabled) return;
    navigator.serviceWorker.ready
      .then((registration) => registration.pushManager.getSubscription())
      .then(setSubscription)
      .catch(() => setSubscription(null));
  }, [isSupported, enabled]);

  const subscribe = useCallback(async () => {
    setIsBusy(true);
    try {
      const permission = await Notification.requestPermission();
      if (permission !== 'granted') return;

      const registration = await navigator.serviceWorker.ready;
      const { publicKey } = await getVapidPublicKey();
      const sub = await registration.pushManager.subscribe({
        userVisibleOnly: true,
        applicationServerKey: urlBase64ToUint8Array(publicKey),
      });
      await subscribePush(sub.toJSON() as PushSubscriptionJSON);
      setSubscription(sub);
      await queryClient.invalidateQueries({
        queryKey: ['notificationPreferences'],
      });
    } finally {
      setIsBusy(false);
    }
  }, [queryClient]);

  const unsubscribe = useCallback(async () => {
    if (!subscription) return;
    setIsBusy(true);
    try {
      await unsubscribePush(subscription.endpoint);
      await subscription.unsubscribe();
      setSubscription(null);
      await queryClient.invalidateQueries({
        queryKey: ['notificationPreferences'],
      });
    } finally {
      setIsBusy(false);
    }
  }, [subscription, queryClient]);

  const setDisableEmailWhenPushAvailable = useCallback(
    async (value: boolean) => {
      await updateNotificationPreferences(value);
      await queryClient.invalidateQueries({
        queryKey: ['notificationPreferences'],
      });
    },
    [queryClient],
  );

  return {
    isSupported,
    isSubscribed: !!subscription,
    isBusy,
    preferences: preferencesQuery.data,
    subscribe,
    unsubscribe,
    setDisableEmailWhenPushAvailable,
  };
}
