import { Link } from 'react-router-dom';

export default function Privacy() {
  return (
    <div className="max-w-2xl mx-auto px-4 py-10 bg-base-100 text-base-content min-h-screen">
      <Link to="/" className="link link-hover text-sm text-primary">
        &larr; トップへ戻る
      </Link>

      <h1 className="text-2xl font-bold mt-6 mb-2">プライバシーポリシー</h1>
      <p className="text-sm text-base-content/60 mb-8">最終改定日: 2026年9月17日</p>

      <div className="space-y-8 text-sm leading-relaxed">
        <section>
          <h2 className="text-lg font-bold mb-2">1. はじめに</h2>
          <p>
            Memory Tracker（以下「本サービス」といいます）は、利用者のプライバシーを尊重し、取得した個人情報を適切に取り扱います。本ポリシーは、本サービスにおける個人情報の取り扱いについて定めるものです。
          </p>
        </section>

        <section>
          <h2 className="text-lg font-bold mb-2">2. 取得する情報</h2>
          <p className="mb-2">本サービスは、Googleアカウントによるログイン時に、Google OAuthを通じて以下の情報を取得します。</p>
          <ul className="list-disc list-inside space-y-1">
            <li>氏名（表示名）</li>
            <li>メールアドレス</li>
            <li>プロフィール画像（アイコン）</li>
          </ul>
          <p className="mt-2">
            また、利用者が本サービスに登録したサイトの情報（サイト名、URL、確認間隔）および訪問（チェックイン）履歴を取得・保存します。
          </p>
        </section>

        <section>
          <h2 className="text-lg font-bold mb-2">3. 利用目的</h2>
          <ul className="list-disc list-inside space-y-1">
            <li>本人確認およびログイン状態の維持</li>
            <li>登録されたサイトの確認間隔が過ぎた際のリマインドメールの送信</li>
            <li>サービスの提供、維持、改善</li>
            <li>不正利用の防止</li>
          </ul>
        </section>

        <section>
          <h2 className="text-lg font-bold mb-2">4. Cookie・セッション情報の利用</h2>
          <p>
            本サービスは、ログイン状態を維持するためにCookieを利用したセッション管理を行います。これらの情報は本人確認の目的以外には利用しません。
          </p>
        </section>

        <section>
          <h2 className="text-lg font-bold mb-2">5. 第三者提供・委託先</h2>
          <p className="mb-2">
            本サービスは、以下の外部サービスを利用しており、必要な範囲で情報を送信します。
          </p>
          <ul className="list-disc list-inside space-y-1">
            <li>Google LLC（Googleアカウントによる認証のため）</li>
            <li>Resend（登録メールアドレス宛のリマインドメール配信のため）</li>
          </ul>
          <p className="mt-2">
            法令に基づく場合を除き、取得した個人情報を本人の同意なく上記以外の第三者に提供することはありません。
          </p>
        </section>

        <section>
          <h2 className="text-lg font-bold mb-2">6. 情報の保管・削除</h2>
          <p>
            利用者情報は、本サービスの提供に必要な期間保管します。アカウントの削除をご希望の場合は、下記のお問い合わせ先までご連絡ください。
          </p>
        </section>

        <section>
          <h2 className="text-lg font-bold mb-2">7. お問い合わせ</h2>
          <p>
            本ポリシーに関するお問い合わせは、リマインドメールへの返信、または本サービスのGitHubリポジトリのIssueにてご連絡ください。
          </p>
        </section>

        <section>
          <h2 className="text-lg font-bold mb-2">8. 改定</h2>
          <p>
            本ポリシーの内容は、必要に応じて予告なく変更されることがあります。変更後のポリシーは、本ページに掲載した時点から効力を生じるものとします。
          </p>
        </section>
      </div>
    </div>
  );
}
