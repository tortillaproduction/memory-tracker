type Props = {
  siteName: string;
  onConfirm: () => void;
  onClose: () => void;
  isPending: boolean;
};

export default function DeleteSiteModal({
  siteName,
  onConfirm,
  onClose,
  isPending,
}: Props) {
  return (
    <div
      className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 px-4"
      onClick={(e) => {
        if (e.target === e.currentTarget) onClose();
      }}
    >
      <div className="bg-base-100 rounded-2xl shadow-xl w-full max-w-sm p-6">
        <h2 className="text-lg font-bold mb-2">delete</h2>
        <p className="text-base-content/70 mb-6">
          Are you sure you want to delete?
          <br />
          <span className="font-bold text-base-content">{siteName}</span>
        </p>
        <div className="flex justify-end gap-2">
          <button
            className="btn btn-ghost"
            onClick={onClose}
            disabled={isPending}
          >
            cancel
          </button>
          <button
            className="btn btn-error"
            onClick={onConfirm}
            disabled={isPending}
          >
            {isPending ? (
              <span className="loading loading-spinner loading-sm" />
            ) : (
              'delete'
            )}
          </button>
        </div>
      </div>
    </div>
  );
}
