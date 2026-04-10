'use client';

import React, { useState, type FormEvent } from 'react';
import { useRouter } from 'next/navigation';
import Image from 'next/image';
import { Mail, Lock, Eye, EyeOff, Loader2 } from 'lucide-react';
import { useAuth } from '@/lib/auth/AuthContext';
import { ApiError } from '@/lib/api/client';

export default function LoginPage() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [showPw, setShowPw] = useState(false);
  const [rememberMe, setRememberMe] = useState(false);
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
    <div
      style={{
        display: 'flex',
        height: '100vh',
        width: '100vw',
        margin: 0,
        padding: 0,
        background: '#ffffff',
        overflow: 'hidden',
      }}
    >
      {/* Left Section - Login Form */}
      <div
        style={{
          flex: 1,
          display: 'flex',
          flexDirection: 'column',
          padding: '60px 40px',
          background: '#ffffff',
          animation: 'slideInLeft 0.6s ease-out',
          position: 'relative',
        }}
      >
        {/* Logo - Top Left */}
        <div
          style={{
            position: 'absolute',
            top: 40,
            left: 40,
            animation: 'fadeIn 0.8s ease-out',
            width: 180,
            height: 60,
          }}
        >
          <Image
            src="/logo-icon-light.svg"
            alt="Moedah"
            width={180}
            height={60}
            style={{
              display: 'block',
              width: '180px',
              height: '60px',
            }}
            priority
          />
        </div>

        {/* Centered Content */}
        <div
          style={{
            flex: 1,
            display: 'flex',
            flexDirection: 'column',
            justifyContent: 'center',
            alignItems: 'center',
          }}
        >
          <div style={{ width: '100%', maxWidth: 420 }}>
            {/* Header */}
            <div style={{ marginBottom: 40, animation: 'fadeIn 0.8s ease-out' }}>
              <h1
                style={{ fontSize: '1.8rem', fontWeight: 800, color: '#1f2937', marginBottom: 8 }}
              >
                Selamat Datang Kembali
              </h1>
              <p style={{ color: '#6b7280', fontSize: '0.95rem' }}>
                Masukkan email dan kata sandi Anda untuk mengakses akun Anda.
              </p>
            </div>

            {/* Error Message */}
            {error && (
              <div
                style={{
                  background: '#fee2e2',
                  border: '1px solid #fecaca',
                  borderRadius: 8,
                  padding: '12px 14px',
                  marginBottom: 20,
                  color: '#991b1b',
                  fontSize: '0.85rem',
                  animation: 'slideDown 0.3s ease-out',
                }}
              >
                {error}
              </div>
            )}

            {/* Form */}
            <form
              onSubmit={handleSubmit}
              style={{
                display: 'flex',
                flexDirection: 'column',
                gap: 16,
                animation: 'fadeIn 1s ease-out',
              }}
            >
              {/* Email Field */}
              <div>
                <label
                  style={{
                    display: 'block',
                    color: '#374151',
                    fontSize: '0.9rem',
                    fontWeight: 500,
                    marginBottom: 6,
                  }}
                >
                  Email
                </label>
                <div style={{ position: 'relative' }}>
                  <Mail
                    size={18}
                    style={{
                      position: 'absolute',
                      left: 12,
                      top: '50%',
                      transform: 'translateY(-50%)',
                      color: '#9ca3af',
                    }}
                  />
                  <input
                    type="email"
                    required
                    value={email}
                    onChange={e => setEmail(e.target.value)}
                    placeholder="toko@perusahaan.com"
                    style={{
                      width: '100%',
                      padding: '11px 12px 11px 40px',
                      border: '1px solid #e5e7eb',
                      borderRadius: 6,
                      fontSize: '0.95rem',
                      outline: 'none',
                      transition: 'all 0.2s',
                      boxSizing: 'border-box',
                    }}
                    onFocus={e => {
                      e.target.style.borderColor = '#0884F6';
                      e.target.style.boxShadow = '0 0 0 3px rgba(8, 132, 246, 0.1)';
                      e.target.style.background = '#f8fbff';
                    }}
                    onBlur={e => {
                      e.target.style.borderColor = '#e5e7eb';
                      e.target.style.boxShadow = 'none';
                      e.target.style.background = '#ffffff';
                    }}
                  />
                </div>
              </div>

              {/* Password Field */}
              <div>
                <label
                  style={{
                    display: 'block',
                    color: '#374151',
                    fontSize: '0.9rem',
                    fontWeight: 500,
                    marginBottom: 6,
                  }}
                >
                  Kata Sandi
                </label>
                <div style={{ position: 'relative' }}>
                  <Lock
                    size={18}
                    style={{
                      position: 'absolute',
                      left: 12,
                      top: '50%',
                      transform: 'translateY(-50%)',
                      color: '#9ca3af',
                    }}
                  />
                  <input
                    type={showPw ? 'text' : 'password'}
                    required
                    value={password}
                    onChange={e => setPassword(e.target.value)}
                    placeholder="Masukkan kata sandi Anda"
                    style={{
                      width: '100%',
                      padding: '11px 40px 11px 40px',
                      border: '1px solid #e5e7eb',
                      borderRadius: 6,
                      fontSize: '0.95rem',
                      outline: 'none',
                      transition: 'all 0.2s',
                      boxSizing: 'border-box',
                    }}
                    onFocus={e => {
                      e.target.style.borderColor = '#0884F6';
                      e.target.style.boxShadow = '0 0 0 3px rgba(8, 132, 246, 0.1)';
                      e.target.style.background = '#f8fbff';
                    }}
                    onBlur={e => {
                      e.target.style.borderColor = '#e5e7eb';
                      e.target.style.boxShadow = 'none';
                      e.target.style.background = '#ffffff';
                    }}
                  />
                  <button
                    type="button"
                    onClick={() => setShowPw(!showPw)}
                    style={{
                      position: 'absolute',
                      right: 12,
                      top: '50%',
                      transform: 'translateY(-50%)',
                      background: 'none',
                      border: 'none',
                      cursor: 'pointer',
                      color: '#9ca3af',
                      padding: '4px 6px',
                      display: 'flex',
                      alignItems: 'center',
                      transition: 'color 0.2s',
                    }}
                    onMouseEnter={e => (e.currentTarget.style.color = '#6b7280')}
                    onMouseLeave={e => (e.currentTarget.style.color = '#9ca3af')}
                  >
                    {showPw ? <EyeOff size={18} /> : <Eye size={18} />}
                  </button>
                </div>
              </div>

              {/* Remember Me */}
              <div style={{ display: 'flex', alignItems: 'center' }}>
                <input
                  type="checkbox"
                  id="remember"
                  checked={rememberMe}
                  onChange={e => setRememberMe(e.target.checked)}
                  style={{
                    width: 16,
                    height: 16,
                    cursor: 'pointer',
                    accentColor: '#0884F6',
                  }}
                />
                <label
                  htmlFor="remember"
                  style={{
                    marginLeft: 8,
                    color: '#6b7280',
                    fontSize: '0.9rem',
                    cursor: 'pointer',
                    userSelect: 'none',
                  }}
                >
                  Ingat Saya
                </label>
              </div>

              {/* Login Button */}
              <button
                type="submit"
                disabled={loading}
                style={{
                  width: '100%',
                  padding: '11px',
                  marginTop: 8,
                  background: loading ? '#B3D9F2' : '#0884F6',
                  color: '#fff',
                  border: 'none',
                  borderRadius: 6,
                  fontSize: '0.95rem',
                  fontWeight: 600,
                  cursor: loading ? 'not-allowed' : 'pointer',
                  transition: 'all 0.2s',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  gap: 8,
                  boxShadow: loading ? 'none' : '0 2px 8px rgba(8, 132, 246, 0.3)',
                }}
                onMouseEnter={e => {
                  if (!loading) {
                    e.currentTarget.style.background = '#0670D4';
                    e.currentTarget.style.boxShadow = '0 4px 12px rgba(8, 132, 246, 0.4)';
                  }
                }}
                onMouseLeave={e => {
                  if (!loading) {
                    e.currentTarget.style.background = '#0884F6';
                    e.currentTarget.style.boxShadow = '0 2px 8px rgba(8, 132, 246, 0.3)';
                  }
                }}
              >
                {loading ? (
                  <Loader2 size={18} style={{ animation: 'spin 1s linear infinite' }} />
                ) : null}
                {loading ? 'Masuk...' : 'Masuk'}
              </button>
            </form>
          </div>
        </div>

        {/* Footer */}
        <div
          style={{
            position: 'absolute',
            bottom: 20,
            left: 0,
            right: 0,
            textAlign: 'center',
            fontSize: '0.8rem',
            color: '#9ca3af',
          }}
        >
          <p>Copyright © 2025 Moedah Enterprises LTD.</p>
        </div>
      </div>

      {/* Right Section - Blue Gradient with Content */}
      <div
        style={{
          flex: 1,
          background: 'linear-gradient(135deg, #0884F6 0%, #0670D4 100%)',
          display: 'flex',
          flexDirection: 'column',
          justifyContent: 'center',
          alignItems: 'center',
          padding: '60px 40px',
          position: 'relative',
          overflow: 'hidden',
          animation: 'slideInRight 0.6s ease-out',
        }}
      >
        {/* Content */}
        <div
          style={{
            position: 'relative',
            zIndex: 1,
            textAlign: 'center',
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
            gap: 30,
            animation: 'fadeIn 1s ease-out 0.3s backwards',
          }}
        >
          {/* Dashboard Image */}
          <div
            style={{
              animation: 'slideUp 0.8s ease-out 0.5s backwards',
              borderRadius: 12,
              overflow: 'hidden',
              width: '100%',
              maxWidth: 600,
            }}
          >
            <Image
              src="/dashboard.webp"
              alt="Pratinjau Dashboard"
              width={600}
              height={375}
              style={{
                display: 'block',
                width: '100%',
                height: 'auto',
              }}
              priority
            />
          </div>

          {/* Text Below Image */}
          <div style={{ maxWidth: 350 }}>
            <h2
              style={{
                color: '#fff',
                fontSize: '1.5rem',
                fontWeight: 700,
                marginBottom: 12,
                lineHeight: 1.3,
              }}
            >
              Kelola restoran dan penjualan Anda dengan efisien.
            </h2>
            <p style={{ color: 'rgba(255,255,255,0.85)', fontSize: '0.85rem', marginBottom: 0 }}>
              Akses dashboard POS Anda untuk mengelola pesanan, inventori, dan penjualan.
            </p>
          </div>
        </div>
      </div>

      <style>{`
        @keyframes spin {
          from { transform: rotate(0deg); }
          to { transform: rotate(360deg); }
        }
        @keyframes slideInLeft {
          from {
            opacity: 0;
            transform: translateX(-30px);
          }
          to {
            opacity: 1;
            transform: translateX(0);
          }
        }
        @keyframes slideInRight {
          from {
            opacity: 0;
            transform: translateX(30px);
          }
          to {
            opacity: 1;
            transform: translateX(0);
          }
        }
        @keyframes fadeIn {
          from {
            opacity: 0;
          }
          to {
            opacity: 1;
          }
        }
        @keyframes slideDown {
          from {
            opacity: 0;
            transform: translateY(-10px);
          }
          to {
            opacity: 1;
            transform: translateY(0);
          }
        }
        @keyframes slideUp {
          from {
            opacity: 0;
            transform: translateY(20px);
          }
          to {
            opacity: 1;
            transform: translateY(0);
          }
        }
      `}</style>
    </div>
  );
}
