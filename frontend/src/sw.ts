/// <reference lib="webworker" />
/// <reference types="vite/client" />

import { precacheAndRoute } from 'workbox-precaching';
import { API_BASE_URL } from './api/client';

// vite-plugin-pwa (injectManifest) はこのファイルを型チェック用のグローバル self を
// ServiceWorkerGlobalScope として扱えるようにビルドする。
declare const self: ServiceWorkerGlobalScope & {
  __WB_MANIFEST: Array<string | { revision: string | null; url: string }>;
};

// オフラインのアプリシェルキャッシュは使わない(globPatterns: []でプリキャッシュ対象は空)が、
// injectManifestはビルド時にこの self.__WB_MANIFEST プレースホルダーを要求するため呼び出しておく。
precacheAndRoute(self.__WB_MANIFEST);

interface OverdueSitePayload {
  name: string;
  statusLabel: string;
  url: string;
}

// TypeScriptの標準NotificationOptions型には `actions` (persistent notificationのボタン)
// がまだ含まれていないため、Notifications APIの実際の仕様に合わせて拡張する。
interface PersistentNotificationOptions extends NotificationOptions {
  actions?: { action: string; title: string; icon?: string }[];
}

// push イベントのペイロード契約はbackend/internal/usecase/notify_overdue_sites/push_payload.go
// (buildPushPayload) と対応する。urlは必ずワンタイムトークン付きのチェックインリンク
// (/go/{siteId}?token=...) であり、これを開くだけでチェックインが記録される。
// 通知の表示結果をサーバーへ報告する(POST /api/push/ack)。プッシュサービスの2xxは
// 「受理した」だけで表示を保証しないため、権限拒否などで表示できなかった理由を
// サーバーのエラーログに残す目的。報告の失敗で通知処理自体は止めない。
async function reportDisplayResult(
  ackToken: string | undefined,
  status: 'shown' | 'failed',
  error?: unknown,
): Promise<void> {
  if (!ackToken) return;
  try {
    await fetch(`${API_BASE_URL}/api/push/ack`, {
      method: 'POST',
      credentials: 'omit',
      keepalive: true,
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        token: ackToken,
        status,
        permission: typeof Notification === 'undefined' ? 'unknown' : Notification.permission,
        error: error instanceof Error ? error.message : String(error ?? ''),
      }),
    });
  } catch {
    // 報告はベストエフォート。
  }
}

self.addEventListener('push', (event: PushEvent) => {
  if (!event.data) return;

  const { sites, ackToken } = event.data.json() as {
    sites: OverdueSitePayload[];
    ackToken?: string;
  };

  // メールの「サイト名+ステータスの一覧」と「サイト名を記載したCTAボタン」という構成を
  // 踏襲するため、サイトごとに個別の通知を出す(1通にまとめると複数サイト分の
  // アクションボタンを同時に表示できないため)。
  const notifications = sites.map((site) =>
    self.registration.showNotification(site.name, {
      body: site.statusLabel,
      tag: site.url,
      icon: '/logo-192.png',
      badge: '/logo-192.png',
      data: { url: site.url },
      actions: [{ action: 'checkin', title: site.name }],
    } as PersistentNotificationOptions),
  );

  event.waitUntil(
    Promise.all(notifications).then(
      () => reportDisplayResult(ackToken, 'shown'),
      (err) => reportDisplayResult(ackToken, 'failed', err),
    ),
  );
});

self.addEventListener('notificationclick', (event: NotificationEvent) => {
  event.notification.close();
  const url = event.notification.data?.url as string | undefined;
  if (!url) return;

  event.waitUntil(
    self.clients.matchAll({ type: 'window' }).then((clientList) => {
      const existing = clientList.find((c) => c.url === url);
      if (existing) {
        return (existing as WindowClient).focus();
      }
      return self.clients.openWindow(url);
    }),
  );
});

self.addEventListener('install', () => {
  self.skipWaiting();
});

self.addEventListener('activate', (event: ExtendableEvent) => {
  event.waitUntil(self.clients.claim());
});
