import { redirectToGoogleLogin } from '../api/auth';
import GoogleSignInButton from '../components/GoogleSignInButton';

export default function Login() {
  return (
    <div className="min-h-[90vh] flex items-center justify-center px-4">
      <div className="card w-full max-w-md aspect-square bg-base-100 shadow-xl">
        <div className="card-body items-center justify-center text-center h-full">
          <h2
            className="text-7xl font-extrabold tracking-wide mb-10 mt-15"
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
          <p
            className="text-base-content/60 text-2xl"
            style={{ textShadow: '1px 1px 2px rgba(0, 0, 0, 0.6)' }}
          >
            before forget
          </p>
          <GoogleSignInButton onClick={redirectToGoogleLogin} className="mb-5" />
        </div>
      </div>
    </div>
  );
}
