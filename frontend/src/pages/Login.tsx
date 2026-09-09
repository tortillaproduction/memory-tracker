import { redirectToGoogleLogin } from '../api/auth';

// Google公式のブランドガイドラインに沿った4色の「G」ロゴ。
// 「Googleでログイン」ボタンに添える用途は公式に想定された使い方。
function GoogleIcon() {
  return (
    <svg width="18" height="18" viewBox="0 0 18 18" aria-hidden="true">
      <path
        fill="#4285F4"
        d="M17.64 9.2c0-.637-.057-1.251-.164-1.84H9v3.481h4.844a4.14 4.14 0 0 1-1.796 2.716v2.259h2.908c1.702-1.567 2.684-3.875 2.684-6.615z"
      />
      <path
        fill="#34A853"
        d="M9 18c2.43 0 4.467-.806 5.956-2.184l-2.908-2.259c-.806.54-1.837.86-3.048.86-2.344 0-4.328-1.584-5.036-3.711H.957v2.332A8.997 8.997 0 0 0 9 18z"
      />
      <path
        fill="#FBBC05"
        d="M3.964 10.706A5.41 5.41 0 0 1 3.682 9c0-.593.102-1.17.282-1.706V4.962H.957A8.996 8.996 0 0 0 0 9c0 1.452.348 2.827.957 4.038l3.007-2.332z"
      />
      <path
        fill="#EA4335"
        d="M9 3.58c1.321 0 2.508.454 3.44 1.345l2.582-2.58C13.463.891 11.426 0 9 0A8.997 8.997 0 0 0 .957 4.962L3.964 7.294C4.672 5.167 6.656 3.58 9 3.58z"
      />
    </svg>
  );
}

export default function Login() {
  return (
    <div className="min-h-[80vh] flex items-center justify-center px-4">
      <div className="card w-full max-w-sm bg-base-100 shadow-xl">
        <div className="card-body items-center text-center">
          <h2
            className="text-5xl font-extrabold tracking-wide mb-10"
            style={{
              color: '#ffffff',
              textShadow: [
                '0 0 2px #ffffff',
                '0 0 4px #ffffff',
                '0 0 8px #c4b5fd',
                '0 0 16px #a78bfa',
                '0 0 32px #7c3aed',
                '0 0 60px #6d28d9',
                '0 0 90px #5b21b6',
                '0 0 4px #f0abfc',
              ].join(', '),
            }}
          >
            Memory Tracker
          </h2>
          <p className="text-base-content/60 mb-4">before forget</p>
          <button
            onClick={redirectToGoogleLogin}
            className="btn btn-outline text-base-content border-base-content/30 hover:border-base-content w-full gap-2"
          >
            <GoogleIcon />
            Sign in with Google
          </button>
        </div>
      </div>
    </div>
  );
}
