import ThemeSwitcher from './components/ThemeSwitcher';
import { useAuth } from './hooks/useAuth';
import Dashboard from './pages/Dashboard';
import Login from './pages/Login';

function App() {
  const { user, isLoading, isAuthenticated, logout } = useAuth();

  return (
    <div className="min-h-screen bg-slate-50">
      {isAuthenticated && (
        <div className="navbar bg-base-100 px-4 shadow-sm">
          <div className="flex-1">
            <span className="text-lg font-bold">Memory Tracker</span>
          </div>

          <div className="flex-none flex items-center gap-3">
            <ThemeSwitcher />
            {user && (
              <div className="dropdown dropdown-end">
                <div tabIndex={0} role="button" className="avatar cursor-pointer">
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
                  <li>
                    <button onClick={() => logout()}>logout</button>
                  </li>
                </ul>
              </div>
            )}
          </div>
        </div>
      )}

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
  );
}

export default App;
