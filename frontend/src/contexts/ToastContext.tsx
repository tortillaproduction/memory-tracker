import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useRef,
  useState,
} from 'react';
import { registerToastEmitter } from '../api/client';

type ToastType = 'success' | 'error';

type Toast = {
  id: number;
  message: string;
  type: ToastType;
  // フェードアウトのアニメーション中はtrueにして、CSSアニメーションのクラスを切り替える
  leaving: boolean;
};

type ToastContextValue = {
  showToast: (message: string, type?: ToastType) => void;
  toasts: Toast[];
};

const ToastContext = createContext<ToastContextValue | undefined>(undefined);

// 表示している時間(フェードイン/アウトの時間を含む)
const AUTO_DISMISS_MS = 3200;
// フェードイン/フェードアウトそれぞれのアニメーション時間。index.cssの
// toast-anim-in / toast-anim-out のanimation-durationと合わせること。
const FADE_DURATION_MS = 450;

export function ToastProvider({ children }: { children: React.ReactNode }) {
  const [toasts, setToasts] = useState<Toast[]>([]);
  const idRef = useRef(0);

  const showToast = useCallback(
    (message: string, type: ToastType = 'success') => {
      const id = idRef.current++;
      setToasts((prev) => [...prev, { id, message, type, leaving: false }]);

      // 表示時間の終わり際にフェードアウトを開始する
      setTimeout(() => {
        setToasts((prev) =>
          prev.map((t) => (t.id === id ? { ...t, leaving: true } : t)),
        );
      }, AUTO_DISMISS_MS - FADE_DURATION_MS);

      // フェードアウトのアニメーションが終わったタイミングで実際に取り除く
      setTimeout(() => {
        setToasts((prev) => prev.filter((t) => t.id !== id));
      }, AUTO_DISMISS_MS);
    },
    [],
  );

  // api/client.ts からもネットワーク/サーバーエラーをトースト表示できるようにグローバル登録する
  useEffect(() => {
    registerToastEmitter(showToast);
    return () => registerToastEmitter(null);
  }, [showToast]);

  // 表示位置はToastList側(呼び出し元)に委ねるため、ここではstateの提供のみ行う
  return (
    <ToastContext.Provider value={{ showToast, toasts }}>
      {children}
    </ToastContext.Provider>
  );
}

export function useToast(): ToastContextValue {
  const ctx = useContext(ToastContext);
  if (!ctx) {
    throw new Error('useToast must be used within a ToastProvider');
  }
  return ctx;
}

type ToastListProps = {
  // 'navbar': position:relativeなナビゲーションバー内に配置し、その中央(横・縦とも)に絶対配置で表示する
  // 'fixed' : ナビゲーションバーが無い画面向けに、画面上部中央に固定表示する(フォールバック)
  variant?: 'navbar' | 'fixed';
};

// トースト本体の見た目。内側余白を絞り、1行で収まる横長のピル型にしている。
export function ToastList({ variant = 'fixed' }: ToastListProps) {
  const { toasts } = useToast();

  if (toasts.length === 0) return null;

  const wrapperClass =
    variant === 'navbar'
      ? 'absolute inset-0 flex items-center justify-center z-[100] pointer-events-none'
      : 'fixed inset-x-0 top-3 flex justify-center z-[100] pointer-events-none';

  return (
    <div className={wrapperClass}>
      <div className="flex flex-col items-center gap-1.5">
        {toasts.map((toast) => (
          <div
            key={toast.id}
            className={`alert ${toast.type === 'success' ? 'alert-success' : 'alert-error'} ${toast.leaving ? 'toast-anim-out' : 'toast-anim-in'} pointer-events-auto min-h-0 w-max max-w-[90vw] flex-row flex-nowrap whitespace-nowrap rounded-full shadow-lg`}
            style={{
              paddingBlock: '0.375rem',
              paddingInline: '0.875rem',
              gap: '0.5rem',
            }}
          >
            <span className="text-sm">{toast.message}</span>
          </div>
        ))}
      </div>
    </div>
  );
}
