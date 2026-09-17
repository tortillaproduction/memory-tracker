import { Link } from 'react-router-dom';

export default function Footer() {
  return (
    <footer className="py-6 text-center text-xs text-slate-400">
      <Link to="/privacy" className="hover:text-slate-600 hover:underline">
        Privacy Policy
      </Link>
      <span className="mx-2">·</span>
      <Link to="/terms" className="hover:text-slate-600 hover:underline">
        Terms of Service
      </Link>
    </footer>
  );
}
