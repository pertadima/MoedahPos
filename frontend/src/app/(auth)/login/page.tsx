'use client';

import { useState, FormEvent } from 'react';
import { useRouter } from 'next/navigation';
import Image from 'next/image';
import { Mail, Lock, Eye, EyeOff, Loader2, Sun, Moon } from 'lucide-react';
import { useAuth } from '@/lib/auth/AuthContext';
import { useTheme } from '@/lib/theme/ThemeContext';
import { ApiError } from '@/lib/api/client';

export default function LoginPage() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [showPw, setShowPw] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const { login } = useAuth();
  const { isDark, toggleTheme } = useTheme();
  const router = useRouter();

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);
    try {
      await login(email, password);
      router.push('/dashboard');
    } catch (err) {
      if (err instanceof ApiError) setError(err.message);
      else setError('Terjadi kesalahan. Coba lagi.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div style={{ width: '100%', maxWidth: 400 }}>
      {/* Theme toggle — top right of card area */}
      <div style={{ display: 'flex', justifyContent: 'flex-end', marginBottom: 8 }}>
        <button
          onClick={toggleTheme}
          title={isDark ? 'Mode terang' : 'Mode gelap'}
          style={{
            background: 'var(--bg-elevated)',
            border: '1px solid var(--border-md)',
            borderRadius: 8,
            padding: '6px 10px',
            cursor: 'pointer',
            color: 'var(--text-2)',
            display: 'flex',
            alignItems: 'center',
            gap: 6,
            fontSize: '0.78rem',
            fontWeight: 500,
          }}
        >
          {isDark ? <Sun size={14} /> : <Moon size={14} />}
          {isDark ? 'Terang' : 'Gelap'}
        </button>
      </div>

      {/* Logo */}
      <div style={{ textAlign: 'center', marginBottom: 28 }}>
        <div
          style={{
            width: 72,
            height: 72,
            borderRadius: 20,
            margin: '0 auto 16px',
            overflow: 'hidden',
            boxShadow: '0 0 40px rgba(8,132,246,0.30), 0 0 80px rgba(8,132,246,0.12)',
            background: isDark ? 'transparent' : 'white',
          }}
        >
          <Image
            src={isDark ? '/logo-icon-dark.svg' : '/logo-icon-light.svg'}
            alt="Moedah"
            width={72}
            height={72}
            style={{ display: 'block', width: '100%', height: '100%' }}
            priority
          />
        </div>
        <h1
          style={{
            fontSize: '1.6rem',
            fontWeight: 800,
            color: 'var(--brand)',
            letterSpacing: '-0.5px',
          }}
        >
          Moedah POS
        </h1>
        <p style={{ color: 'var(--text-2)', marginTop: 4, fontSize: '0.9rem' }}>
          Masuk ke akun Anda
        </p>
      </div>

      {/* Card */}
      <div className="card" style={{ padding: '28px 28px 24px' }}>
        {error && (
          <div
            style={{
              background: 'rgba(239,68,68,0.12)',
              border: '1px solid rgba(239,68,68,0.3)',
              borderRadius: 8,
              padding: '10px 14px',
              marginBottom: 18,
              color: '#f87171',
              fontSize: '0.85rem',
            }}
          >
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit} style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
          {/* Email */}
          <div className="input-group">
            <label className="input-label">Email</label>
            <div style={{ position: 'relative' }}>
              <Mail
                size={15}
                style={{
                  position: 'absolute',
                  left: 12,
                  top: '50%',
                  transform: 'translateY(-50%)',
                  color: 'var(--text-3)',
                }}
              />
              <input
                type="email"
                required
                value={email}
                onChange={e => setEmail(e.target.value)}
                placeholder="nama@email.com"
                className="input"
                style={{ paddingLeft: 36 }}
              />
            </div>
          </div>

          {/* Password */}
          <div className="input-group">
            <label className="input-label">Password</label>
            <div style={{ position: 'relative' }}>
              <Lock
                size={15}
                style={{
                  position: 'absolute',
                  left: 12,
                  top: '50%',
                  transform: 'translateY(-50%)',
                  color: 'var(--text-3)',
                }}
              />
              <input
                type={showPw ? 'text' : 'password'}
                required
                value={password}
                onChange={e => setPassword(e.target.value)}
                placeholder="Password Anda"
                className="input"
                style={{ paddingLeft: 36, paddingRight: 40 }}
              />
              <button
                type="button"
                onClick={() => setShowPw(!showPw)}
                style={{
                  position: 'absolute',
                  right: 10,
                  top: '50%',
                  transform: 'translateY(-50%)',
                  background: 'none',
                  border: 'none',
                  cursor: 'pointer',
                  color: 'var(--text-3)',
                  padding: 2,
                }}
              >
                {showPw ? <EyeOff size={15} /> : <Eye size={15} />}
              </button>
            </div>
          </div>

          <button
            type="submit"
            disabled={loading}
            className="btn btn-lg"
            style={{
              marginTop: 4,
              background: 'linear-gradient(135deg, #0884F6, #0670d4)',
              color: '#fff',
              boxShadow: loading ? 'none' : '0 4px 20px rgba(8,132,246,0.4)',
              transition: 'all 0.2s',
            }}
          >
            {loading ? <Loader2 size={18} className="loading-spin" /> : null}
            {loading ? 'Masuk...' : 'Masuk'}
          </button>
        </form>
      </div>

      <p
        style={{ textAlign: 'center', color: 'var(--text-3)', fontSize: '0.78rem', marginTop: 20 }}
      >
        Moedah &copy; {new Date().getFullYear()} — Point of Sale System
      </p>
    </div>
  );
}
