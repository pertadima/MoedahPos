'use client';

import { useState, FormEvent } from 'react';
import { useRouter } from 'next/navigation';
import { CreditCard, Mail, Lock, Eye, EyeOff, Loader2 } from 'lucide-react';
import { useAuth } from '@/lib/auth/AuthContext';
import { ApiError } from '@/lib/api/client';

export default function LoginPage() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [showPw, setShowPw] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const { login } = useAuth();
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
      {/* Logo */}
      <div style={{ textAlign: 'center', marginBottom: 32 }}>
        <div style={{
          width: 56, height: 56, borderRadius: 16, margin: '0 auto 12px',
          background: 'linear-gradient(135deg, var(--accent-em), #059669)',
          display: 'flex', alignItems: 'center', justifyContent: 'center',
          boxShadow: '0 0 32px rgba(16,185,129,0.25)',
        }}>
          <CreditCard size={26} color="#fff" />
        </div>
        <h1 style={{ fontSize: '1.6rem', fontWeight: 800, color: 'var(--text-1)', letterSpacing: '-0.5px' }}>
          MoedahPOS
        </h1>
        <p style={{ color: 'var(--text-2)', marginTop: 4, fontSize: '0.9rem' }}>
          Masuk ke akun Anda
        </p>
      </div>

      {/* Card */}
      <div className="card" style={{ padding: '28px 28px 24px' }}>
        {error && (
          <div style={{
            background: 'rgba(239,68,68,0.12)', border: '1px solid rgba(239,68,68,0.3)',
            borderRadius: 8, padding: '10px 14px', marginBottom: 18,
            color: '#f87171', fontSize: '0.85rem',
          }}>
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit} style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
          {/* Email */}
          <div className="input-group">
            <label className="input-label">Email</label>
            <div style={{ position: 'relative' }}>
              <Mail size={15} style={{ position: 'absolute', left: 12, top: '50%', transform: 'translateY(-50%)', color: 'var(--text-3)' }} />
              <input
                type="email" required value={email} onChange={e => setEmail(e.target.value)}
                placeholder="nama@email.com"
                className="input" style={{ paddingLeft: 36 }}
              />
            </div>
          </div>

          {/* Password */}
          <div className="input-group">
            <label className="input-label">Password</label>
            <div style={{ position: 'relative' }}>
              <Lock size={15} style={{ position: 'absolute', left: 12, top: '50%', transform: 'translateY(-50%)', color: 'var(--text-3)' }} />
              <input
                type={showPw ? 'text' : 'password'} required value={password}
                onChange={e => setPassword(e.target.value)}
                placeholder="Password Anda"
                className="input" style={{ paddingLeft: 36, paddingRight: 40 }}
              />
              <button type="button" onClick={() => setShowPw(!showPw)}
                style={{ position: 'absolute', right: 10, top: '50%', transform: 'translateY(-50%)', background: 'none', border: 'none', cursor: 'pointer', color: 'var(--text-3)', padding: 2 }}>
                {showPw ? <EyeOff size={15} /> : <Eye size={15} />}
              </button>
            </div>
          </div>

          <button type="submit" disabled={loading} className="btn btn-primary btn-lg" style={{ marginTop: 4 }}>
            {loading ? <Loader2 size={18} className="loading-spin" /> : null}
            {loading ? 'Masuk...' : 'Masuk'}
          </button>
        </form>
      </div>

      <p style={{ textAlign: 'center', color: 'var(--text-3)', fontSize: '0.78rem', marginTop: 20 }}>
        MoedahPOS &copy; {new Date().getFullYear()} — Multi-store POS System
      </p>
    </div>
  );
}
