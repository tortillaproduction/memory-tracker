// ブックマークレットは `/?add_title=...&add_url=...` でアプリを開く。
// その値をAdd Siteモーダルの初期値として受け渡すためのヘルパー。

// `url` はViteの開発サーバーが予約しているクエリ(?url)で403になるため、別名にしている。
const PARAM_TITLE = 'add_title';
const PARAM_URL = 'add_url';

export type SitePrefill = { name: string; url: string };

const PREFILL_KEY = 'mt_pending_add_site';
// ログイン途中で放置された値が後から突然モーダルを開かないよう、保持期間を限定する。
const PREFILL_TTL_MS = 10 * 60 * 1000;

function parsePrefill(search: string): SitePrefill | null {
  const params = new URLSearchParams(search);
  const rawUrl = params.get(PARAM_URL);
  if (!rawUrl) return null;

  // javascript: 等を弾くため http/https のみ受け付ける
  try {
    const parsed = new URL(rawUrl);
    if (parsed.protocol !== 'http:' && parsed.protocol !== 'https:') {
      return null;
    }
  } catch {
    return null;
  }

  return { name: (params.get(PARAM_TITLE) ?? '').trim(), url: rawUrl };
}

// 未ログインでOAuthリダイレクトに入る前に呼ぶ。遷移後にクエリが失われるため退避する。
export function stashPrefill(search: string = window.location.search) {
  const prefill = parsePrefill(search);
  if (!prefill) return;
  try {
    sessionStorage.setItem(
      PREFILL_KEY,
      JSON.stringify({ ...prefill, savedAt: Date.now() }),
    );
  } catch {
    // sessionStorageが使えなくてもログイン自体は継続する
  }
}

function readStashed(): SitePrefill | null {
  try {
    const raw = sessionStorage.getItem(PREFILL_KEY);
    if (!raw) return null;
    const saved = JSON.parse(raw) as SitePrefill & { savedAt: number };
    if (Date.now() - saved.savedAt > PREFILL_TTL_MS) return null;
    return { name: saved.name, url: saved.url };
  } catch {
    return null;
  }
}

// URLのクエリを優先し、なければログイン前に退避した値を返す。
// StrictModeで初期化が2回走っても安全なよう、読み取りのみで副作用は持たない。
export function readPrefill(): SitePrefill | null {
  return parsePrefill(window.location.search) ?? readStashed();
}

// 読み取り済みの値を消し、リロードでモーダルが再表示されないようURLも掃除する。
export function clearPrefill() {
  try {
    sessionStorage.removeItem(PREFILL_KEY);
  } catch {
    // 無視してよい
  }
  const params = new URLSearchParams(window.location.search);
  if (params.has(PARAM_URL) || params.has(PARAM_TITLE)) {
    params.delete(PARAM_URL);
    params.delete(PARAM_TITLE);
    const query = params.toString();
    window.history.replaceState(
      null,
      '',
      window.location.pathname +
        (query ? `?${query}` : '') +
        window.location.hash,
    );
  }
}
