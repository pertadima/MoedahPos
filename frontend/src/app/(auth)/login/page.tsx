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
  const [mounted, setMounted] = useState(false);
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

  useEffect(() => {
    setMounted(true);
  }, []);

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
        background: 'var(--bg-base)',
        color: 'var(--text-1)',
        overflow: 'hidden',
        transition: 'background-color 0.4s ease',
      }}
    >
      {/* Left Section - Form Container */}
      <div
        style={{
          flex: '0 0 520px',
          display: 'flex',
          flexDirection: 'column',
          background: 'var(--bg-card)',
          position: 'relative',
          padding: '40px 60px',
          zIndex: 10,
          borderRight: '1px solid var(--border)',
          boxShadow: '4px 0 24px rgba(0,0,0,0.02)',
        }}
      >
        {/* Animated Background Element (Subtle) */}
        <div
          style={{
            position: 'absolute',
            top: '20%',
            left: '-10%',
            width: '400px',
            height: '400px',
            background: 'var(--brand)',
            opacity: isDark ? 0.03 : 0.02,
            filter: 'blur(100px)',
            borderRadius: '50%',
            pointerEvents: 'none',
          }}
        />

        {/* Header - Logo & Theme Toggle */}
        <div
          style={{
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            marginBottom: 'auto',
            animation: 'fadeInDown 0.8s cubic-bezier(0.16, 1, 0.3, 1) both',
          }}
        >
          <div style={{ position: 'relative', width: 130, height: 36 }}>
            <Image
              src={mounted && isDark ? '/logo-icon-dark.svg' : '/logo-icon-light.svg'}
              alt="Moedah Logo"
              fill
              style={{ objectFit: 'contain', objectPosition: 'left' }}
              priority
            />
          </div>
          <button
            onClick={toggleTheme}
            style={{
              width: 40,
              height: 40,
              borderRadius: '12px',
              border: '1px solid var(--border-md)',
              background: 'var(--bg-elevated)',
              color: 'var(--text-2)',
              cursor: 'pointer',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              transition: 'all 0.2s cubic-bezier(0.16, 1, 0.3, 1)',
            }}
            onMouseEnter={e => {
              e.currentTarget.style.borderColor = 'var(--brand)';
              e.currentTarget.style.color = 'var(--brand)';
              e.currentTarget.style.transform = 'scale(1.05)';
            }}
            onMouseLeave={e => {
              e.currentTarget.style.borderColor = 'var(--border-md)';
              e.currentTarget.style.color = 'var(--text-2)';
              e.currentTarget.style.transform = 'scale(1)';
            }}
          >
            {mounted ? isDark ? <Sun size={18} /> : <Moon size={18} /> : <Sun size={18} />}
          </button>
        </div>

        {/* Form Body */}
        <div
          style={{ flex: 2, display: 'flex', flexDirection: 'column', justifyContent: 'center' }}
        >
          <div style={{ animation: 'fadeInUp 0.8s cubic-bezier(0.16, 1, 0.3, 1) 0.1s both' }}>
            <h1
              style={{
                fontSize: '2.25rem',
                fontWeight: 800,
                letterSpacing: '-0.03em',
                marginBottom: 12,
                background: 'linear-gradient(to bottom right, var(--text-1), var(--text-2))',
                WebkitBackgroundClip: 'text',
                WebkitTextFillColor: 'transparent',
              }}
            >
              Selamat datang kembali
            </h1>
            <p
              style={{
                color: 'var(--text-2)',
                fontSize: '1rem',
                lineHeight: 1.6,
                marginBottom: 40,
                maxWidth: '90%',
              }}
            >
              Kelola operasional dan pantau performa bisnis Anda dalam satu dashboard terintegrasi.
            </p>
          </div>

          <form
            onSubmit={handleSubmit}
            style={{
              display: 'flex',
              flexDirection: 'column',
              gap: 24,
            }}
          >
            <div
              style={{ animation: 'fadeInUp 0.8s cubic-bezier(0.16, 1, 0.3, 1) 0.2s both' }}
              className="input-group"
            >
              <label className="input-label" style={{ marginBottom: 8, display: 'block' }}>
                Email Bisnis
              </label>
              <div style={{ position: 'relative' }}>
                <Mail
                  size={19}
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
                  placeholder="admin@bisnisanda.com"
                  style={{
                    width: '100%',
                    padding: '14px 16px 14px 46px',
                    background: 'var(--bg-elevated)',
                    border: '1.5px solid var(--border)',
                    borderRadius: '14px',
                    fontSize: '0.95rem',
                    color: 'var(--text-1)',
                    outline: 'none',
                    transition: 'all 0.3s cubic-bezier(0.16, 1, 0.3, 1)',
                  }}
                  onFocus={e => {
                    e.target.style.borderColor = 'var(--brand)';
                    e.target.style.background = 'var(--bg-card)';
                    e.target.style.boxShadow = '0 0 0 4px rgba(8, 132, 246, 0.08)';
                  }}
                  onBlur={e => {
                    e.target.style.borderColor = 'var(--border)';
                    e.target.style.background = 'var(--bg-elevated)';
                    e.target.style.boxShadow = 'none';
                  }}
                />
              </div>
            </div>

            <div
              style={{ animation: 'fadeInUp 0.8s cubic-bezier(0.16, 1, 0.3, 1) 0.3s both' }}
              className="input-group"
            >
              <label className="input-label" style={{ marginBottom: 8, display: 'block' }}>
                Kata Sandi
              </label>
              <div style={{ position: 'relative' }}>
                <Lock
                  size={19}
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
                  placeholder="••••••••"
                  style={{
                    width: '100%',
                    padding: '14px 48px 14px 46px',
                    background: 'var(--bg-elevated)',
                    border: '1.5px solid var(--border)',
                    borderRadius: '14px',
                    fontSize: '0.95rem',
                    color: 'var(--text-1)',
                    outline: 'none',
                    transition: 'all 0.3s cubic-bezier(0.16, 1, 0.3, 1)',
                  }}
                  onFocus={e => {
                    e.target.style.borderColor = 'var(--brand)';
                    e.target.style.background = 'var(--bg-card)';
                    e.target.style.boxShadow = '0 0 0 4px rgba(8, 132, 246, 0.08)';
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
                    transition: 'all 0.2s',
                    zIndex: 1,
                  }}
                >
                  {showPw ? <EyeOff size={19} /> : <Eye size={19} />}
                </button>
              </div>
            </div>

            <div
              style={{
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'space-between',
                animation: 'fadeInUp 0.8s cubic-bezier(0.16, 1, 0.3, 1) 0.4s both',
              }}
            >
              <label
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: 10,
                  cursor: 'pointer',
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
                padding: '16px',
                borderRadius: '16px',
                fontSize: '1rem',
                fontWeight: 700,
                marginTop: 8,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                gap: 12,
                animation: 'fadeInUp 0.8s cubic-bezier(0.16, 1, 0.3, 1) 0.5s both',
                boxShadow: '0 8px 20px rgba(8, 132, 246, 0.2)',
              }}
            >
              {loading ? (
                <Loader2 size={20} style={{ animation: 'spin 1s linear infinite' }} />
              ) : null}
              {loading ? 'Menghubungkan...' : 'Masuk ke Akun'}
            </button>
          </form>

          {error && (
            <div
              style={{
                marginTop: 24,
                background: 'rgba(239, 68, 68, 0.08)',
                border: '1px solid rgba(239, 68, 68, 0.1)',
                borderRadius: '12px',
                padding: '12px 16px',
                color: '#ef4444',
                fontSize: '0.9rem',
                display: 'flex',
                alignItems: 'center',
                gap: 10,
                animation: 'shake 0.4s cubic-bezier(0.36, 0.07, 0.19, 0.97) both',
              }}
            >
              <div style={{ width: 6, height: 6, background: '#ef4444', borderRadius: '50%' }} />
              {error}
            </div>
          )}
        </div>

        {/* Footer */}
        <div
          style={{
            marginTop: 'auto',
            textAlign: 'center',
            fontSize: '0.85rem',
            color: 'var(--text-3)',
            animation: 'fadeInUp 0.8s cubic-bezier(0.16, 1, 0.3, 1) 0.6s both',
          }}
        >
          <p>© 2025 Moedah Enterprises. Built for scale.</p>
        </div>
      </div>

      {/* Right Section - Immersive Slider */}
      <div
        style={{
          flex: 1,
          background: 'linear-gradient(135deg, var(--brand) 0%, var(--brand-dk) 100%)',
          position: 'relative',
          overflow: 'hidden',
          display: 'flex',
          flexDirection: 'column',
          justifyContent: 'center',
          alignItems: 'center',
          transition: 'background 0.4s ease',
        }}
        onMouseEnter={() => setIsPaused(true)}
        onMouseLeave={() => setIsPaused(false)}
      >
        {/* Animated Background Orbs */}
        <div className="orb-container">
          <div className="orb orb-1" />
          <div className="orb orb-2" />
          <div className="orb orb-3" />
        </div>

        {/* Slider Logic */}
        <div
          style={{
            width: '100%',
            height: '100%',
            position: 'relative',
            zIndex: 2,
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
            justifyContent: 'center',
            padding: '60px',
          }}
        >
          {SLIDES.map((slide, index) => (
            <div
              key={slide.id}
              style={{
                position: 'absolute',
                inset: 0,
                display: 'flex',
                flexDirection: 'column',
                alignItems: 'center',
                justifyContent: 'center',
                opacity: currentSlide === index ? 1 : 0,
                transform: `scale(${currentSlide === index ? 1 : 0.9}) translateY(${currentSlide === index ? 0 : 20}px)`,
                visibility: currentSlide === index ? 'visible' : 'hidden',
                transition: 'all 0.8s cubic-bezier(0.16, 1, 0.3, 1)',
                padding: '60px',
              }}
            >
              {/* Image with Floating Effect */}
              <div
                style={{
                  width: '100%',
                  maxWidth: 800,
                  aspectRatio: '16/10',
                  position: 'relative',
                  marginBottom: 30,
                  borderRadius: 24,
                  overflow: 'hidden',
                  boxShadow: 'none',
                  transform: currentSlide === index ? 'translateY(0)' : 'translateY(40px)',
                  transition: 'all 1s cubic-bezier(0.16, 1, 0.3, 1) 0.1s',
                  animation: currentSlide === index ? 'float 6s ease-in-out infinite' : 'none',
                }}
              >
                <Image
                  src={mounted && isDark ? slide.dark : slide.light}
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
                  maxWidth: 600,
                  color: 'white',
                  transform: currentSlide === index ? 'translateY(0)' : 'translateY(20px)',
                  transition: 'all 0.8s cubic-bezier(0.16, 1, 0.3, 1) 0.2s',
                }}
              >
                <h2
                  style={{
                    fontSize: '1.6rem',
                    fontWeight: 800,
                    marginBottom: 12,
                    letterSpacing: '-0.03em',
                    whiteSpace: 'nowrap',
                    overflow: 'hidden',
                    textOverflow: 'ellipsis',
                  }}
                >
                  {slide.title}
                </h2>
                <p
                  style={{
                    fontSize: '0.95rem',
                    opacity: 0.8,
                    lineHeight: 1.6,
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

          {/* Side Arrows - Aligned to Image Center */}
          <button
            onClick={prevSlide}
            className="nav-btn nav-btn-left"
            style={{ top: '42%', width: 48, height: 48, borderRadius: '15px' }}
            aria-label="Previous Slide"
          >
            <ChevronLeft size={24} strokeWidth={2} />
          </button>

          <button
            onClick={nextSlide}
            className="nav-btn nav-btn-right"
            style={{ top: '42%', width: 48, height: 48, borderRadius: '15px' }}
            aria-label="Next Slide"
          >
            <ChevronRight size={24} strokeWidth={2} />
          </button>

          {/* Progress Indicators */}
          <div className="progress-container">
            {SLIDES.map((_, i) => (
              <div
                key={i}
                onClick={() => setCurrentSlide(i)}
                className={`progress-bar ${currentSlide === i ? 'active' : ''}`}
                style={{
                  width: currentSlide === i ? 10 : 6,
                  height: currentSlide === i ? 10 : 6,
                  borderRadius: '50%',
                  opacity: currentSlide === i ? 1 : 0.3,
                }}
              />
            ))}
          </div>
        </div>
      </div>

      <style>{`
        /* ── Modern Keyframes ── */
        @keyframes spin {
          from { transform: rotate(0deg); }
          to { transform: rotate(360deg); }
        }
        @keyframes fadeInUp {
          from { opacity: 0; transform: translateY(20px); }
          to { opacity: 1; transform: translateY(0); }
        }
        @keyframes fadeInDown {
          from { opacity: 0; transform: translateY(-20px); }
          to { opacity: 1; transform: translateY(0); }
        }
        @keyframes shake {
          10%, 90% { transform: translateX(-1px); }
          20%, 80% { transform: translateX(2px); }
          30%, 50%, 70% { transform: translateX(-4px); }
          40%, 60% { transform: translateX(4px); }
        }
        @keyframes float {
          0% { transform: translateY(0px); }
          50% { transform: translateY(-12px); }
          100% { transform: translateY(0px); }
        }
        @keyframes orbit {
          from { transform: rotate(0deg) translateX(40px) rotate(0deg); }
          to { transform: rotate(360deg) translateX(40px) rotate(-360deg); }
        }

        /* ── Dynamic Layout Styles ── */
        .nav-btn {
          position: absolute;
          top: 50%;
          width: 54px;
          height: 54px;
          border-radius: 18px;
          background: rgba(255, 255, 255, 0.1);
          backdrop-filter: blur(12px);
          border: 1px solid rgba(255, 255, 255, 0.15);
          color: white;
          cursor: pointer;
          display: flex;
          align-items: center;
          justify-content: center;
          transition: all 0.4s cubic-bezier(0.16, 1, 0.3, 1);
          z-index: 20;
          transform: translateY(-50%);
        }
        .nav-btn:hover {
          background: rgba(255, 255, 255, 0.2);
          box-shadow: 0 10px 30px rgba(0, 0, 0, 0.15);
        }
        .nav-btn-left { left: 40px; }
        .nav-btn-left:hover { transform: translateY(-50%) translateX(-8px); }
        .nav-btn-right { right: 40px; }
        .nav-btn-right:hover { transform: translateY(-50%) translateX(8px); }

        .progress-container {
          position: absolute;
          bottom: 50px;
          display: flex;
          gap: 12px;
          align-items: center;
          z-index: 10;
        }
        .progress-bar {
          background: white;
          cursor: pointer;
          transition: all 0.5s cubic-bezier(0.16, 1, 0.3, 1);
        }

        /* ── Animated Background Orbs ── */
        .orb-container {
          position: absolute;
          inset: 0;
          pointer-events: none;
        }
        .orb {
          position: absolute;
          border-radius: 50%;
          filter: blur(140px);
          opacity: 0.3;
          animation: float 10s ease-in-out infinite alternate;
        }
        .orb-1 {
          width: 600px;
          height: 600px;
          background: var(--brand);
          top: -200px;
          right: -200px;
        }
        .orb-2 {
          width: 500px;
          height: 500px;
          background: var(--brand-dk);
          bottom: -150px;
          left: -150px;
          animation-delay: -3s;
        }
        .orb-3 {
          width: 300px;
          height: 300px;
          background: var(--brand-lt);
          top: 40%;
          left: 30%;
          opacity: 0.1;
          animation-delay: -7s;
        }

        /* ── Form Fine-tuning ── */
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
