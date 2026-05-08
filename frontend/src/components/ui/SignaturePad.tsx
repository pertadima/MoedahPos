'use client';

import { useEffect, useRef, useState, useCallback } from 'react';
import { RotateCcw, Check } from 'lucide-react';

interface SignaturePadProps {
  value?: string;
  onSave: (dataUrl: string) => void;
  onClear?: () => void;
  label?: string;
  readOnly?: boolean;
}

export default function SignaturePad({
  value,
  onSave,
  onClear,
  label = 'Tanda Tangan',
  readOnly = false,
}: SignaturePadProps) {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const [isDrawing, setIsDrawing] = useState(false);
  const [hasSignature, setHasSignature] = useState(!!value);

  const getPos = useCallback((e: MouseEvent | TouchEvent) => {
    const canvas = canvasRef.current;
    if (!canvas) return { x: 0, y: 0 };
    const rect = canvas.getBoundingClientRect();
    if ('touches' in e) {
      const touch = e.touches[0];
      return { x: touch.clientX - rect.left, y: touch.clientY - rect.top };
    }
    return { x: e.clientX - rect.left, y: e.clientY - rect.top };
  }, []);

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    ctx.strokeStyle = '#000';
    ctx.lineWidth = 2;
    ctx.lineCap = 'round';

    if (value) {
      const img = new Image();
      img.onload = () => {
        ctx.clearRect(0, 0, canvas.width, canvas.height);
        ctx.drawImage(img, 0, 0);
      };
      img.src = value;
    }
  }, [value]);

  const startDraw = (e: React.MouseEvent | React.TouchEvent) => {
    if (readOnly) return;
    e.preventDefault();
    setIsDrawing(true);
    const pos = getPos(e.nativeEvent);
    const canvas = canvasRef.current;
    if (!canvas) return;
    const ctx = canvas.getContext('2d');
    if (!ctx) return;
    ctx.beginPath();
    ctx.moveTo(pos.x, pos.y);
  };

  const draw = (e: React.MouseEvent | React.TouchEvent) => {
    if (!isDrawing || readOnly) return;
    e.preventDefault();
    const canvas = canvasRef.current;
    if (!canvas) return;
    const ctx = canvas.getContext('2d');
    if (!ctx) return;
    const pos = getPos(e.nativeEvent);
    ctx.lineTo(pos.x, pos.y);
    ctx.stroke();
    setHasSignature(true);
  };

  const stopDraw = () => {
    setIsDrawing(false);
  };

  const clear = () => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    const ctx = canvas.getContext('2d');
    if (!ctx) return;
    ctx.clearRect(0, 0, canvas.width, canvas.height);
    setHasSignature(false);
    onClear?.();
  };

  const save = () => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    const dataUrl = canvas.toDataURL('image/png');
    onSave(dataUrl);
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
          position: 'relative',
          background: '#fff',
        }}
      >
        <canvas
          ref={canvasRef}
          width={300}
          height={100}
          style={{
            display: 'block',
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
