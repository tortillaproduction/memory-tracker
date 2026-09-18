import { useCallback, useEffect, useState } from 'react';

// beforeinstallprompt はSpecに正式採用されていないため標準libに型が無い。
interface BeforeInstallPromptEvent extends Event {
  prompt: () => Promise<void>;
  userChoice: Promise<{ outcome: 'accepted' | 'dismissed' }>;
}

function isStandalone(): boolean {
  return (
    window.matchMedia('(display-mode: standalone)').matches ||
    // iOS Safariの独自プロパティ
    (window.navigator as Navigator & { standalone?: boolean }).standalone ===
      true
  );
}

export function useInstallPrompt() {
  const [deferredPrompt, setDeferredPrompt] =
    useState<BeforeInstallPromptEvent | null>(null);
  const [isInstalled, setIsInstalled] = useState(isStandalone);

  useEffect(() => {
    function onBeforeInstallPrompt(event: Event) {
      event.preventDefault();
      setDeferredPrompt(event as BeforeInstallPromptEvent);
    }
    function onAppInstalled() {
      setIsInstalled(true);
      setDeferredPrompt(null);
    }

    window.addEventListener('beforeinstallprompt', onBeforeInstallPrompt);
    window.addEventListener('appinstalled', onAppInstalled);
    return () => {
      window.removeEventListener('beforeinstallprompt', onBeforeInstallPrompt);
      window.removeEventListener('appinstalled', onAppInstalled);
    };
  }, []);

  const promptInstall = useCallback(async () => {
    if (!deferredPrompt) return;
    await deferredPrompt.prompt();
    await deferredPrompt.userChoice;
    setDeferredPrompt(null);
  }, [deferredPrompt]);

  return {
    isInstalled,
    // Chrome/Android/desktopではbeforeinstallpromptで判定できる。
    // iOS Safariはこのイベントを発火しないため、canPromptはfalseのままになる
    // (呼び出し側でisIOSと組み合わせて「ホーム画面に追加」の案内文を出す)。
    canPrompt: !!deferredPrompt,
    isIOS: /iphone|ipad|ipod/i.test(window.navigator.userAgent),
    promptInstall,
  };
}
