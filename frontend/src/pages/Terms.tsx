import { Link } from 'react-router-dom';

export default function Terms() {
  return (
    <div className="max-w-2xl mx-auto px-4 py-10 bg-base-100 text-base-content min-h-screen">
      <Link to="/" className="link link-hover text-sm text-primary">
        &larr; トップへ戻る
      </Link>

      <h1 className="text-2xl font-bold mt-6 mb-2">利用規約</h1>
      <p className="text-sm text-base-content/60 mb-8">最終改定日: 2026年9月17日</p>

      <div className="space-y-8 text-sm leading-relaxed">
        <section>
          <h2 className="text-lg font-bold mb-2">1. 適用</h2>
          <p>
            この利用規約（以下「本規約」といいます）は、Memory Tracker（以下「本サービス」といいます）の利用条件を定めるものです。利用者は、本サービスを利用することにより、本規約に同意したものとみなされます。
          </p>
        </section>

        <section>
          <h2 className="text-lg font-bold mb-2">2. サービス内容</h2>
          <p>
            本サービスは、利用者が登録したウェブサイト・サービスについて、設定した確認間隔を過ぎた際にリマインドメールを送信する機能を提供します。本サービスの内容は、予告なく変更、追加、または終了されることがあります。
          </p>
        </section>

        <section>
          <h2 className="text-lg font-bold mb-2">3. アカウント登録</h2>
          <p>
            本サービスの利用にはGoogleアカウントによるログインが必要です。利用者は、自己の責任においてアカウントを管理するものとし、第三者による不正利用について本サービスは責任を負いません。
          </p>
        </section>

        <section>
          <h2 className="text-lg font-bold mb-2">4. 禁止事項</h2>
          <ul className="list-disc list-inside space-y-1">
            <li>法令または公序良俗に違反する行為</li>
            <li>本サービスの運営を妨害する行為</li>
            <li>他の利用者または第三者の権利を侵害する行為</li>
            <li>不正アクセスやリバースエンジニアリング等、本サービスの脆弱性を探る行為</li>
          </ul>
        </section>

        <section>
          <h2 className="text-lg font-bold mb-2">5. 免責事項</h2>
          <p>
            本サービスは現状有姿で提供されるものとし、リマインドメールの到達・送信タイミングを含め、その完全性・正確性・有用性についていかなる保証も行いません。本サービスの利用により利用者に生じた損害について、運営者は故意または重過失がある場合を除き、責任を負わないものとします。
          </p>
        </section>

        <section>
          <h2 className="text-lg font-bold mb-2">6. サービス内容の変更・停止</h2>
          <p>
            運営者は、利用者への事前の通知なく、本サービスの全部または一部の提供を変更、中断、または終了できるものとします。これにより利用者に生じた損害について、運営者は責任を負いません。
          </p>
        </section>

        <section>
          <h2 className="text-lg font-bold mb-2">7. 準拠法・管轄裁判所</h2>
          <p>
            本規約の解釈にあたっては、日本法を準拠法とします。本サービスに関して紛争が生じた場合には、運営者の所在地を管轄する裁判所を専属的合意管轄とします。
          </p>
        </section>

        <section>
          <h2 className="text-lg font-bold mb-2">8. お問い合わせ</h2>
          <p>
            本規約に関するお問い合わせは、リマインドメールへの返信、または本サービスのGitHubリポジトリのIssueにてご連絡ください。
          </p>
        </section>
      </div>
    </div>
  );
}
