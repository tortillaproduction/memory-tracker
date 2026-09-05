import { useAuth } from './hooks/useAuth'
import Dashboard from './pages/Dashboard'
import Login from './pages/Login'

function App() {
  const { isLoading, isAuthenticated } = useAuth()

  return (
    <div className="min-h-screen bg-slate-50">
      {isLoading ? (
        <div className="min-h-screen flex items-center justify-center text-slate-400">
          Loading...
        </div>
      ) : isAuthenticated ? (
        <Dashboard />
      ) : (
        <Login />
      )}
    </div>
  )
}

export default App
