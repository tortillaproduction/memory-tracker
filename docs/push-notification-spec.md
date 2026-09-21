# プッシュ通知（Web Push）仕様

期限切れサイトのプッシュ通知が、どの状況で届き、どのように表示されるかをまとめた仕様書。
セットアップ手順（VAPID鍵の生成など）は[README](../README.md)の「ブラウザ通知（Web Push / PWA）の設定」を参照。

## 情報の確度について

- **コードで確認済み**: 該当ファイルを読んで確認した挙動。
- **ブラウザの一般的な挙動（要実機確認）**: ブラウザやOSの仕様に依存する挙動。コードだけでは確定できないため、変更や検証の際は実機で確認すること。該当箇所には「(要実機確認)」と付けている。

## 仕組み（コードで確認済み）

- 通知バッチが5分ごとに動き、期限切れサイトを検出してWeb Pushを送信する（`backend/cmd/api/main.go`、`NewNotificationScheduler(..., notificationInterval(logger)（デフォルト5分、環境変数`NOTIFICATION_INTERVAL`で変更可）, ...)`）。
- 送信先は、ブラウザ提供元のプッシュサービス（Chromeなら FCM など）。サーバーは `webpush-go` でそこへメッセージを送る（`backend/internal/infrastructure/notification/push_sender.go`）。メッセージのTTLは1時間。
- 通知を表示するのはService Worker（`frontend/src/sw.ts` の `push` イベント）であり、アプリのタブやウィンドウではない。そのため、アプリのタブが開いているかどうかは通知が届くかどうかに影響しない。

### 通知の見え方

サイトごとに1件ずつ、OSの通知として表示される。

| 項目 | 内容 |
|---|---|
| タイトル | サイト名 |
| 本文 | `last checked 30h ago - every 24h` / `not checked yet - every 24h` のような文言 |
| アイコン | `/logo-192.png` |
| ボタン | サイト名（1つ） |

### 通知をクリックしたとき

- ワンタイムトークン付きのチェックインリンク `/go/{siteId}?token=...` が開き、チェックインが記録される。
- 同じURLのウィンドウが既にあればそれにフォーカスし、なければ新しく開く（`notificationclick`）。

## タブで利用する場合

| 状況 | 届くか | 届き方 |
|---|---|---|
| アプリのタブだけ閉じた（ブラウザは起動中） | 届く | OSの通知として表示される。 |
| ウィンドウを閉じたが、ブラウザのプロセスは残っている | 届く (要実機確認) | 同じくOSの通知として表示される。Windowsの Chrome/Edge（バックグラウンド実行が既定でON）と、Macで ⌘Q せずウィンドウだけ閉じた場合が該当する。 |
| ブラウザを完全終了した、またはPCがスリープ・オフライン | その場では届かない (要実機確認) | ブラウザが起動した時点でTTL（1時間）以内なら、起動時にまとめて表示される。1時間を過ぎたメッセージは破棄される。 |

## インストール（PWA）して利用する場合

仕組みは同じで、Service Workerがプッシュを受け取る。アプリのウィンドウを閉じていても通知は届く。

- **スマホ（Android / iOS）** (要実機確認): OSのプッシュ基盤（FCM / APNs）経由で届くため、アプリを閉じていても、ブラウザを起動していなくてもOS通知として表示される。
  - iOSは、ホーム画面に追加したPWAでないとプッシュ自体が使えない。未インストールの場合、設定画面のトグルは無効になる（`frontend/src/hooks/usePushSubscription.ts`、コードで確認済み）。
- **デスクトップ** (要実機確認): タブ利用の場合と同様、ブラウザのプロセスが動いていることが条件。
- 通知クリック時にPWAのウィンドウで開くか、ブラウザで開くかはOS・ブラウザによって異なる (要実機確認)。

## 注意すべき仕様

1. **プッシュが失われてもメールには切り替わらない。**
   - サーバーは、プッシュサービスが送信を受け付けた時点で「成功」とみなす（`backend/internal/usecase/notify_overdue_sites/notify_overdue_sites.go` の `sendPush`）。
   - 成功すると、既定ではメールを送らず、`notification_logs` に記録する。
   - ブラウザが終了していて、1時間以内に受け取れなかった場合、メールへのフォールバックも再送もない。次の通知は `interval_hours` が経過してからになる。
2. **メールに切り替わるのは、購読が失効した場合だけ。** プッシュサービスが 404 / 410 を返した購読は自動で削除され（セルフヒーリング）、同じサイクル内でメールが送られる。ブラウザ側で通知の許可を解除した場合も、次回の送信時にこの経路で検出される。
3. **通知のボタン（`actions`）の表示はOS・ブラウザ依存。** Chrome/EdgeのデスクトップとAndroidでは表示される。iOSやmacOSのSafariでは表示されない場合があるため、通知本体をタップしたときの動作が主な導線になる (要実機確認)。
4. 同じサイトの通知は `tag` にチェックインURLを設定しているため、同じURLの通知は重複せず置き換えられる（`sw.ts`）。

## 関連ファイル

| 役割 | ファイル |
|---|---|
| Service Worker（表示・クリック処理） | `frontend/src/sw.ts` |
| 購読・解除のフック | `frontend/src/hooks/usePushSubscription.ts` |
| PWAマニフェスト・SW設定 | `frontend/vite.config.ts` |
| バッチの通知ロジック | `backend/internal/usecase/notify_overdue_sites/notify_overdue_sites.go` |
| プッシュのペイロード組み立て | `backend/internal/usecase/notify_overdue_sites/push_payload.go` |
| Web Push送信 | `backend/internal/infrastructure/notification/push_sender.go` |
