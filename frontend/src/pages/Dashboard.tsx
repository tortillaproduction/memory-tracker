import { useAuth } from '../hooks/useAuth';

// NOTE: サイト一覧・ストリークはまだ /api/sites 等のAPIが無いためダミーデータを使用。
// バックエンド側のエンドポイントが揃い次第、useAuthと同様にReact Queryのフックへ置き換える想定。

type SiteStatus = {
  id: string;
  name: string;
  hoursSinceLastCheckIn: number;
  intervalHours: number;
  streakDays: number;
};

const dummySites: SiteStatus[] = [
  {
    id: '1',
    name: 'Progate',
    hoursSinceLastCheckIn: 3,
    intervalHours: 24,
    streakDays: 12,
  },
  {
    id: '2',
    name: 'Udemy - Go入門',
    hoursSinceLastCheckIn: 20,
    intervalHours: 24,
    streakDays: 2,
  },
  {
    id: '3',
    name: 'AtCoder',
    hoursSinceLastCheckIn: 30,
    intervalHours: 24,
    streakDays: 0,
  },
];

function statusBadgeClass(hoursSince: number, interval: number): string {
  const ratio = hoursSince / interval;
  if (ratio < 0.5) return 'badge-success';
  if (ratio < 1) return 'badge-warning';
  return 'badge-error';
}

export default function Dashboard() {
  const { user } = useAuth();
  const userStreak = 12; // TODO: /api/streaks 実装後に置き換え

  return (
    <div className="max-w-2xl mx-auto px-4 py-8">
      <div className="text-center mb-8">
        <p className="text-4xl font-extrabold text-slate-700">
          {' '}
          🔥{userStreak} days streak
        </p>
        {user?.planType == 'free' && (
          <p className="mt-2 text-xs text-slate-400">
            Free Plan - Add to {user.maxSites} sites. Upgrade to premium for
            unlimited sites.
          </p>
        )}
      </div>

      <div className="space-y-3">
        {dummySites.map((site) => (
          <div key={site.id} className="card bg-base-100 shadow-sm">
            <div className="card-body flex-row items-center justify-between py-4">
              <div>
                <p className="font-semibold">{site.name}</p>
                <div className="flex items-center gap-2 mt-1">
                  <span
                    className={`badge badge-sm ${statusBadgeClass(site.hoursSinceLastCheckIn, site.intervalHours)}`}
                  >
                    {site.hoursSinceLastCheckIn} hours passed since last
                    check-in
                  </span>
                  <span className="text-xs text-base-content/60">
                    {site.streakDays} days streak
                  </span>
                </div>
              </div>
              <a href={`/go/${site.id}`} className="btn btn-primary btn-sm">
                open
              </a>
            </div>
          </div>
        ))}
      </div>

      <button className="btn w-full mt-6 border-dashed btn-block hover:border-base-content">
        + Add Site
      </button>
    </div>
  );
}
