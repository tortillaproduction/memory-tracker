import { useEffect, type ReactNode } from 'react';
import { Route, Routes } from 'react-router-dom';
import ThemeSwitcher from './components/ThemeSwitcher';
import Footer from './components/Footer';
import { LOGIN_TOAST_FLAG_KEY, LOGOUT_TOAST_FLAG_KEY } from './api/auth';
import { useAuth } from './hooks/useAuth';
import { ToastList, useToast } from './contexts/ToastContext';
import Dashboard from './pages/Dashboard';
import Login from './pages/Login';
import Privacy from './pages/Privacy';
import Terms from './pages/Terms';
import { usePushSubscription } from './hooks/usePushSubscription';
import { useInstallPrompt } from './hooks/useInstallPrompt';

// メニュー項目の先頭に付ける共通アイコン(線画SVG)。
function MenuIcon({ children }: { children: ReactNode }) {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth={2}
      strokeLinecap="round"
      strokeLinejoin="round"
      className="w-4 h-4 shrink-0"
      aria-hidden="true"
    >
      {children}
    </svg>
  );
}

function UserAvatar({
  user,
  className,
}: {
  user: { name: string; pictureUrl?: string };
  className: string;
}) {
  return (
    <div
      className={`${className} aspect-square overflow-hidden rounded-full ring ring-base-300 ring-offset-base-100 ring-offset-1 shrink-0`}
    >
      {user.pictureUrl ? (
        <img
          src={user.pictureUrl}
          alt={user.name}
          referrerPolicy="no-referrer"
        />
      ) : (
        // pictureURLが取得できない場合はイニシャルをフォールバック表示
        <div className="bg-primary text-primary-content flex items-center justify-center w-full h-full text-sm font-bold rounded-full">
          {user.name.charAt(0).toUpperCase()}
        </div>
      )}
    </div>
  );
}

function App() {
  const { user, isLoading, isAuthenticated, logout } = useAuth();
  const { showToast } = useToast();
  const push = usePushSubscription(isAuthenticated);
  const install = useInstallPrompt();

  // ログイン/ログアウトはページ全体のリロードを伴うため、直前にsessionStorageへ
  // 立てておいたフラグをマウント時に確認し、あれば一度だけトースト表示する。
  useEffect(() => {
    try {
      if (sessionStorage.getItem(LOGOUT_TOAST_FLAG_KEY)) {
        sessionStorage.removeItem(LOGOUT_TOAST_FLAG_KEY);
        showToast('Signed out.', 'success');
      }
    } catch {
      // sessionStorageが使えない環境では通知をスキップする
    }
  }, [showToast]);

  useEffect(() => {
    if (!isAuthenticated) return;
    try {
      if (sessionStorage.getItem(LOGIN_TOAST_FLAG_KEY)) {
        sessionStorage.removeItem(LOGIN_TOAST_FLAG_KEY);
        showToast(`Signed in as ${user?.name ?? 'user'}.`, 'success');
      }
    } catch {
      // sessionStorageが使えない環境では通知をスキップする
    }
  }, [isAuthenticated, user, showToast]);

  return (
    <Routes>
      <Route path="/privacy" element={<Privacy />} />
      <Route path="/terms" element={<Terms />} />
      <Route
        path="*"
        element={
          <div
            className={`min-h-screen flex flex-col ${isAuthenticated ? 'bg-base-200' : 'bg-slate-200'}`}
          >
            {isAuthenticated ? (
              <div className="navbar bg-base-100 px-4 shadow-sm relative">
                <div className="flex-1">
                  <span className="text-lg font-bold">Memory Tracker</span>
                </div>

                <div className="flex-none flex items-center gap-3">
                  <ThemeSwitcher />
                  {user && (
                    <div className="dropdown dropdown-end">
                      <div
                        tabIndex={0}
                        role="button"
                        className="avatar cursor-pointer"
                      >
                        <UserAvatar user={user} className="w-8" />
                      </div>
                      <ul
                        tabIndex={0}
                        className="dropdown-content menu menu-sm bg-base-100 rounded-box shadow-lg z-50 mt-3 w-64 p-2"
                      >
                        <li className="pointer-events-none mb-6">
                          <div className="flex items-center gap-3 px-2 py-2">
                            <UserAvatar user={user} className="w-10" />
                            <div className="min-w-0">
                              <div className="font-semibold truncate">
                                {user.name}
                              </div>
                              <div className="text-xs text-base-content/60 truncate">
                                {user.email}
                              </div>
                            </div>
                          </div>
                        </li>

                        <li className="menu-title text-xs px-2.5 py-1 font-normal text-base-content">
                          <span className="flex items-center gap-2">
                            <MenuIcon>
                              <path d="M6 8a6 6 0 0 1 12 0c0 7 3 9 3 9H3s3-2 3-9" />
                              <path d="M10.3 21a1.9 1.9 0 0 0 3.4 0" />
                            </MenuIcon>
                            Notifications
                          </span>
                        </li>
                        <li>
                          <div
                            className={`active:bg-transparent! active:text-inherit! focus:bg-transparent! ${
                              !push.isSupported ? 'tooltip tooltip-left' : ''
                            }`}
                            data-tip={
                              !push.isSupported
                                ? 'Install the app to enable push notifications'
                                : undefined
                            }
                          >
                            <div
                              className="flex items-center justify-center gap-3 px-2 py-1 ml-8"
                              onMouseDown={(e) => e.preventDefault()}
                            >
                              <button
                                type="button"
                                className={`flex items-center gap-1 ${
                                  push.channel === 'email'
                                    ? 'font-semibold'
                                    : 'text-base-content/50'
                                }`}
                                disabled={!push.isSupported}
                                onClick={() => push.selectChannel('email')}
                              >
                                <MenuIcon>
                                  <rect
                                    x="3"
                                    y="5"
                                    width="18"
                                    height="14"
                                    rx="2"
                                  />
                                  <path d="m3 7 9 6 9-6" />
                                </MenuIcon>
                                Email
                              </button>
                              {/* 無効時も丸(つまみ)を塗りつぶして表示し、選択側へ寄せる */}
                              <input
                                type="checkbox"
                                className="toggle toggle-sm toggle-primary [--input-color:var(--color-primary)] disabled:before:bg-primary"
                                checked={push.channel === 'push'}
                                disabled={!push.isSupported}
                                onChange={(e) =>
                                  push.selectChannel(
                                    e.target.checked ? 'push' : 'email',
                                  )
                                }
                              />
                              <button
                                type="button"
                                className={`flex items-center gap-1 ${
                                  push.channel === 'push'
                                    ? 'font-semibold'
                                    : 'text-base-content/50'
                                }`}
                                disabled={!push.isSupported}
                                onClick={() => push.selectChannel('push')}
                              >
                                <MenuIcon>
                                  <rect
                                    x="7"
                                    y="2"
                                    width="10"
                                    height="20"
                                    rx="2"
                                  />
                                  <path d="M11 18h2" />
                                </MenuIcon>
                                Push
                              </button>
                            </div>
                          </div>
                        </li>

                        <div className="my-2 h-px bg-base-content/10"></div>
                        {!install.isInstalled && install.canPrompt && (
                          <li>
                            <button onClick={() => install.promptInstall()}>
                              <MenuIcon>
                                <path d="M12 3v12" />
                                <path d="M7 10l5 5 5-5" />
                                <path d="M5 21h14" />
                              </MenuIcon>
                              Install app
                            </button>
                          </li>
                        )}

                        <div className="my-2 h-px bg-base-content/10"></div>
                        <li>
                          <button onClick={() => logout()}>
                            <MenuIcon>
                              <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4" />
                              <path d="m16 17 5-5-5-5" />
                              <path d="M21 12H9" />
                            </MenuIcon>
                            logout
                          </button>
                        </li>
                      </ul>
                    </div>
                  )}
                </div>

                {/* ナビゲーションバー中央にトーストを表示 */}
                <ToastList variant="navbar" />
              </div>
            ) : (
              // ナビゲーションバーが無い画面(ログイン/ロード中)向けのフォールバック表示
              <ToastList variant="fixed" />
            )}

            <div className="flex-1">
              {isLoading ? (
                <div className="min-h-[80vh] flex items-center justify-center">
                  <span className="loading loading-spinner loading-lg text-primary"></span>
                </div>
              ) : isAuthenticated ? (
                <Dashboard />
              ) : (
                <Login />
              )}
            </div>

            {!isLoading && !isAuthenticated && <Footer />}
          </div>
        }
      />
    </Routes>
  );
}

export default App;
