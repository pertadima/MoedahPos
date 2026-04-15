'use client';

import React, { useState, useRef, useEffect, useMemo } from 'react';
import { Calendar as CalendarIcon, ChevronLeft, ChevronRight, Zap } from 'lucide-react';

interface DatePickerProps {
  value: string; // YYYY-MM-DD
  onChange: (value: string) => void;
  placeholder?: string;
  className?: string;
  showPresets?: boolean;
  variant?: 'input' | 'minimal' | 'ghost';
}

const MONTH_NAMES = [
  'Januari',
  'Februari',
  'Maret',
  'April',
  'Mei',
  'Juni',
  'Juli',
  'Agustus',
  'September',
  'Oktober',
  'November',
  'Desember',
];

const WEEK_DAYS = ['M', 'S', 'S', 'R', 'K', 'J', 'S'];

const getTodayStr = () => {
  const d = new Date();
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`;
};

const getRelativeDateStr = (daysAgo: number) => {
  const d = new Date();
  d.setDate(d.getDate() - daysAgo);
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`;
};

export default function DatePicker({
  value,
  onChange,
  placeholder = 'Pilih Tanggal',
  className = '',
  showPresets = true,
  variant = 'input',
}: DatePickerProps) {
  const [isOpen, setIsOpen] = useState(false);
  const [popoverSide, setPopoverSide] = useState<'left' | 'right'>('left');

  const initialDate = useMemo(() => {
    if (!value) return new Date();
    const parts = value.split('-');
    if (parts.length !== 3) return new Date();
    const [y, m, d] = parts.map(Number);
    if (isNaN(y) || isNaN(m) || isNaN(d)) return new Date();
    return new Date(y, m - 1, d, 12, 0, 0);
  }, [value]);

  const [viewDate, setViewDate] = useState(initialDate);
  const containerRef = useRef<HTMLDivElement>(null);

  const handleOpen = () => {
    const nextState = !isOpen;
    if (nextState && containerRef.current) {
      // Determine side immediately on click
      const rect = containerRef.current.getBoundingClientRect();
      const screenWidth = window.innerWidth;
      const side = rect.left > screenWidth / 2 ? 'right' : 'left';
      setPopoverSide(side);
      setViewDate(initialDate);
    }
    setIsOpen(nextState);
  };

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (containerRef.current && !containerRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    };
    if (isOpen) {
      document.addEventListener('mousedown', handleClickOutside);
    }
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, [isOpen]);

  const year = viewDate.getFullYear();
  const month = viewDate.getMonth();

  const daysInMonth = new Date(year, month + 1, 0).getDate();
  const firstDayOfMonth = new Date(year, month, 1).getDay();
  const prevMonthDays = new Date(year, month, 0).getDate();

  const handlePrevMonth = (e: React.MouseEvent) => {
    e.stopPropagation();
    setViewDate(new Date(year, month - 1, 1));
  };

  const handleNextMonth = (e: React.MouseEvent) => {
    e.stopPropagation();
    setViewDate(new Date(year, month + 1, 1));
  };

  const handleDateSelect = (d: number, m?: number, y?: number) => {
    const targetY = y ?? year;
    const targetM = m ?? month;
    const newDate = new Date(targetY, targetM, d);
    const yyyy = newDate.getFullYear();
    const mm = String(newDate.getMonth() + 1).padStart(2, '0');
    const dd = String(newDate.getDate()).padStart(2, '0');
    onChange(`${yyyy}-${mm}-${dd}`);
    setIsOpen(false);
  };

  const isSelected = (d: number) => {
    if (!value) return false;
    const [y, m, dd] = value.split('-').map(Number);
    return y === year && m === month + 1 && dd === d;
  };

  const isToday = (d: number) => {
    const today = new Date();
    return today.getFullYear() === year && today.getMonth() === month && today.getDate() === d;
  };

  const formattedDisplayDate = useMemo(() => {
    if (!value) return '';
    const parts = value.split('-');
    if (parts.length !== 3) return '';
    const [y, m, d] = parts.map(Number);
    return `${d} ${MONTH_NAMES[m - 1]} ${y}`;
  }, [value]);

  const presets = [
    { label: 'Hari Ini', date: getTodayStr() },
    { label: 'Kemarin', date: getRelativeDateStr(1) },
    { label: '7 Hari Lalu', date: getRelativeDateStr(7) },
    { label: '30 Hari Lalu', date: getRelativeDateStr(30) },
  ];

  const getTriggerClass = () => {
    if (variant === 'input') {
      return 'input flex items-center gap-2 w-full h-full cursor-pointer hover:bg-surface-hv transition-all px-3';
    }
    if (variant === 'ghost') {
      return 'flex items-center gap-2 w-full h-full cursor-pointer hover:bg-surface-hv/50 transition-all px-2 rounded-md';
    }
    return 'flex items-center gap-2 w-full h-full cursor-pointer hover:bg-white/10 transition-all rounded transition-all transition-colors px-1.5';
  };

  return (
    <div className={`relative ${className}`} ref={containerRef}>
      <button
        type="button"
        onClick={handleOpen}
        className={getTriggerClass()}
        style={variant === 'input' ? { padding: '0 0.75rem' } : {}}
      >
        <CalendarIcon size={variant === 'minimal' ? 12 : 14} className="text-3 shrink-0" />
        <span
          className={`truncate ${variant === 'minimal' ? 'text-[11px]' : 'text-xs'} ${value ? 'text-1 font-bold' : 'text-3'}`}
        >
          {formattedDisplayDate || placeholder}
        </span>
      </button>

      {isOpen && (
        <div
          className={`absolute top-full mt-2 z-[9999] flex shadow-2xl rounded-2xl overflow-hidden animate-in fade-in slide-in-from-top-2 duration-150 ${popoverSide === 'right' ? 'right-0' : 'left-0'}`}
          style={{
            backdropFilter: 'blur(20px)',
            background: 'var(--bg-card)',
            border: '1px solid var(--border-md)',
            opacity: 1,
            minWidth: showPresets ? '420px' : '300px',
          }}
        >
          {/* Quick Presets Sidebar */}
          {showPresets && (
            <div className="w-36 bg-elevated/40 border-r border-[var(--border)] p-4 flex flex-col gap-1.5 shrink-0">
              <div className="flex items-center gap-2 text-[10px] font-black uppercase tracking-wider text-3 mb-3 px-1">
                <Zap size={10} className="text-accent-em" />
                CEPAT
              </div>
              {presets.map(p => (
                <button
                  key={p.label}
                  type="button"
                  onClick={() => {
                    const [y, m, d] = p.date.split('-').map(Number);
                    handleDateSelect(d, m - 1, y);
                  }}
                  className="w-full text-left px-3 py-2 rounded-lg text-xs font-semibold text-2 hover:bg-surface-hv hover:text-accent-em transition-all"
                >
                  {p.label}
                </button>
              ))}
            </div>
          )}

          <div className="p-6 flex-1">
            {/* Header */}
            <div className="flex justify-between items-center mb-6">
              <div className="text-sm font-black text-1 tracking-tight flex items-baseline gap-1">
                {MONTH_NAMES[month]}{' '}
                <span className="text-[var(--text-3)] font-medium text-xs ml-1">{year}</span>
              </div>
              <div className="flex gap-1.5">
                <button
                  onClick={handlePrevMonth}
                  className="p-1.5 hover:bg-surface-hv rounded-lg transition-colors border border-transparent hover:border-[var(--border-md)]"
                  type="button"
                >
                  <ChevronLeft size={16} />
                </button>
                <button
                  onClick={handleNextMonth}
                  className="p-1.5 hover:bg-surface-hv rounded-lg transition-colors border border-transparent hover:border-[var(--border-md)]"
                  type="button"
                >
                  <ChevronRight size={16} />
                </button>
              </div>
            </div>

            {/* Week Days */}
            <div className="grid grid-cols-7 mb-4">
              {WEEK_DAYS.map((d, i) => (
                <div
                  key={i}
                  className="text-center text-[10px] font-black text-3 uppercase tracking-tighter opacity-50"
                >
                  {d}
                </div>
              ))}
            </div>

            {/* Calendar Grid */}
            <div className="grid grid-cols-7 gap-1">
              {Array.from({ length: firstDayOfMonth }).map((_, i) => (
                <div
                  key={`prev-${i}`}
                  className="h-9 flex items-center justify-center text-[11px] text-3 opacity-15 pointer-events-none"
                >
                  {prevMonthDays - firstDayOfMonth + i + 1}
                </div>
              ))}

              {Array.from({ length: daysInMonth }).map((_, i) => {
                const d = i + 1;
                const selected = isSelected(d);
                const today = isToday(d);

                return (
                  <button
                    key={d}
                    type="button"
                    onClick={() => handleDateSelect(d)}
                    className={`h-9 w-9 flex items-center justify-center rounded-xl text-[12px] font-bold transition-all duration-200 relative group
                      ${
                        selected
                          ? 'text-white scale-105 z-10 shadow-lg'
                          : today
                            ? 'text-accent-em bg-accent-em/5 ring-1 ring-accent-em/20'
                            : 'text-2 hover:bg-surface-hv hover:text-1'
                      }
                    `}
                    style={
                      selected
                        ? {
                            background:
                              'linear-gradient(135deg, var(--accent-em), var(--accent-em-dk))',
                          }
                        : {}
                    }
                  >
                    {d}
                    {!selected && today && (
                      <span className="absolute bottom-1 right-1 w-1 h-1 rounded-full bg-accent-em" />
                    )}
                  </button>
                );
              })}

              {/* Pad with next month days */}
              {Array.from({
                length:
                  firstDayOfMonth + daysInMonth > 35
                    ? 42 - (firstDayOfMonth + daysInMonth)
                    : 35 - (firstDayOfMonth + daysInMonth),
              }).map((_, i) => (
                <div
                  key={`next-${i}`}
                  className="h-9 flex items-center justify-center text-[11px] text-3 opacity-15 pointer-events-none"
                >
                  {i + 1}
                </div>
              ))}
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
