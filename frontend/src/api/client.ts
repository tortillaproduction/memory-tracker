const API_BASE_URL =
  import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080';

export class ApiError extends Error {
  status: number;
  constructor(status: number, message: string) {
    super(message);
    this.status = status;
  }
}

// ToastProvider がマウント時に自身の showToast を登録する。
// client.ts はReactツリーの外側にあるので、共通のネットワーク/サーバーエラーを
// トーストで通知するために簡易的なグローバルエミッターを経由させる。
type ToastEmitter = (message: string, type?: 'success' | 'error') => void;
let toastEmitter: ToastEmitter | null = null;

export function registerToastEmitter(fn: ToastEmitter | null) {
  toastEmitter = fn;
}

// credentials: 'include' が必須。これが無いとCookieベースセッションのCookieが送られず、
// どのAPIも401 Unauthorizedになる。
export async function apiFetch<T>(
  path: string,
  options: RequestInit = {},
): Promise<T> {
  let res: Response;
  try {
    res = await fetch(`${API_BASE_URL}${path}`, {
      ...options,
      credentials: 'include',
      headers: {
        'Content-Type': 'application/json',
        ...options.headers,
      },
    });
  } catch (err) {
    // オフライン等、リクエスト自体が失敗したケース。個別画面での判定に頼らず共通で通知する。
    toastEmitter?.('Network error. Please check your connection.', 'error');
    throw err;
  }

  if (!res.ok) {
    // 5xxはどの画面でも起こり得る共通の障害なので、個別のエラーハンドリングに任せず一律で通知する。
    // 4xx（401/402/404など）は呼び出し元がケース別に文脈を持ったメッセージを出すので、ここでは通知しない。
    if (res.status >= 500) {
      toastEmitter?.('Server error. Please try again later.', 'error');
    }
    throw new ApiError(res.status, `Request failed: ${res.status}`);
  }

  if (res.status === 204) {
    return undefined as T;
  }
  return res.json() as Promise<T>;
}

export { API_BASE_URL };
