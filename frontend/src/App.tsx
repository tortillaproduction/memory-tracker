import { useEffect } from 'react';
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
                        <div className="w-8 rounded-full ring ring-base-300 ring-offset-base-100 ring-offset-1">
                          {user.pictureUrl ? (
                            <img
                              src={user.pictureUrl}
                              alt={user.name}
                              referrerPolicy="no-referrer"
                            />
                          ) : (
                            // pictureURLが取得できない場合はイニシャルをフォールバック表示
                            <div className="bg-primary text-primary-content flex items-center justify-center w-full h-full text-sm font-bold">
                              {user.name.charAt(0).toUpperCase()}
                            </div>
                          )}
                        </div>
                      </div>
                      <ul
                        tabIndex={0}
                        className="dropdown-content menu menu-sm bg-base-100 rounded-box shadow-lg z-50 mt-3 w-48 p-2"
                      >
                        <li className="menu-title text-xs px-2 py-1 truncate">
                          {user.name}
                        </li>
                        <li className="menu-title text-xs px-2 py-1">
                          Notifications
                        </li>
                        <li>
                          <button
                            onClick={() => push.subscribe()}
                            disabled={push.isBusy}
                          >
                            {push.isSubscribed
                              ? 'Browser notifications: on'
                              : 'Enable browser notifications'}
                          </button>
                        </li>
                        {push.isSubscribed && (
                          <>
                            <li>
                              <button
                                onClick={() => push.unsubscribe()}
                                disabled={push.isBusy}
                              >
                                Turn off browser notifications
                              </button>
                            </li>
                            <li>
                              <label className="flex items-center justify-between gap-2 cursor-pointer">
                                <span>
                                  Turn off email when push is available
                                </span>
                                <input
                                  type="checkbox"
                                  className="toggle toggle-sm toggle-primary"
                                  checked={
                                    push.preferences
                                      ?.disableEmailWhenPushAvailable ?? true
                                  }
                                  onChange={(e) =>
                                    push.setDisableEmailWhenPushAvailable(
                                      e.target.checked,
                                    )
                                  }
                                />
                              </label>
                            </li>
                          </>
                        )}
                        {!install.isInstalled && install.canPrompt && (
                          <li>
                            <button onClick={() => install.promptInstall()}>
                              Install app
                            </button>
                          </li>
                        )}
                        <li>
                          <button onClick={() => logout()}>logout</button>
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
