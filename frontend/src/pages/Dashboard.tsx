import { useState } from 'react';
import { useAuth } from '../hooks/useAuth';
import { useDeleteSite, useSites } from '../hooks/useSites';
import { API_BASE_URL } from '../api/client';
import AddSiteModal from '../components/AddSiteModal';

function statusBadgeClass(hoursSince: number, interval: number): string {
  const ratio = hoursSince / interval;
  if (ratio < 0.5) return 'badge-success';
  if (ratio < 1) return 'badge-warning';
  return 'badge-error';
}

export default function Dashboard() {
  const { user } = useAuth();
  const { data, isLoading, isError } = useSites();
  const { mutate: deleteSite } = useDeleteSite();
  const [showAddModal, setShowAddModal] = useState(false);

  const sites = data?.sites ?? [];
  const userStreak = data?.userStreak ?? 0;

  return (
    <div className="max-w-2xl mx-auto px-4 py-8">
      <div className="text-center mb-8">
        <p className="text-4xl font-extrabold text-slate-700">
          {' '}
          🔥{userStreak} {userStreak === 1 ? 'day' : 'days'} streak
        </p>
        {user?.planType == 'free' && (
          <p className="mt-2 text-xs text-slate-400">
            Free Plan - {sites.length}/{user.maxSites} sites registered.
          </p>
        )}
      </div>

      {isLoading && (
        <div className="flex justify-center py-12">
          <span className="loading loading-spinner loading-lg text-primary" />
        </div>
      )}

      {isError && (
        <div className="alert alert-error shadow-lg">
          <div>
            <svg
              xmlns="http://www.w3.org/2000/svg"
              className="stroke-current shrink-0 h-6 w-6"
              fill="none"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth="2"
                d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z"
              />
            </svg>
            <span>Failed to load sites. Please try again.</span>
          </div>
        </div>
      )}

      {!isLoading && !isError && sites.length === 0 && (
        <div className="text-center py-12 text-base-content/50">
          <p className="text-lg mb-2">No sites registered.</p>
          <p className="text-sm">
            Click{' '}
            <span className="font-bold underline decoration-indigo-500 underline-offset-2300">
              + Add Site
            </span>{' '}
            to register your favorite sites.
          </p>
        </div>
      )}

      <div className="space-y-3">
        {sites.map((site) => (
          <div key={site.id} className="card bg-base-100 shadow-sm">
            <div className="card-body flex-row items-center justify-between py-4">
              <div className="min-w-0 flex-1">
                <p className="font-semibold truncate">{site.name}</p>
                <div className="flex items-center gap-2 mt-1 flex-wrap">
                  <span
                    className={`badge badge-sm ${statusBadgeClass(site.hoursSinceLastCheck, site.intervalHours)}`}
                  >
                    {site.hoursSinceLastCheck < 1
                      ? 'Less than 1 hour'
                      : `${Math.floor(site.hoursSinceLastCheck)} hours passed`}
                  </span>
                  <span className="text-xs text-base-content/60">
                    {site.siteStreak} {site.siteStreak === 1 ? 'day' : 'days'} streak
                  </span>
                </div>
              </div>

              <div className="flex items-center gap-2 ml-3 shrink-0">
                {/* 経由リンク方式でチェックイン & リダイレクト */}
                <a
                  href={`${API_BASE_URL}/go/${site.id}`}
                  className="btn btn-primary btn-sm"
                  target="_blank"
                  rel="noopener noreferrer"
                >
                  open
                </a>
                <button
                  className="btn btn-sm btn-ghost text-error"
                  onClick={() => {
                    if (
                      confirm(`Are you sure you want to delete ${site.name}?`)
                    ) {
                      deleteSite(site.id);
                    }
                  }}
                >
                  delete
                </button>
              </div>
            </div>
          </div>
        ))}
      </div>

      <button
        className="btn w-full mt-6 border-dashed btn-block hover:border-base-content font-bold"
        onClick={() => setShowAddModal(true)}
      >
        + Add Site
      </button>

      {showAddModal && <AddSiteModal onClose={() => setShowAddModal(false)} />}
    </div>
  );
}
