import { useCallback, useEffect, useState } from 'react';
import { getVapidPublicKey, subscribePush, unsubscribePush } from '../api/push';

export type NotificationChannel = 'email' | 'push';

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
  // Push APIの有無はブラウザ/OSで異なる: Chrome/Edge/Firefoxはブラウザタブのままでも
  // trueになるが、iOS SafariはPWAとしてインストールするまでPushManagerが存在しないため
  // falseのままになる。「インストール済みかどうか」を別途判定しなくても、この1つの
  // フラグだけでUIの有効/無効を正しく出し分けられる。
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
    } finally {
      setIsBusy(false);
    }
  }, []);

  const unsubscribe = useCallback(async () => {
    if (!subscription) return;
    setIsBusy(true);
    try {
      await unsubscribePush(subscription.endpoint);
      await subscription.unsubscribe();
      setSubscription(null);
    } finally {
      setIsBusy(false);
    }
  }, [subscription]);

  // メール/プッシュの二択スライド用。'push'を選ぶと購読、'email'を選ぶと解除する。
  const selectChannel = useCallback(
    (channel: NotificationChannel) =>
      channel === 'push' ? subscribe() : unsubscribe(),
    [subscribe, unsubscribe],
  );

  return {
    isSupported,
    isBusy,
    channel: (subscription ? 'push' : 'email') as NotificationChannel,
    selectChannel,
  };
}
