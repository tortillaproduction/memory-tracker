import { useCallback, useEffect, useState } from 'react';
import { useToast } from '../contexts/ToastContext';
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

// Service Workerが未登録だとreadyは永遠に解決されないため、タイムアウトで打ち切る。
function getRegistration(): Promise<ServiceWorkerRegistration> {
  return Promise.race([
    navigator.serviceWorker.ready,
    new Promise<never>((_, reject) =>
      setTimeout(() => reject(new Error('Service worker not ready')), 5000),
    ),
  ]);
}

export function usePushSubscription(enabled: boolean) {
  const { showToast } = useToast();
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
  // 購読処理中も選択した側へつまみを即座に動かすための、一時的な選択状態。
  const [pendingChannel, setPendingChannel] =
    useState<NotificationChannel | null>(null);

  useEffect(() => {
    if (!isSupported || !enabled) return;
    getRegistration()
      .then((registration) => registration.pushManager.getSubscription())
      .then(setSubscription)
      .catch(() => setSubscription(null));
  }, [isSupported, enabled]);

  const subscribe = useCallback(async () => {
    setIsBusy(true);
    try {
      const permission = await Notification.requestPermission();
      if (permission !== 'granted') return;

      const registration = await getRegistration();
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
    setIsBusy(true);
    try {
      // stateではなくブラウザの現在の購読を直接取得する(stateが古い場合でも確実に解除するため)。
      const registration = await getRegistration();
      const current = await registration.pushManager.getSubscription();
      if (!current) {
        setSubscription(null);
        return;
      }
      try {
        await unsubscribePush(current.endpoint);
      } finally {
        // サーバー側の削除に失敗してもブラウザ側は解除する。残った購読は
        // 送信時に410で自動削除されるため、切替が戻ってしまう状態を避けられる。
        await current.unsubscribe();
        setSubscription(null);
      }
    } finally {
      setIsBusy(false);
    }
  }, []);

  // メール/プッシュの二択スライド用。'push'を選ぶと購読、'email'を選ぶと解除する。
  // 失敗・拒否された場合はpendingを解除して実際の購読状態に戻す。
  const selectChannel = useCallback(
    async (channel: NotificationChannel) => {
      // 処理中の二重操作は無視する(無効化するとフォーカスが外れてメニューが閉じるため、
      // disabledにはせずここで弾く)。
      if (isBusy) return;
      setPendingChannel(channel);
      try {
        await (channel === 'push' ? subscribe() : unsubscribe());
      } catch {
        showToast('Failed to update notification setting.', 'error');
      } finally {
        setPendingChannel(null);
      }
    },
    [isBusy, subscribe, unsubscribe, showToast],
  );

  return {
    isSupported,
    isBusy,
    channel: (pendingChannel ??
      (subscription ? 'push' : 'email')) as NotificationChannel,
    selectChannel,
  };
}
