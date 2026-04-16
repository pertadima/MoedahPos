'use client';

import React, { useState, useEffect, useCallback, type FormEvent } from 'react';
import { useRouter } from 'next/navigation';
import Image from 'next/image';
import {
  Mail,
  Lock,
  Eye,
  EyeOff,
  Loader2,
  Sun,
  Moon,
  ChevronRight,
  ChevronLeft,
} from 'lucide-react';
import { useAuth } from '@/lib/auth/AuthContext';
import { ApiError } from '@/lib/api/client';
import { useTheme } from '@/lib/theme/ThemeContext';

const SLIDES = [
  {
    id: 1,
    light: '/order-pos-light.png',
    dark: '/order-pos-dark.png',
    title: 'Point of Sale yang Mudah',
    subtitle: 'Kelola pesanan dan transaksi toko Anda dengan cepat dan akurat.',
  },
  {
    id: 2,
    light: '/order-hpp-light.png',
    dark: '/order-hpp-dark.png',
    title: 'Kalkulator HPP Otomatis',
    subtitle: 'Hitung Harga Pokok Penjualan secara instan untuk margin keuntungan yang lebih baik.',
  },
  {
    id: 3,
    light: '/order-purchase-light.png',
    dark: '/order-purchase-dark.png',
    title: 'Manajemen Purchase Order',
    subtitle: 'Pantau stok masuk dan kelola pesanan ke supplier dalam satu dashboard.',
  },
  {
    id: 4,
    light: '/order-report-light.png',
    dark: '/order-report-dark.png',
    title: 'Laporan Bisnis Real-time',
    subtitle: 'Dapatkan wawasan mendalam tentang performa bisnis Anda kapan saja, di mana saja.',
  },
  {
    id: 5,
    light: '/order-cashflow-light.png',
    dark: '/order-cashflow-dark.png',
    title: 'Catatan Arus Kas',
    subtitle: 'Lacak setiap uang masuk dan keluar dengan detail untuk transparansi finansial.',
  },
];

export default function LoginPage() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [showPw, setShowPw] = useState(false);
  const [rememberMe, setRememberMe] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const { login } = useAuth();
  const router = useRouter();
  const { toggleTheme, isDark } = useTheme();

  // Slider State
  const [currentSlide, setCurrentSlide] = useState(0);
  const [isPaused, setIsPaused] = useState(false);

  const nextSlide = useCallback(() => {
    setCurrentSlide(prev => (prev + 1) % SLIDES.length);
  }, []);

  const prevSlide = useCallback(() => {
    setCurrentSlide(prev => (prev - 1 + SLIDES.length) % SLIDES.length);
  }, []);

  useEffect(() => {
    if (isPaused) return;
    const timer = setInterval(nextSlide, 6000);
    return () => clearInterval(timer);
  }, [isPaused, nextSlide]);

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
        background: 'var(--bg-base)',
        color: 'var(--text-1)',
        overflow: 'hidden',
        transition: 'background-color 0.3s ease',
      }}
    >
      {/* Left Section - Login Form */}
      <div
        style={{
          flex: 1,
          display: 'flex',
          flexDirection: 'column',
          padding: '60px 40px',
          background: 'var(--bg-card)',
          position: 'relative',
          zIndex: 10,
          boxShadow: '20px 0 50px rgba(0,0,0,0.05)',
        }}
        className="reveal-animate"
      >
        {/* Top Header with Logo & Theme Toggle */}
        <div
          style={{
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            marginBottom: 60,
          }}
        >
          <div style={{ width: 140, height: 40, position: 'relative' }}>
            <Image
              src={isDark ? '/logo-icon-dark.svg' : '/logo-icon-light.svg'}
              alt="Moedah"
              fill
              style={{ objectFit: 'contain', objectPosition: 'left' }}
              priority
            />
          </div>
          <button
            onClick={toggleTheme}
            style={{
              padding: '10px',
              borderRadius: '12px',
              border: '1px solid var(--border-md)',
              background: 'var(--bg-elevated)',
              color: 'var(--text-2)',
              cursor: 'pointer',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              transition: 'all 0.2s ease',
            }}
            onMouseEnter={e => {
              e.currentTarget.style.borderColor = 'var(--brand)';
              e.currentTarget.style.color = 'var(--brand)';
            }}
            onMouseLeave={e => {
              e.currentTarget.style.borderColor = 'var(--border-md)';
              e.currentTarget.style.color = 'var(--text-2)';
            }}
            aria-label="Toggle Theme"
          >
            {isDark ? <Sun size={20} /> : <Moon size={20} />}
          </button>
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
          <div style={{ width: '100%', maxWidth: 400 }}>
            {/* Header */}
            <div style={{ marginBottom: 40 }}>
              <h1
                style={{
                  fontSize: '2rem',
                  fontWeight: 800,
                  color: 'var(--text-1)',
                  marginBottom: 12,
                  letterSpacing: '-0.02em',
                }}
              >
                Selamat Datang
              </h1>
              <p style={{ color: 'var(--text-2)', fontSize: '1rem', lineHeight: 1.6 }}>
                Masuk untuk mengelola operasional bisnis Anda dengan Moedah POS.
              </p>
            </div>

            {/* Error Message */}
            {error && (
              <div
                style={{
                  background: 'rgba(239, 68, 68, 0.1)',
                  border: '1px solid rgba(239, 68, 68, 0.2)',
                  borderRadius: 12,
                  padding: '12px 16px',
                  marginBottom: 24,
                  color: '#ef4444',
                  fontSize: '0.9rem',
                  display: 'flex',
                  alignItems: 'center',
                  gap: 10,
                }}
              >
                <div style={{ width: 6, height: 6, background: '#ef4444', borderRadius: '50%' }} />
                {error}
              </div>
            )}

            {/* Form */}
            <form
              onSubmit={handleSubmit}
              style={{
                display: 'flex',
                flexDirection: 'column',
                gap: 20,
              }}
            >
              <div className="input-group">
                <label className="input-label">Email Kantor / Toko</label>
                <div style={{ position: 'relative' }}>
                  <Mail
                    size={20}
                    style={{
                      position: 'absolute',
                      left: 14,
                      top: '50%',
                      transform: 'translateY(-50%)',
                      color: 'var(--text-3)',
                      zIndex: 1,
                    }}
                  />
                  <input
                    type="email"
                    required
                    value={email}
                    onChange={e => setEmail(e.target.value)}
                    placeholder="nama@toko.com"
                    style={{
                      width: '100%',
                      padding: '14px 16px 14px 46px',
                      background: 'var(--bg-elevated)',
                      border: '1px solid var(--border)',
                      borderRadius: 12,
                      fontSize: '0.95rem',
                      color: 'var(--text-1)',
                      outline: 'none',
                      transition: 'all 0.2s',
                    }}
                    onFocus={e => {
                      e.target.style.borderColor = 'var(--brand)';
                      e.target.style.background = 'var(--bg-card)';
                      e.target.style.boxShadow = '0 0 0 4px rgba(8, 132, 246, 0.1)';
                    }}
                    onBlur={e => {
                      e.target.style.borderColor = 'var(--border)';
                      e.target.style.background = 'var(--bg-elevated)';
                      e.target.style.boxShadow = 'none';
                    }}
                  />
                </div>
              </div>

              <div className="input-group">
                <label className="input-label">Kata Sandi</label>
                <div style={{ position: 'relative' }}>
                  <Lock
                    size={20}
                    style={{
                      position: 'absolute',
                      left: 14,
                      top: '50%',
                      transform: 'translateY(-50%)',
                      color: 'var(--text-3)',
                      zIndex: 1,
                    }}
                  />
                  <input
                    type={showPw ? 'text' : 'password'}
                    required
                    value={password}
                    onChange={e => setPassword(e.target.value)}
                    placeholder="Minimal 8 karakter"
                    style={{
                      width: '100%',
                      padding: '14px 48px 14px 46px',
                      background: 'var(--bg-elevated)',
                      border: '1px solid var(--border)',
                      borderRadius: 12,
                      fontSize: '0.95rem',
                      color: 'var(--text-1)',
                      outline: 'none',
                      transition: 'all 0.2s',
                    }}
                    onFocus={e => {
                      e.target.style.borderColor = 'var(--brand)';
                      e.target.style.background = 'var(--bg-card)';
                      e.target.style.boxShadow = '0 0 0 4px rgba(8, 132, 246, 0.1)';
                    }}
                    onBlur={e => {
                      e.target.style.borderColor = 'var(--border)';
                      e.target.style.background = 'var(--bg-elevated)';
                      e.target.style.boxShadow = 'none';
                    }}
                  />
                  <button
                    type="button"
                    onClick={() => setShowPw(!showPw)}
                    style={{
                      position: 'absolute',
                      right: 14,
                      top: '50%',
                      transform: 'translateY(-50%)',
                      background: 'none',
                      border: 'none',
                      cursor: 'pointer',
                      color: 'var(--text-3)',
                      padding: 4,
                      display: 'flex',
                      alignItems: 'center',
                      transition: 'color 0.2s',
                      zIndex: 1,
                    }}
                    onMouseEnter={e => (e.currentTarget.style.color = 'var(--text-2)')}
                    onMouseLeave={e => (e.currentTarget.style.color = 'var(--text-3)')}
                  >
                    {showPw ? <EyeOff size={20} /> : <Eye size={20} />}
                  </button>
                </div>
              </div>

              <div
                style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}
              >
                <label
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: 10,
                    cursor: 'pointer',
                    userSelect: 'none',
                    fontSize: '0.9rem',
                    color: 'var(--text-2)',
                  }}
                >
                  <input
                    type="checkbox"
                    checked={rememberMe}
                    onChange={e => setRememberMe(e.target.checked)}
                    style={{
                      width: 18,
                      height: 18,
                      accentColor: 'var(--brand)',
                      cursor: 'pointer',
                    }}
                  />
                  Tetap masuk
                </label>
                <button
                  type="button"
                  style={{
                    background: 'none',
                    border: 'none',
                    color: 'var(--brand)',
                    fontSize: '0.9rem',
                    fontWeight: 600,
                    cursor: 'pointer',
                  }}
                >
                  Lupa Sandi?
                </button>
              </div>

              <button
                type="submit"
                disabled={loading}
                className="btn-primary"
                style={{
                  width: '100%',
                  padding: '14px',
                  borderRadius: 12,
                  fontSize: '1rem',
                  fontWeight: 700,
                  marginTop: 10,
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  gap: 10,
                }}
              >
                {loading ? (
                  <Loader2 size={20} style={{ animation: 'spin 1s linear infinite' }} />
                ) : null}
                {loading ? 'Menghubungkan...' : 'Masuk ke Dashboard'}
              </button>
            </form>
          </div>
        </div>

        {/* Footer */}
        <div
          style={{
            marginTop: 40,
            textAlign: 'center',
            fontSize: '0.85rem',
            color: 'var(--text-3)',
          }}
        >
          <p>© 2025 Moedah Enterprises. Bangga melayani UKM Indonesia.</p>
        </div>
      </div>

      {/* Right Section - Slider Content */}
      <div
        style={{
          flex: 1.2,
          background: 'var(--brand)',
          position: 'relative',
          overflow: 'hidden',
          display: 'flex',
          flexDirection: 'column',
          justifyContent: 'center',
          alignItems: 'center',
        }}
        onMouseEnter={() => setIsPaused(true)}
        onMouseLeave={() => setIsPaused(false)}
      >
        {/* Animated Background Patterns */}
        <div
          style={{
            position: 'absolute',
            top: 0,
            left: 0,
            right: 0,
            bottom: 0,
            opacity: 0.1,
            zIndex: 0,
            pointerEvents: 'none',
          }}
        >
          <div
            style={{
              position: 'absolute',
              top: '-10%',
              right: '-5%',
              width: '60%',
              height: '60%',
              borderRadius: '50%',
              background: 'white',
              filter: 'blur(80px)',
            }}
          />
          <div
            style={{
              position: 'absolute',
              bottom: '-5%',
              left: '-10%',
              width: '50%',
              height: '50%',
              borderRadius: '50%',
              background: 'white',
              filter: 'blur(100px)',
            }}
          />
        </div>

        {/* Slider Container */}
        <div
          style={{
            width: '100%',
            height: '100%',
            position: 'relative',
            zIndex: 1,
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
            justifyContent: 'center',
            padding: '40px',
          }}
        >
          {SLIDES.map((slide, index) => (
            <div
              key={slide.id}
              style={{
                position: 'absolute',
                top: 0,
                left: 0,
                width: '100%',
                height: '100%',
                display: 'flex',
                flexDirection: 'column',
                alignItems: 'center',
                justifyContent: 'center',
                opacity: currentSlide === index ? 1 : 0,
                transform: `translateX(${(index - currentSlide) * 20}%)`,
                transition: 'all 0.8s cubic-bezier(0.165, 0.84, 0.44, 1)',
                padding: '40px',
                pointerEvents: currentSlide === index ? 'auto' : 'none',
              }}
            >
              {/* Image Display */}
              <div
                style={{
                  width: '100%',
                  maxWidth: 700,
                  aspectRatio: '16/10',
                  position: 'relative',
                  marginBottom: 40,
                  borderRadius: 16,
                  overflow: 'hidden',
                  transform: currentSlide === index ? 'translateY(0)' : 'translateY(40px)',
                  transition: 'all 0.8s cubic-bezier(0.165, 0.84, 0.44, 1) 0.1s',
                }}
              >
                <Image
                  src={isDark ? slide.dark : slide.light}
                  alt={slide.title}
                  fill
                  style={{ objectFit: 'contain' }}
                  priority={index === 0}
                />
              </div>

              {/* Text Content */}
              <div
                style={{
                  textAlign: 'center',
                  maxWidth: 550,
                  color: 'white',
                  transform: currentSlide === index ? 'translateY(0)' : 'translateY(20px)',
                  transition: 'all 0.8s cubic-bezier(0.165, 0.84, 0.44, 1) 0.2s',
                  opacity: currentSlide === index ? 1 : 0,
                }}
              >
                <h2
                  style={{
                    fontSize: '1.85rem',
                    fontWeight: 800,
                    marginBottom: 12,
                    letterSpacing: '-0.02em',
                    whiteSpace: 'nowrap',
                    overflow: 'hidden',
                    textOverflow: 'ellipsis',
                  }}
                >
                  {slide.title}
                </h2>
                <p
                  style={{
                    fontSize: '1rem',
                    opacity: 0.85,
                    lineHeight: 1.5,
                    margin: 0,
                    display: '-webkit-box',
                    WebkitLineClamp: 2,
                    WebkitBoxOrient: 'vertical',
                    overflow: 'hidden',
                    height: '3rem',
                  }}
                >
                  {slide.subtitle}
                </p>
              </div>
            </div>
          ))}

          {/* Navigation Arrows */}
          <button
            onClick={prevSlide}
            style={{
              position: 'absolute',
              left: 20,
              top: '50%',
              width: 44,
              height: 44,
              borderRadius: '14px',
              background: 'rgba(255,255,255,0.08)',
              backdropFilter: 'blur(8px)',
              border: '1px solid rgba(255,255,255,0.1)',
              color: 'white',
              cursor: 'pointer',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
              zIndex: 20,
              transform: 'translateY(-50%)',
            }}
            onMouseEnter={e => {
              e.currentTarget.style.background = 'rgba(255,255,255,0.15)';
              e.currentTarget.style.transform = 'translateY(-50%) translateX(-6px)';
            }}
            onMouseLeave={e => {
              e.currentTarget.style.background = 'rgba(255,255,255,0.08)';
              e.currentTarget.style.transform = 'translateY(-50%) translateX(0)';
            }}
          >
            <ChevronLeft size={24} strokeWidth={2.5} />
          </button>

          <button
            onClick={nextSlide}
            style={{
              position: 'absolute',
              right: 20,
              top: '50%',
              width: 44,
              height: 44,
              borderRadius: '14px',
              background: 'rgba(255,255,255,0.08)',
              backdropFilter: 'blur(8px)',
              border: '1px solid rgba(255,255,255,0.1)',
              color: 'white',
              cursor: 'pointer',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
              zIndex: 20,
              transform: 'translateY(-50%)',
            }}
            onMouseEnter={e => {
              e.currentTarget.style.background = 'rgba(255,255,255,0.15)';
              e.currentTarget.style.transform = 'translateY(-50%) translateX(6px)';
            }}
            onMouseLeave={e => {
              e.currentTarget.style.background = 'rgba(255,255,255,0.08)';
              e.currentTarget.style.transform = 'translateY(-50%) translateX(0)';
            }}
          >
            <ChevronRight size={24} strokeWidth={2.5} />
          </button>

          {/* Bottom Indicators */}
          <div
            style={{
              position: 'absolute',
              bottom: 40,
              display: 'flex',
              gap: 8,
              alignItems: 'center',
              zIndex: 10,
            }}
          >
            {SLIDES.map((_, i) => (
              <div
                key={i}
                onClick={() => setCurrentSlide(i)}
                style={{
                  width: currentSlide === i ? 36 : 10,
                  height: 5,
                  borderRadius: 10,
                  background: 'white',
                  opacity: currentSlide === i ? 1 : 0.25,
                  cursor: 'pointer',
                  transition: 'all 0.4s cubic-bezier(0.4, 0, 0.2, 1)',
                }}
              />
            ))}
          </div>
        </div>
      </div>

      <style>{`
        @keyframes spin {
          from { transform: rotate(0deg); }
          to { transform: rotate(360deg); }
        }
        @keyframes fadeIn {
          from { opacity: 0; }
          to { opacity: 1; }
        }
        .reveal-animate {
          animation: fadeIn 0.8s ease-out;
        }
        input:-webkit-autofill,
        input:-webkit-autofill:hover, 
        input:-webkit-autofill:focus {
          -webkit-text-fill-color: var(--text-1);
          -webkit-box-shadow: 0 0 0px 1000px var(--bg-elevated) inset;
          transition: background-color 5000s ease-in-out 0s;
        }
      `}</style>
    </div>
  );
}
