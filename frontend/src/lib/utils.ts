import { type ClassValue, clsx } from 'clsx';
import { twMerge } from 'tailwind-merge';

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

/** Format number as Indonesian Rupiah */
export function formatRp(amount: number): string {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
  }).format(amount);
}

/** Format a datetime string to readable local format */
export function formatDate(iso: string): string {
  return new Intl.DateTimeFormat('id-ID', {
    day: '2-digit',
    month: 'short',
    year: 'numeric',
  }).format(new Date(iso));
}

export function formatDateTime(iso: string): string {
  return new Intl.DateTimeFormat('id-ID', {
    day: '2-digit',
    month: 'short',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  }).format(new Date(iso));
}

/** Today as YYYY-MM-DD */
export function todayStr(): string {
  return new Date().toISOString().slice(0, 10);
}

/** 30 days ago as YYYY-MM-DD */
export function thirtyDaysAgoStr(): string {
  const d = new Date();
  d.setDate(d.getDate() - 30);
  return d.toISOString().slice(0, 10);
}

/** Truncate a UUID to shorter display form */
export function shortId(id: string): string {
  return id.slice(0, 8).toUpperCase();
}

/** Get emoji icon for product based on name heuristic */
export function productEmoji(name: string): string {
  const n = name.toLowerCase();
  if (
    n.includes('coffee') ||
    n.includes('kopi') ||
    n.includes('americano') ||
    n.includes('espresso')
  )
    return '☕';
  if (n.includes('tea') || n.includes('teh')) return '🍵';
  if (n.includes('juice') || n.includes('jus')) return '🧃';
  if (n.includes('water') || n.includes('air')) return '💧';
  if (n.includes('milk') || n.includes('susu')) return '🥛';
  if (n.includes('cake') || n.includes('kue')) return '🎂';
  if (n.includes('bread') || n.includes('roti')) return '🍞';
  if (n.includes('rice') || n.includes('nasi')) return '🍚';
  if (n.includes('chicken') || n.includes('ayam')) return '🍗';
  if (n.includes('snack') || n.includes('chips')) return '🍟';
  return '📦';
}

/** Formats a number input field text value with thousands separators (dot localized per id-ID). Uses empty string logic for clean emptying */
export function formatNumberInput(value: number | string): string {
  if (value === 0 || value === '0') return '0';
  const numString = String(value).replace(/\D/g, '');
  if (!numString) return '';
  return new Intl.NumberFormat('id-ID').format(Number(numString));
}

/** Reverses a specifically formatted id-ID thousands separated number string back to an integer primitive */
export function parseNumberInput(value: string): number {
  const numString = value.replace(/\D/g, '');
  if (!numString) return 0;
  return Number(numString);
}
