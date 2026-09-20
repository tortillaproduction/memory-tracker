import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import tailwindcss from '@tailwindcss/vite';
import { VitePWA } from 'vite-plugin-pwa';

export default defineConfig({
  plugins: [
    react(),
    tailwindcss(),
    VitePWA({
      strategies: 'injectManifest',
      srcDir: 'src',
      filename: 'sw.ts',
      registerType: 'autoUpdate',
      // 開発サーバーでもService Workerを登録する。無効だとnavigator.serviceWorker.readyが
      // 永遠に解決されず、プッシュの購読/解除が開発時に動かない。
      devOptions: {
        enabled: true,
        type: 'module',
      },
      injectManifest: {
        // アプリシェルのオフラインキャッシュは不要(push/notificationclickのみ自前実装)なので、
        // プリキャッシュ対象を空にして純粋なメッセージング用Service Workerとして使う。
        globPatterns: [],
      },
      manifest: {
        name: 'Memory Tracker',
        short_name: 'Memory Tracker',
        description: 'Get reminded before you forget to check on your sites.',
        start_url: '/',
        display: 'standalone',
        background_color: '#EEEBF6',
        theme_color: '#3B1160',
        screenshots: [
          {
            src: '/screenshot-wide.png',
            sizes: '1280x720',
            type: 'image/png',
            form_factor: 'wide',
            label: 'Memory Tracker on desktop',
          },
          {
            src: '/screenshot-narrow.png',
            sizes: '750x1334',
            type: 'image/png',
            form_factor: 'narrow',
            label: 'Memory Tracker on mobile',
          },
        ],
        icons: [
          {
            src: '/logo-192.png',
            sizes: '192x192',
            type: 'image/png',
            purpose: 'any',
          },
          {
            src: '/logo-512.png',
            sizes: '512x512',
            type: 'image/png',
            purpose: 'any',
          },
          {
            src: '/logo-512-maskable.png',
            sizes: '512x512',
            type: 'image/png',
            purpose: 'maskable',
          },
        ],
      },
    }),
  ],
  server: {
    host: true,
    port: 5173,
  },
});
