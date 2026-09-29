# 通知の運用ガイド（リトライ・ログ・切り分け）

通知バッチ（メール通知）の失敗時の挙動と、届かないときの調査方法をまとめた運用資料。
Resendなどのセットアップは[README](../README.md)を参照。

## 通知バッチの基本

- `NOTIFICATION_INTERVAL`（デフォルト5分）ごとに、期限切れかつ未通知のサイトをユーザー単位で1通のメールにまとめて通知する（`backend/internal/usecase/notify_overdue_sites`）。
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

### 代表的なエラー

| ログ | 原因と対処 |
|---|---|
| `resend send: You can only send testing emails to your own email address ...` | Resendでドメイン未認証のまま、`from`が`onboarding@resend.dev`等のまま。resend.com/domainsでドメインを認証し、`EMAIL_FROM_ADDRESS`をそのドメインのアドレスにする。 |

## ログの見方

| レベル | メッセージ | 意味 |
|---|---|---|
| Info | `notification sent` | メールを送信し、`notification_logs`に記録した（`siteCount`、`channel`付き）。 |
| Error | `failed to send notification` | 送信に失敗した。`error`に原因、`retryIn`に次の再試行までの待ち時間が入る。 |

## 既知の限界

- バックオフの状態はメモリ上のみ。再起動するとリセットされる。
- `RESEND_API_KEY`が未設定の場合、メール送信は何もしない実装になるが、送信成功として`notification_logs`に記録される。

## 関連ファイル

| 役割 | ファイル |
|---|---|
| バッチ・バックオフ | `backend/internal/usecase/notify_overdue_sites/notify_overdue_sites.go` |
| メール本文 | `backend/internal/usecase/notify_overdue_sites/email_template.go` |
| メール送信 | `backend/internal/infrastructure/notification/email_sender.go` |
