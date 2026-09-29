import { useState } from 'react';
import { QRCodeSVG } from 'qrcode.react';
import { useToast } from '../contexts/ToastContext';
import { useGateTokenStatus, useIssueGateToken } from '../hooks/useGate';

// セットアップ手順書のURL。未設定ならリンクを表示しない。
const GATE_GUIDE_URL: string | undefined = import.meta.env.VITE_GATE_GUIDE_URL;

type Props = {
  onClose: () => void;
};

function formatDate(iso: string): string {
  return new Date(iso).toLocaleString(undefined, {
    dateStyle: 'medium',
    timeStyle: 'short',
  });
}

// ゲート(Androidアプリ)の接続用QRコードを表示する。
// トークンの平文は発行直後のレスポンスでしか受け取れないため、このモーダルの
// stateにだけ保持し、閉じたら破棄する(再表示には再発行が必要)。
export default function GateSetupModal({ onClose }: Props) {
  const { showToast } = useToast();
  const status = useGateTokenStatus(true);
  const issue = useIssueGateToken();
  const [qrPayload, setQrPayload] = useState<string | null>(null);

  const handleIssue = () => {
    issue.mutate(undefined, {
      onSuccess: (data) => setQrPayload(data.qrPayload),
      onError: () => showToast('Could not create a QR code.', 'error'),
    });
  };

  const handleCopy = async () => {
    if (!qrPayload) return;
    try {
      await navigator.clipboard.writeText(qrPayload);
      showToast('Setup code copied.', 'success');
    } catch {
      showToast('Could not copy.', 'error');
    }
  };

  const renderQrArea = () => {
    if (qrPayload) {
      return (
        <>
          <div className="bg-white p-3 rounded-xl">
            <QRCodeSVG value={qrPayload} size={200} />
          </div>
          <p className="text-xs text-base-content/60 text-center">
            Shown only once. Keep this open until scanned.
          </p>
          {/* QRを読み取れない場合は、この文字列をアプリの「Setup code」に貼り付けて接続できる */}
          <button className="btn btn-ghost btn-xs" onClick={handleCopy}>
            Copy setup code
          </button>
        </>
      );
    }

    if (status.isLoading) {
      return <span className="loading loading-spinner" />;
    }

    if (status.data?.exists) {
      return (
        <>
          <p className="text-sm text-center">
            {status.data.lastUsedAt
              ? `Device connected. Last seen ${formatDate(status.data.lastUsedAt)}.`
              : 'QR code issued. No device connected yet.'}
          </p>
          <button
            className="btn btn-outline btn-sm"
            onClick={handleIssue}
            disabled={issue.isPending}
          >
            {issue.isPending ? (
              <span className="loading loading-spinner loading-sm" />
            ) : (
              'New QR code'
            )}
          </button>
          <p className="text-xs text-base-content/60 text-center">
            The current device will be disconnected.
          </p>
        </>
      );
    }

    return (
      <button
        className="btn btn-primary btn-sm"
        onClick={handleIssue}
        disabled={issue.isPending}
      >
        {issue.isPending ? (
          <span className="loading loading-spinner loading-sm" />
        ) : (
          'Show QR code'
        )}
      </button>
    );
  };

  return (
    <div
      className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 px-4"
      onClick={(e) => {
        if (e.target === e.currentTarget) onClose();
      }}
    >
      <div className="bg-base-100 rounded-2xl shadow-xl w-full max-w-md p-6">
        <h2 className="text-lg font-semibold mb-1">Gate setup</h2>
        <p className="text-sm text-base-content/70 mb-4">
          Android only. Shows overdue sites when you open selected apps.
        </p>

        <ol className="list-decimal list-inside text-sm space-y-1 mb-4">
          <li>Install the Gate app (APK).</li>
          <li>Open the app and scan the QR code.</li>
          <li>Follow the checklist in the app.</li>
        </ol>

        {GATE_GUIDE_URL && (
          <a
            href={GATE_GUIDE_URL}
            target="_blank"
            rel="noopener noreferrer"
            className="link link-primary text-sm"
          >
            Setup guide
          </a>
        )}

        <div className="flex flex-col items-center gap-3 my-6">
          {renderQrArea()}
        </div>

        <div className="flex justify-end">
          <button className="btn btn-ghost" onClick={onClose}>
            Close
          </button>
        </div>
      </div>
    </div>
  );
}
