import { useRef, useState } from 'react';
import { useRegisterSite } from '../hooks/useSites';
import { ApiError } from '../api/client';
import { useToast } from '../contexts/ToastContext';

type Props = {
  onClose: () => void;
};

const INTERVAL_OPTIONS = [
  { label: '12 hour', value: 12 },
  { label: '24 hour (default)', value: 24 },
  { label: '48 hour', value: 48 },
  { label: '72 hour', value: 72 },
  { label: '1 week', value: 168 },
];

export default function AddSiteModal({ onClose }: Props) {
  const [name, setName] = useState('');
  const [url, setUrl] = useState('');
  const [intervalHours, setIntervalHours] = useState(24);
  const [errorMsg, setErrorMsg] = useState('');
  const backdropRef = useRef<HTMLDivElement>(null);

  const { mutate: register, isPending } = useRegisterSite();
  const { showToast } = useToast();

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setErrorMsg('');

    // URLの簡易バリデーション
    try {
      new URL(url);
    } catch {
      setErrorMsg('Please enter a valid URL (e.g., https://example.com)');
      return;
    }

    const trimmedName = name.trim();

    register(
      { name: trimmedName, url: url.trim(), intervalHours },
      {
        onSuccess: () => {
          showToast(`"${trimmedName}" added.`, 'success');
          onClose();
        },
        onError: (err) => {
          // サーバー側のエラーはトーストで通知し、モーダルは開いたままにして再入力できるようにする
          if (err instanceof ApiError && err.status === 402) {
            showToast(
              'You have reached the maximum number of sites. Please upgrade your plan to add more sites.',
              'error',
            );
          } else {
            showToast(
              'Failed to register site. Please try again later.',
              'error',
            );
          }
        },
      },
    );
  }

  // 背景クリックで閉じる
  function handleBackdropClick(e: React.MouseEvent) {
    if (e.target === backdropRef.current) onClose();
  }

  return (
    <div
      ref={backdropRef}
      onClick={handleBackdropClick}
      className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 px-4"
    >
      <div className="bg-base-100 rounded-2xl shadow-xl w-full max-w-md p-6">
        <h2 className="text-lg font-semibold mb-4">Add Site</h2>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="form-control">
            <label className="label">
              <span className="label-text">Site Name</span>
            </label>
            <input
              type="text"
              className="input input-bordered w-full"
              placeholder="Enter site name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              required
              autoFocus
            />
          </div>

          <div className="form-control">
            <label className="label">
              <span className="label-text">URL</span>
            </label>
            <input
              type="url"
              className="input input-bordered w-full"
              placeholder="https://example.com"
              value={url}
              onChange={(e) => setUrl(e.target.value)}
              required
            />
          </div>

          <div className="form-control">
            <label className="label">
              <span className="label-text">Check Interval</span>
            </label>
            <select
              className="select select-bordered w-full"
              value={intervalHours}
              onChange={(e) => setIntervalHours(Number(e.target.value))}
            >
              {INTERVAL_OPTIONS.map((opt) => (
                <option key={opt.value} value={opt.value}>
                  {opt.label}
                </option>
              ))}
            </select>
          </div>

          {errorMsg && (
            <div className="alert alert-error text-sm py-2">
              <span>{errorMsg}</span>
            </div>
          )}

          <div className="flex gap-2 justify-end pt-2">
            <button type="button" onClick={onClose} className="btn btn-ghost">
              Cancel
            </button>
            <button
              type="submit"
              className="btn btn-primary"
              disabled={isPending}
            >
              {isPending ? <span className="loading loading-spinner" /> : 'Add'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
