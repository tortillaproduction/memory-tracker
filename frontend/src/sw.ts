/// <reference lib="webworker" />
/// <reference types="vite/client" />

import { precacheAndRoute } from 'workbox-precaching';

// vite-plugin-pwa (injectManifest) はこのファイルを型チェック用のグローバル self を
// ServiceWorkerGlobalScope として扱えるようにビルドする。
declare const self: ServiceWorkerGlobalScope & {
  __WB_MANIFEST: Array<string | { revision: string | null; url: string }>;
};

// オフラインのアプリシェルキャッシュは使わない(globPatterns: []でプリキャッシュ対象は空)が、
// injectManifestはビルド時にこの self.__WB_MANIFEST プレースホルダーを要求するため呼び出しておく。
precacheAndRoute(self.__WB_MANIFEST);

self.addEventListener('install', () => {
  self.skipWaiting();
});

self.addEventListener('activate', (event: ExtendableEvent) => {
  event.waitUntil(self.clients.claim());
});
