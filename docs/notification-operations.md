# 通知の運用ガイド（リトライ・ログ・切り分け）

通知バッチの失敗時の挙動と、プッシュ通知が「送信成功なのに届かない」ときの調査方法をまとめた運用資料。
通知が届く条件・見え方は[push-notification-spec.md](push-notification-spec.md)、VAPID鍵などのセットアップは[README](../README.md)を参照。

## 通知バッチの基本

- `NOTIFICATION_INTERVAL`（デフォルト5分）ごとに、期限切れかつ未通知のサイトをユーザー単位でまとめて通知する（`backend/internal/usecase/notify_overdue_sites`）。
- 通知が成功したときだけ `notification_logs` に記録され、以後 `interval_hours` の間は再通知されない。
- 失敗した場合はログに記録されないため、次のサイクルでも対象に残る（＝再試行される）。

## 失敗時のリトライ（バックオフ）

失敗を無制限に毎サイクル再試行しないよう、ユーザー単位で指数バックオフをかけている。

| 項目 | 値 |
|---|---|
| 初回の待ち時間 | 5分 |
| 増え方 | 失敗のたびに2倍 |
| 上限 | 6時間 |
| 解除 | 通知に成功したとき |
| 保持場所 | プロセス内メモリのみ（再起動でリセットされ、次サイクルで1回再試行する） |

- バックオフ中のユーザーはそのサイクルでスキップされる。
- 失敗時のエラーログに `retryIn`（次の再試行までの待ち時間）が出る。
- 送信先起因の恒久的なエラーでも同じ扱い。原因を直さない限り失敗し続け、6時間おきの再試行になる。
- プッシュ成功・メール失敗の場合、次の再試行でプッシュも再送される（重複し得る）。

### 代表的なエラー

| ログ | 原因と対処 |
|---|---|
| `resend send: You can only send testing emails to your own email address ...` | Resendでドメイン未認証のまま、`from`が`onboarding@resend.dev`等のまま。resend.com/domainsでドメインを認証し、`EMAIL_FROM_ADDRESS`をそのドメインのアドレスにする。 |
| `webpush send: unexpected status N from <host>: <本文> (retry-after=...)` | プッシュサービスが拒否。401/403はVAPID鍵・subject不正、413はペイロード過大、429はレート制限（`retry-after`を参照）。 |
| `push notifications are not configured (VAPID keys missing)` | `VAPID_PUBLIC_KEY`/`VAPID_PRIVATE_KEY`未設定。メールにフォールバックする。 |

## プッシュ通知の「届いたか」の追跡

プッシュサービスの2xxは「受理した」だけで、端末での表示は保証しない。そのためService Workerが表示結果をサーバーへ報告する（ACK）。

### 流れ

1. バッチが購読（端末）ごとに署名付きACKトークンを発行し、ペイロード（`ackToken`）に含めて送信する。
2. `frontend/src/sw.ts` が `showNotification` の結果を `POST /api/push/ack` へ報告する。Cookieは使わずトークンで認証する。
3. サーバーが結果をログに残す。10分以内に報告がなければ警告を1回出す。

### ログの見方

| レベル | メッセージ | 意味 |
|---|---|---|
| Info | `push accepted by push service` | プッシュサービスが受理した（`pushHost`、`status`付き）。 |
| Info | `push notification displayed on client` | 端末で表示できた。`latency`は送信から報告までの時間。 |
| Error | `push notification was not displayed on client` | 端末で表示できなかった。`reason`に権限が`denied`等の理由が入る。 |
| Warn | `push notification accepted by push service but not acknowledged by client` | 10分報告なし。`hint`に想定原因を出す（下記）。 |
| Warn | `push subscription removed by client` | ブラウザ側が権限喪失を検知して購読を自動解除した。`reason`は`permission_denied`等。 |

### 「未確認」警告の想定原因

- ブラウザが終了している、またはPCがスリープ・オフライン（TTL1時間以内なら復帰時に表示され、`late=true`付きで遅れて報告が届く）。
- ブラウザの通知権限が取り消された（次回のアプリ起動時に自動解除される）。
- 古いService Workerのまま（ACK未対応版）で、報告自体が送られていない。

### 権限喪失の自動解除

アプリ起動時と、通知権限の変更イベント時に `Notification.permission` を確認する。`granted`でなければ購読を解除し、トーストで通知し、サーバーへ理由を送る。解除後はメールにフォールバックする。

## 既知の限界

- ブラウザ権限が`granted`のまま、OSの集中モード・通知オフ等で表示だけ抑制された場合は、`showNotification`が成功するため検知できない。
- ACKの追跡状態はメモリ上のみ。再起動をまたいだ未確認の警告は出ない。
- ACKトークンは`SESSION_SECRET`で署名する（用途プレフィックスでチェックイントークンとは別物）。変更すると発行済みトークンは無効になる。

## 関連ファイル

| 役割 | ファイル |
|---|---|
| バッチ・バックオフ | `backend/internal/usecase/notify_overdue_sites/notify_overdue_sites.go` |
| ACK追跡・警告 | `backend/internal/usecase/push_delivery/tracker.go` |
| ACKトークン | `backend/internal/infrastructure/auth/push_ack_token.go` |
| プッシュ送信・ログ | `backend/internal/infrastructure/notification/push_sender.go` |
| ACK/解除エンドポイント | `backend/internal/interface/http/handler/push_handler.go` |
| 表示報告 | `frontend/src/sw.ts` |
| 権限同期 | `frontend/src/hooks/usePushSubscription.ts` |
