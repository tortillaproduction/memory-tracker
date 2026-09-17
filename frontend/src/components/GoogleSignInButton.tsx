type GoogleSignInButtonProps = {
  /** Called on click — e.g. redirectToGoogleLogin from ../api/auth. */
  onClick: () => void;
  /** Layout-only classes supplied by the caller (width, margin, etc). Do NOT use this to
   *  override the brand colors/sizing below — those are fixed by Google's guidelines. */
  className?: string;
  /** One of Google's approved copy variants (developers.google.com/identity/branding-guidelines). */
  label?: 'Sign in with Google' | 'Sign up with Google' | 'Continue with Google';
};

/**
 * "Sign in with Google" button, permanently fixed to Google's official Dark theme.
 *
 * This intentionally does NOT use daisyUI theme tokens (btn, text-base-content, bg-base-100, ...)
 * because Google's branding guidelines mandate an exact, non-adaptive color scheme. This button
 * must look identical regardless of which of the app's daisyUI themes (see ThemeSwitcher.tsx)
 * is active — do not restyle it to match the page.
 */
export default function GoogleSignInButton({
  onClick,
  className = '',
  label = 'Sign in with Google',
}: GoogleSignInButtonProps) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={[
        'inline-flex items-center justify-center gap-x-[10px]',
        'h-[40px] px-3 rounded-[4px]',
        'border border-solid border-[#8E918F] bg-[#131314] text-[#E3E3E3]',
        'text-[14px] leading-[20px] font-medium tracking-[0.25px]',
        'cursor-pointer select-none transition-colors duration-150',
        'hover:bg-[#1a1a1b] hover:border-[#9aa0a6]',
        'active:bg-[#282a2c]',
        'focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-[#8ab4f8]',
        className,
      ].join(' ')}
      style={{ fontFamily: "'Google Sans', Roboto, Arial, sans-serif" }}
    >
      <span
        className="inline-flex items-center justify-center shrink-0 h-[20px] w-[20px] rounded-[2px] bg-white"
        aria-hidden="true"
      >
        <GoogleGLogo />
      </span>
      <span>{label}</span>
    </button>
  );
}

// Official 4-color Google "G" mark. Path data/colors must not be modified per Google's
// branding guidelines ("you can't change the size or color of the Google 'G' logo").
function GoogleGLogo() {
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
