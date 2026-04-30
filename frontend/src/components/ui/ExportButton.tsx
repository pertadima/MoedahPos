'use client';

import { useState } from 'react';
import { Download, Loader2 } from 'lucide-react';
import { reportsApi } from '@/lib/api/store-apis';

export type ExportReportType = 'sales' | 'inventory' | 'profit';
export type ExportFileType = 'csv' | 'pdf';

interface ExportButtonProps {
  storeId: string;
  report: ExportReportType;
  fileType: ExportFileType;
  dateFrom?: string;
  dateTo?: string;
  /** Optional label override; defaults to "Export {fileType.toUpperCase()}" */
  label?: string;
  /** Extra CSS classes */
  className?: string;
  /** Called after a successful export */
  onSuccess?: () => void;
  /** Called when the export fails */
  onError?: (err: Error) => void;
}

/**
 * ExportButton — triggers a file download for the given report & format.
 *
 * Displays a loading spinner while the request is in-flight and reverts
 * to the idle state once the download is complete (or on error).
 */
export default function ExportButton({
  storeId,
  report,
  fileType,
  dateFrom,
  dateTo,
  label,
  className = '',
  onSuccess,
  onError,
}: ExportButtonProps) {
  const [loading, setLoading] = useState(false);

  const handleClick = async () => {
    if (loading) return;
    setLoading(true);
    try {
      await reportsApi.exportReport(storeId, fileType, report, dateFrom, dateTo);
      onSuccess?.();
    } catch (err) {
      const error = err instanceof Error ? err : new Error(String(err));
      onError?.(error);
    } finally {
      setLoading(false);
    }
  };

  const defaultLabel = label ?? `Export ${fileType.toUpperCase()}`;

  return (
    <button
      type="button"
      onClick={handleClick}
      disabled={loading}
      className={[
        'inline-flex items-center gap-1.5 px-4 text-xs font-semibold shadow-sm h-[44px]',
        'border border-[var(--border)] bg-[var(--bg-card)] text-[var(--text-1)]',
        'hover:bg-[var(--bg-elevated)] transition-colors rounded-[22px]',
        'disabled:opacity-50 disabled:cursor-not-allowed',
        className,
      ]
        .filter(Boolean)
        .join(' ')}
    >
      {loading ? <Loader2 size={13} className="animate-spin" /> : <Download size={13} />}
      {defaultLabel}
    </button>
  );
}
