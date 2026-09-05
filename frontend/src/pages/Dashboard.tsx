import { useAuth } from '../hooks/useAuth'

// NOTE: サイト一覧・ストリークはまだ /api/sites 等のAPIが無いためダミーデータを使用。
// バックエンド側のエンドポイントが揃い次第、useAuthと同様にReact Queryのフックへ置き換える想定。


type SiteStatus = {
  id: string
  name: string
  hoursSinceLastCheckIn: number
  intervalHours: number
  streakDays: number
}

const dummySites: SiteStatus[] = [
  { id: '1', name: 'Progate', hoursSinceLastCheckIn: 3, intervalHours: 24, streakDays: 12 },
  { id: '2', name: 'Udemy - Go入門', hoursSinceLastCheckIn: 20, intervalHours: 24, streakDays: 2 },
  { id: '3', name: 'AtCoder', hoursSinceLastCheckIn: 30, intervalHours: 24, streakDays: 0 },
]

function statusColor(hoursSince: number, interval: number): string {
  const ratio = hoursSince / interval
  if (ratio < 0.5) return 'bg-green-100 border-green-400 text-green-800'
  if (ratio < 1) return 'bg-yellow-100 border-yellow-400 text-yellow-800'
  return 'bg-red-100 border-red-400 text-red-800'
}

export default function Dashboard() {
  const { user, logout } = useAuth()
  const userStreak = 12 // TODO: APIから取得

  return (
    <div className="max-w-2xl mx-auto px-4 py-8">
      <div className="flex items-center justify-between mb-4">
        <p>Logged in as {user?.name}</p>
        <button
          onClick={() => logout()}
          className="text-sm text-slate-400 hover:text-slate-600"
        >
          Logout
        </button>
      </div>

      <header className="mb-8 text-center">
        <h1 className="text-2xl font-bold text-slate-800">Study Tracker</h1>
        <p className="mt-2 text-4xl">🔥 {userStreak}日連続</p>
      </header>

      <div className="space-y-3">
        {dummySites.map((site) => (
          <div
            key={site.id}
            className={`rounded-lg border-2 p-4 flex items-center justify-between ${statusColor(
              site.hoursSinceLastCheckIn,
              site.intervalHours,
            )}`}
          >
            <div>
              <p className="font-semibold">{site.name}</p>
              <p className="text-sm opacity-80">
                最終訪問から{site.hoursSinceLastCheckIn}時間 ・ 連続{site.streakDays}日
              </p>
            </div>
            <a
              href={`/go/${site.id}`}
              className="px-4 py-2 rounded-md bg-slate-800 text-white text-sm font-medium hover:bg-slate-700"
            >
              開く
            </a>
          </div>
        ))}
      </div>

      <button className="mt-6 w-full py-3 rounded-lg border-2 border-dashed border-slate-300 text-slate-500 hover:border-slate-400">
        + サイトを追加
      </button>
    </div>
  )
}
