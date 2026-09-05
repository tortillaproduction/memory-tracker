import { redirectToGoogleLogin } from '../api/auth';

export default function Login() {
    return (
        <div className="min-h-screen flex items-center justify-center px-4">
            <div className="max-w-sm w-full text-center">
                <h1 className="text-2xl font-bold text-slate-800 mb-2">Memory Tracker</h1>
                <p className="text-slate-500 mb-8">before forget</p>
                <button
                    onClick={redirectToGoogleLogin}
                    className="w-full flex items-center justify-center gap-2 rounded-lg border border-slate-300 bg-white px-4 py-3 font-medium text-slate-700 hover:bg-slate-50 shadow-sm"
                >
                    Sign in with Google
                </button>
            </div>
        </div>
    )
}