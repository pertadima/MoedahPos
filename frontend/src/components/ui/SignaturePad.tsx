'use client';

import { useEffect, useRef, useState } from 'react';
import { RotateCcw, Check } from 'lucide-react';

interface SignaturePadProps {
  value?: string;
  onSave: (dataUrl: string) => void;
  onClear?: () => void;
  label?: string;
  readOnly?: boolean;
}

const DPR = typeof window !== 'undefined' ? window.devicePixelRatio || 1 : 1;
const W = 300;
const H = 100;

export default function SignaturePad({
  value,
  onSave,
  onClear,
  label = 'Tanda Tangan',
  readOnly = false,
}: SignaturePadProps) {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const ctxRef = useRef<CanvasRenderingContext2D | null>(null);
  const isDrawingRef = useRef(false);
  const lastRef = useRef<{ x: number; y: number } | null>(null);
  const [hasSignature, setHasSignature] = useState(!!value);

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    canvas.width = W * DPR;
    canvas.height = H * DPR;
    const ctx = canvas.getContext('2d', { willReadFrequently: true });
    if (!ctx) return;
    ctx.scale(DPR, DPR);
    ctxRef.current = ctx;

    if (value) {
      const img = new Image();
      img.onload = () => {
        ctx.clearRect(0, 0, W, H);
        ctx.drawImage(img, 0, 0, W, H);
        setHasSignature(true);
      };
      img.src = value;
    }
  }, []);

  const getPos = (e: MouseEvent | TouchEvent) => {
    const canvas = canvasRef.current;
    if (!canvas) return { x: 0, y: 0 };
    const rect = canvas.getBoundingClientRect();
    if ('touches' in e) {
      const touch = e.touches[0];
      return { x: touch.clientX - rect.left, y: touch.clientY - rect.top };
    }
    return { x: e.clientX - rect.left, y: e.clientY - rect.top };
  };

  const startDraw = (e: React.MouseEvent | React.TouchEvent) => {
    if (readOnly) return;
    e.preventDefault();
    const ctx = ctxRef.current;
    if (!ctx) return;
    isDrawingRef.current = true;
    lastRef.current = null;
    const pos = getPos(e.nativeEvent as MouseEvent | TouchEvent);
    ctx.beginPath();
    ctx.moveTo(pos.x, pos.y);
    setHasSignature(true);
  };

  const draw = (e: React.MouseEvent | React.TouchEvent) => {
    if (!isDrawingRef.current || readOnly) return;
    e.preventDefault();
    const ctx = ctxRef.current;
    if (!ctx) return;
    const pos = getPos(e.nativeEvent as MouseEvent | TouchEvent);
    const last = lastRef.current;
    ctx.lineWidth = 1.5;
    ctx.lineCap = 'round';
    ctx.lineJoin = 'round';
    ctx.strokeStyle = '#1a1a1a';
    if (last) {
      ctx.beginPath();
      ctx.moveTo(last.x, last.y);
      ctx.lineTo(pos.x, pos.y);
      ctx.stroke();
    } else {
      ctx.beginPath();
      ctx.moveTo(pos.x, pos.y);
      ctx.lineTo(pos.x + 0.1, pos.y + 0.1);
      ctx.stroke();
    }
    lastRef.current = pos;
  };

  const stopDraw = () => {
    isDrawingRef.current = false;
    lastRef.current = null;
  };

  const clear = () => {
    const ctx = ctxRef.current;
    if (!ctx) return;
    ctx.clearRect(0, 0, W, H);
    setHasSignature(false);
    onClear?.();
  };

  const save = () => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    onSave(canvas.toDataURL('image/png'));
  };

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
      {label && (
        <span style={{ fontSize: '0.72rem', fontWeight: 600, color: 'var(--text-2)' }}>
          {label}
        </span>
      )}
      <div
        style={{
          border: '1px solid var(--border)',
          borderRadius: 8,
          overflow: 'hidden',
          background: '#fff',
        }}
      >
        <canvas
          ref={canvasRef}
          style={{
            display: 'block',
            width: W,
            height: H,
            cursor: readOnly ? 'default' : 'crosshair',
            touchAction: 'none',
          }}
          onMouseDown={startDraw}
          onMouseMove={draw}
          onMouseUp={stopDraw}
          onMouseLeave={stopDraw}
          onTouchStart={startDraw}
          onTouchMove={draw}
          onTouchEnd={stopDraw}
        />
        {!readOnly && (
          <div
            style={{
              display: 'flex',
              gap: 6,
              padding: '4px 6px',
              borderTop: '1px solid var(--border)',
              background: 'var(--bg-elevated)',
            }}
          >
            <button
              type="button"
              onClick={clear}
              style={{
                background: 'none',
                border: 'none',
                cursor: 'pointer',
                padding: '3px 6px',
                fontSize: '0.72rem',
                color: 'var(--text-2)',
                display: 'flex',
                alignItems: 'center',
                gap: 3,
              }}
            >
              <RotateCcw size={12} /> Clear
            </button>
            {hasSignature && (
              <button
                type="button"
                onClick={save}
                style={{
                  background: 'var(--brand)',
                  border: 'none',
                  cursor: 'pointer',
                  padding: '3px 8px',
                  fontSize: '0.72rem',
                  color: '#fff',
                  borderRadius: 5,
                  display: 'flex',
                  alignItems: 'center',
                  gap: 3,
                }}
              >
                <Check size={12} /> Simpan
              </button>
            )}
          </div>
        )}
      </div>
    </div>
  );
}
