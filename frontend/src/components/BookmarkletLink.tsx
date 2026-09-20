// 開いているページのtitle/urlを付けてこのアプリを新規タブで開くブックマークレット。
// `url` はViteの開発サーバーが予約しているクエリのため、`add_*` の別名を使う。
function buildBookmarklet(origin: string): string {
  return (
    `javascript:(()=>{const u=new URL(${JSON.stringify(origin + '/')});` +
    `u.searchParams.set('add_title',document.title);` +
    `u.searchParams.set('add_url',location.href);` +
    `window.open(u.href,'_blank')})()`
  );
}

// ドラッグしてブックマークバーに置いてもらうためのリンク。Push/Emailトグルと同じ文字スタイルにする。
export default function BookmarkletLink() {
  return (
    <a
      // ReactはJSX上のjavascript:を警告・ブロックするため、DOMへ直接設定する
      ref={(el) =>
        el?.setAttribute('href', buildBookmarklet(window.location.origin))
      }
      onClick={(e) => e.preventDefault()}
      draggable
      className="flex items-center gap-1 whitespace-nowrap font-semibold"
    >
      <img src="/favicon.png" alt="" className="w-4 h-4 shrink-0" />+ Memory
      Tracker
    </a>
  );
}
