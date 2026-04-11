'use client';

import { useEffect, useState, useCallback, useMemo } from 'react';
import {
  BarChart3,
  Loader2,
  TrendingUp,
  TrendingDown,
  Activity,
  DollarSign,
  CreditCard,
  Banknote,
  Smartphone,
  ArrowUpRight,
  Calendar,
  Layers,
  ChevronDown,
  ChevronRight,
} from 'lucide-react';
import { useAuth } from '@/lib/auth/AuthContext';
import { reportsApi } from '@/lib/api/store-apis';
import { transactionsApi } from '@/lib/api/transactions';
import { formatRp, todayStr, thirtyDaysAgoStr, formatDate } from '@/lib/utils';
import type {
  SalesSummaryResponse,
  SalesByProductRow,
  SalesSummaryRow,
  Transaction,
} from '@/types';
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  Tooltip,
  ResponsiveContainer,
  CartesianGrid,
  ComposedChart,
  Line,
  Area,
  Legend,
  ReferenceLine,
} from 'recharts';

// ── Types ──

type Tab = 'sales' | 'products' | 'profit' | 'cashflow' | 'valuation';
type GroupBy = 'day' | 'week' | 'month';

interface CashFlowDayRow {
  date: string;
  cash_in: number;
  cash_out: number;
  net_cash: number;
  cash_in_by_method: Record<string, number>;
}

interface CashFlowResponse {
  total_cash_in: number;
  total_cash_out: number;
  net_cash: number;
  cash_in_by_method: Record<string, number>;
  rows: CashFlowDayRow[];
}

interface StockValuationRow {
  product_id: string;
  product_name: string;
  sku: string;
  unit: string;
  cost_price: number;
  quantity: number;
  total_value: number;
}

interface StockValuationResponse {
  rows: StockValuationRow[];
  grand_total: number;
}

interface ProfitPeriodRow {
  period: string;
  total_sales: number;
  total_cost: number;
  gross_profit: number;
  total_expense: number;
  net_profit: number;
  profit_margin: number;
}

interface ProfitSummaryResponse {
  rows: ProfitPeriodRow[];
  total_sales: number;
  total_cost: number;
  gross_profit: number;
  total_expense: number;
  net_profit: number;
  profit_margin: number;
}

// ── Shared Components ──

function ProfitMarginBadge({ margin }: { margin: number }) {
  const color = margin >= 30 ? '#10b981' : margin >= 15 ? '#f59e0b' : '#ef4444';
  return (
    <span
      className="badge"
      style={{
        display: 'inline-flex',
        alignItems: 'center',
        gap: 2,
        padding: '2px 8px',
        borderRadius: 6,
        fontSize: '0.75rem',
        fontWeight: 700,
        background: `${color}18`,
        color,
      }}
    >
      {margin.toFixed(1)}%
    </span>
  );
}

const TOOLTIP_STYLE: React.CSSProperties = {
  background: 'var(--bg-card)',
  border: '1px solid var(--border)',
  borderRadius: 8,
  color: 'var(--text-1)',
  fontSize: 12,
  boxShadow: '0 4px 12px rgba(0,0,0,0.1)',
};

// Use a generic object for tooltip props to avoid Recharts version-specific Type conflicts
// while keeping the internal usage safe.
// eslint-disable-next-line @typescript-eslint/no-explicit-any
function CfTooltip({ active, payload, label }: any) {
  if (!active || !payload || payload.length === 0) return null;
  return (
    <div style={TOOLTIP_STYLE} className="p-3">
      <div style={{ fontWeight: 600, marginBottom: 6 }}>{label}</div>
      {/* eslint-disable-next-line @typescript-eslint/no-explicit-any */}
      {payload.map((p: any, i: number) => (
        <div
          key={p.name || i}
          style={{ display: 'flex', gap: 8, alignItems: 'center', marginBottom: 3 }}
        >
          <span
            style={{
              width: 10,
              height: 10,
              borderRadius: 2,
              background: (p.color as string) || (p.fill as string),
              display: 'inline-block',
            }}
          />
          <span style={{ color: 'var(--text-3)' }}>{p.name}:</span>
          <span style={{ fontWeight: 600 }}>{formatRp(Number(p.value) ?? 0)}</span>
        </div>
      ))}
    </div>
  );
}

function methodLabel(m: string) {
  const map: Record<string, string> = {
    cash: 'Tunai',
    qris: 'QRIS',
    transfer: 'Transfer',
    card: 'Kartu',
    debit: 'Debit',
    credit: 'Kredit',
  };
  return map[m.toLowerCase()] ?? m;
}

function methodIcon(m: string) {
  const lower = m.toLowerCase();
  if (lower === 'qris') return <Smartphone size={14} />;
  if (lower === 'transfer') return <CreditCard size={14} />;
  if (lower === 'cash') return <Banknote size={14} />;
  return <CreditCard size={14} />;
}

function UnifiedCard({
  label,
  value,
  icon: Icon,
  color,
  sub,
}: {
  label: string;
  value: number;
  icon: React.ElementType;
  color: string;
  sub?: string;
}) {
  const isNeg = value < 0;
  return (
    <div
      className="card"
      style={{
        padding: '16px 20px',
        display: 'flex',
        flexDirection: 'column',
        justifyContent: 'space-between',
        flex: 1,
        minWidth: 180,
        borderTop: `3px solid ${color}`,
      }}
    >
      <div className="flex justify-between items-start">
        <div>
          <div className="text-[10px] font-bold uppercase tracking-wider text-3">{label}</div>
          <div
            className="text-xl font-black mt-1"
            style={{ color: isNeg ? '#ef4444' : 'var(--text-1)' }}
          >
            {isNeg && '− '}
            {formatRp(Math.abs(value))}
          </div>
        </div>
        <div
          className="w-8 h-8 rounded-lg flex items-center justify-center"
          style={{ background: `${color}15`, color }}
        >
          <Icon size={18} />
        </div>
      </div>
      {sub && <div className="text-[10px] text-3 mt-2">{sub}</div>}
    </div>
  );
}

function ShoppingCartIcon(props: React.SVGProps<SVGSVGElement>) {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="24"
      height="24"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      {...props}
    >
      <circle cx="9" cy="21" r="1" />
      <circle cx="20" cy="21" r="1" />
      <path d="M1 1h4l2.68 13.39a2 2 0 0 0 2 1.61h9.72a2 2 0 0 0 2-1.61L23 6H6" />
    </svg>
  );
}

function WarehouseIcon(props: React.SVGProps<SVGSVGElement>) {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="24"
      height="24"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      {...props}
    >
      <path d="M3 21V10l9-7 9 7v11" />
      <path d="M9 21V9" />
      <path d="M15 21V9" />
    </svg>
  );
}

const TAB_CONFIG: { key: Tab; label: string; icon: React.ElementType }[] = [
  { key: 'sales', label: 'Penjualan', icon: ShoppingCartIcon },
  { key: 'products', label: 'Per Produk', icon: Layers },
  { key: 'profit', label: '💰 Profit', icon: Activity },
  { key: 'cashflow', label: '💵 Arus Kas', icon: DollarSign },
  { key: 'valuation', label: 'Stok', icon: WarehouseIcon },
];

export default function UnifiedReportsPage() {
  const { selectedStore } = useAuth();
  const storeId = selectedStore?.store_id;

  const [dateFrom, setDateFrom] = useState(thirtyDaysAgoStr());
  const [dateTo, setDateTo] = useState(todayStr());
  const [tab, setTab] = useState<Tab>('sales');
  const [groupBy, setGroupBy] = useState<GroupBy>('day');
  const [loading, setLoading] = useState(false);
  const [isMounted, setIsMounted] = useState(false);

  useEffect(() => {
    setIsMounted(true);
  }, []);

  const [summary, setSummary] = useState<SalesSummaryResponse | null>(null);
  const [byProduct, setByProduct] = useState<SalesByProductRow[]>([]);
  const [valuation, setValuation] = useState<StockValuationResponse | null>(null);
  const [profitData, setProfitData] = useState<ProfitSummaryResponse | null>(null);
  const [cfData, setCfData] = useState<CashFlowResponse | null>(null);
  const [expandedDates, setExpandedDates] = useState<Set<string>>(new Set());
  const [transactionsByDate, setTransactionsByDate] = useState<Record<string, Transaction[]>>({});
  const [loadingTransactions, setLoadingTransactions] = useState<Set<string>>(new Set());

  const fetchAll = useCallback(async () => {
    if (!storeId) return;
    setLoading(true);
    try {
      const results = await Promise.allSettled([
        reportsApi.salesSummary(storeId, dateFrom, dateTo),
        reportsApi.byProduct(storeId, dateFrom, dateTo),
        reportsApi.stockValuation(storeId),
        reportsApi.profit(storeId, dateFrom, dateTo, groupBy),
        reportsApi.cashFlow(storeId, dateFrom, dateTo),
      ]);

      if (results[0].status === 'fulfilled')
        setSummary(results[0].value.data as SalesSummaryResponse);
      if (results[1].status === 'fulfilled')
        setByProduct(results[1].value.data as SalesByProductRow[]);
      if (results[2].status === 'fulfilled')
        setValuation(results[2].value.data as StockValuationResponse);
      if (results[3].status === 'fulfilled')
        setProfitData(results[3].value.data as ProfitSummaryResponse);
      if (results[4].status === 'fulfilled') setCfData(results[4].value.data as CashFlowResponse);
    } catch (err) {
      console.error('Fetch error:', err);
    } finally {
      setLoading(false);
    }
  }, [storeId, dateFrom, dateTo, groupBy]);

  useEffect(() => {
    fetchAll();
  }, [fetchAll, storeId]);

  const toggleDateExpanded = useCallback(
    async (date: string) => {
      const now = new Set(expandedDates);
      if (now.has(date)) {
        now.delete(date);
        setExpandedDates(now);
      } else {
        // Fetch transactions for this date if not already loaded
        if (!transactionsByDate[date] && !loadingTransactions.has(date)) {
          setLoadingTransactions(prev => new Set(prev).add(date));
          try {
            const response = await transactionsApi.list(storeId || '', {
              date_from: date,
              date_to: date,
              per_page: 100,
            });
            setTransactionsByDate(prev => ({
              ...prev,
              [date]: response.data.data || [],
            }));
          } catch (err) {
            console.error(`Failed to load transactions for ${date}:`, err);
          } finally {
            setLoadingTransactions(prev => {
              const updated = new Set(prev);
              updated.delete(date);
              return updated;
            });
          }
        }
        now.add(date);
        setExpandedDates(now);
      }
    },
    [expandedDates, transactionsByDate, loadingTransactions, storeId]
  );

  const salesDataForChart = useMemo(
    () =>
      (summary?.rows ?? [])
        .slice()
        .sort((a, b) => new Date(a.date).getTime() - new Date(b.date).getTime())
        .map((r: SalesSummaryRow) => ({ date: r.date.slice(5), sales: r.total_sales })),
    [summary]
  );
  const profitChartData = useMemo(() => profitData?.rows ?? [], [profitData]);
  const cfChartData = useMemo(
    () =>
      (cfData?.rows ?? []).map(r => ({
        date: r.date.slice(5),
        'Uang Masuk': r.cash_in,
        'Uang Keluar': r.cash_out,
        'Net Cash': r.net_cash,
      })),
    [cfData]
  );
  const cfMethodEntries = useMemo(
    () => Object.entries(cfData?.cash_in_by_method ?? {}).sort((a, b) => b[1] - a[1]),
    [cfData]
  );

  if (!selectedStore) {
    return (
      <div className="p-8">
        <div className="empty-state card py-16 flex flex-col items-center gap-4">
          <BarChart3 size={48} className="opacity-20" />
          <p>Pilih toko terlebih dahulu</p>
        </div>
      </div>
    );
  }

  return (
    <div className="p-6 max-w-[1400px] mx-auto">
      <div className="flex justify-between items-start mb-6 flex-wrap gap-4">
        <div>
          <h1 className="text-2xl font-black flex items-center gap-2">📊 Laporan & Keuangan</h1>
          <p className="text-3 text-sm mt-1">
            {selectedStore.store_name} · Analisis performa & arus kas
          </p>
        </div>

        <div className="card flex items-center gap-2 p-1.5 px-3">
          <Calendar size={14} className="text-3" />
          <input
            type="date"
            className="input-subtle text-xs w-32 border-none bg-transparent"
            value={dateFrom}
            onChange={e => setDateFrom(e.target.value)}
          />
          <span className="text-3 text-[10px] opacity-40">TO</span>
          <input
            type="date"
            className="input-subtle text-xs w-32 border-none bg-transparent"
            value={dateTo}
            onChange={e => setDateTo(e.target.value)}
          />
          <button
            className="btn btn-primary btn-xs px-3 ml-2"
            onClick={fetchAll}
            disabled={loading}
          >
            {loading ? <Loader2 size={12} className="loading-spin" /> : 'Update Laporan'}
          </button>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-3 mb-6">
        {tab === 'cashflow' ? (
          <>
            <UnifiedCard
              label="Cash IN"
              value={cfData?.total_cash_in ?? 0}
              icon={TrendingUp}
              color="#10b981"
            />
            <UnifiedCard
              label="Cash OUT"
              value={cfData?.total_cash_out ?? 0}
              icon={TrendingDown}
              color="#ef4444"
            />
            <UnifiedCard
              label="Net Cash"
              value={cfData?.net_cash ?? 0}
              icon={ArrowUpRight}
              color={(cfData?.net_cash ?? 0) >= 0 ? '#3b82f6' : '#ef4444'}
            />
            <div className="card flex flex-col justify-center px-5 border-t-2 border-indigo-500">
              <div className="text-[10px] font-bold text-3 uppercase">Total Days</div>
              <div className="text-2xl font-black text-indigo-500 mt-1">
                {cfData?.rows.length ?? 0}
              </div>
            </div>
          </>
        ) : tab === 'valuation' ? (
          <UnifiedCard
            label="Total Valuasi Stok"
            value={valuation?.grand_total ?? 0}
            icon={WarehouseIcon}
            color="#6366f1"
          />
        ) : (
          <>
            <UnifiedCard
              label="Omzet"
              value={summary?.total_sales ?? 0}
              icon={TrendingUp}
              color="#10b981"
            />
            <UnifiedCard
              label="Laba Kotor"
              value={summary?.gross_profit ?? 0}
              icon={Activity}
              color="#f59e0b"
            />
            <UnifiedCard
              label="Laba Bersih"
              value={summary?.net_profit ?? 0}
              icon={Activity}
              color="#3b82f6"
            />
            <div className="card flex flex-col justify-center px-5 border-t-2 border-indigo-500">
              <div className="text-[10px] font-bold text-3 uppercase">Avg Margin</div>
              <div className="text-2xl font-black text-indigo-500 mt-1">
                {(summary?.profit_margin ?? 0).toFixed(1)}%
              </div>
            </div>
          </>
        )}
      </div>

      <div className="card inline-flex p-1 gap-1 mb-5 bg-surface border-none shadow-sm">
        {TAB_CONFIG.map(({ key, label, icon: TabIcon }) => (
          <button
            key={key}
            onClick={() => setTab(key)}
            className={`flex items-center gap-2 px-4 py-2 rounded-lg text-xs font-bold transition-all ${
              tab === key ? 'bg-accent-em text-white' : 'text-3 hover:bg-surface-hv'
            }`}
          >
            <TabIcon size={12} />
            {label}
          </button>
        ))}
      </div>

      {loading ? (
        <div className="flex justify-center py-32">
          <Loader2 size={32} className="loading-spin text-accent-em" />
        </div>
      ) : (
        <div className="animate-in fade-in slide-in-from-bottom-1 duration-300">
          {tab === 'sales' && (
            <div className="flex flex-col gap-5">
              <div className="card p-6">
                <h3 className="text-sm font-bold mb-6">📈 Grafik Omzet Harian</h3>
                <div className="w-full h-72">
                  {isMounted && (
                    <ResponsiveContainer>
                      <BarChart data={salesDataForChart}>
                        <CartesianGrid
                          strokeDasharray="3 3"
                          vertical={false}
                          stroke="var(--border)"
                        />
                        <XAxis
                          dataKey="date"
                          tick={{ fontSize: 10, fill: 'var(--text-3)' }}
                          axisLine={false}
                          tickLine={false}
                        />
                        <YAxis
                          tick={{ fontSize: 10, fill: 'var(--text-3)' }}
                          axisLine={false}
                          tickLine={false}
                          tickFormatter={(v: number) =>
                            v >= 1000000 ? `${v / 1000000}jt` : `${v / 1000}rb`
                          }
                        />
                        <Tooltip
                          contentStyle={TOOLTIP_STYLE}
                          // eslint-disable-next-line @typescript-eslint/no-explicit-any
                          formatter={(v: any) => [formatRp(Number(v || 0)), 'Omzet']}
                        />
                        <Bar dataKey="sales" fill="#10b981" radius={[4, 4, 0, 0]} />
                      </BarChart>
                    </ResponsiveContainer>
                  )}
                </div>
              </div>
              <div className="card overflow-hidden">
                <table className="tbl text-sm">
                  <thead>
                    <tr>
                      <th className="w-[30px]" />
                      <th className="w-[100px]">Tanggal</th>
                      <th className="!text-right w-[120px]">Total Penjualan</th>
                      <th className="!text-right w-[110px]">Total Pajak</th>
                      <th className="!text-right w-[130px]">Total Diskon</th>
                      <th className="!text-right w-[110px]">Jumlah Akhir</th>
                      <th className="!text-right w-[110px]">Total HPP</th>
                      <th className="!text-right w-[120px]">Laba Kotor</th>
                      <th className="!text-right w-[130px]">Total Pengeluaran</th>
                      <th className="!text-right w-[110px]">Laba Bersih</th>
                    </tr>
                  </thead>
                  <tbody>
                    {(summary?.rows ?? [])
                      .slice()
                      .sort((a, b) => new Date(b.date).getTime() - new Date(a.date).getTime())
                      .flatMap(r => {
                        const isExpanded = expandedDates.has(r.date);
                        const isLoading = loadingTransactions.has(r.date);
                        const transactions = transactionsByDate[r.date] || [];
                        const rows: React.ReactNode[] = [
                          <tr key={`date-${r.date}`}>
                            <td className="text-center">
                              <button
                                onClick={() => toggleDateExpanded(r.date)}
                                className="p-1 hover:bg-surface-hv rounded transition"
                              >
                                {isExpanded ? (
                                  <ChevronDown size={16} />
                                ) : (
                                  <ChevronRight size={16} />
                                )}
                              </button>
                            </td>
                            <td className="font-bold">{formatDate(r.date)}</td>
                            <td className="!text-right font-black text-accent-em">
                              {formatRp(r.total_sales ?? 0)}
                            </td>
                            <td className="!text-right opacity-75">{formatRp(r.total_tax ?? 0)}</td>
                            <td className="!text-right opacity-75">
                              {formatRp(r.total_discount ?? 0)}
                            </td>
                            <td className="!text-right font-semibold">
                              {formatRp((r.total_sales ?? 0) - (r.total_discount ?? 0))}
                            </td>
                            <td className="!text-right opacity-75">
                              {formatRp(r.total_cost ?? 0)}
                            </td>
                            <td
                              className={`!text-right font-semibold ${
                                (r.gross_profit ?? 0) >= 0 ? 'text-warning' : 'text-accent-rd'
                              }`}
                            >
                              {formatRp(r.gross_profit ?? 0)}
                            </td>
                            <td className="!text-right opacity-75">
                              {formatRp(r.total_expense ?? 0)}
                            </td>
                            <td
                              className={`!text-right font-bold ${
                                (r.net_profit ?? 0) >= 0 ? 'text-blue-500' : 'text-accent-rd'
                              }`}
                            >
                              {formatRp(r.net_profit ?? 0)}
                            </td>
                          </tr>,
                        ];

                        if (isExpanded) {
                          if (isLoading) {
                            rows.push(
                              <tr key={`loading-${r.date}`}>
                                <td colSpan={7} className="text-center py-4">
                                  <Loader2 size={16} className="loading-spin mx-auto" />
                                </td>
                              </tr>
                            );
                          } else if (transactions.length > 0) {
                            // Add a row with nested transaction table
                            rows.push(
                              <tr key={`nested-table-${r.date}`}>
                                <td colSpan={10} className="p-0">
                                  <div className="bg-surface overflow-hidden rounded">
                                    <table className="tbl text-xs w-full">
                                      <thead>
                                        <tr className="bg-surface-hv border-b border-border">
                                          <th className="w-[12%]">ID</th>
                                          <th className="w-[12%]">WAKTU</th>
                                          <th className="w-[16%]">KASIR</th>
                                          <th className="w-[14%]">PELANGGAN</th>
                                          <th className="w-[12%]">METODE</th>
                                          <th className="!text-right w-[13%]">SUBTOTAL</th>
                                          <th className="!text-right w-[13%]">DISKON</th>
                                          <th className="!text-right w-[13%]">PAJAK</th>
                                          <th className="!text-right w-[13%]">TOTAL</th>
                                          <th className="!text-center w-[6%]">STATUS</th>
                                        </tr>
                                      </thead>
                                      <tbody>
                                        {transactions.map((tx, txIdx) => (
                                          <tr
                                            key={`tx-${r.date}-${txIdx}`}
                                            className="border-b border-border hover:bg-surface-hv transition"
                                          >
                                            <td className="font-mono text-[10px] opacity-75">
                                              {tx.id?.slice(0, 8) || 'N/A'}
                                            </td>
                                            <td className="text-[10px] opacity-75">
                                              {tx.transaction_timestamp
                                                ? new Date(
                                                    tx.transaction_timestamp
                                                  ).toLocaleTimeString('id-ID', {
                                                    hour: '2-digit',
                                                    minute: '2-digit',
                                                    hour12: false,
                                                  })
                                                : 'N/A'}
                                            </td>
                                            <td className="text-[10px]">
                                              <div className="flex items-center gap-2">
                                                <div className="w-6 h-6 rounded-full bg-accent-em/20 text-accent-em flex items-center justify-center text-[9px] font-bold">
                                                  {tx.cashier_name?.charAt(0).toUpperCase() || 'A'}
                                                </div>
                                                <span className="opacity-75">
                                                  {tx.cashier_name || 'N/A'}
                                                </span>
                                              </div>
                                            </td>
                                            <td className="text-[10px] opacity-75">
                                              {tx.customer_name || '−'}
                                            </td>
                                            <td>
                                              <span className="inline-block px-2 py-1 bg-blue-500/20 text-blue-500 rounded text-[9px] font-medium">
                                                {tx.payment_method || 'N/A'}
                                              </span>
                                            </td>
                                            <td className="!text-right text-[10px] opacity-75">
                                              {formatRp(tx.subtotal || 0)}
                                            </td>
                                            <td className="!text-right text-[10px] opacity-75">
                                              {formatRp(tx.discount_amt || 0)}
                                            </td>
                                            <td className="!text-right text-[10px] opacity-75">
                                              {formatRp(tx.tax_amt || 0)}
                                            </td>
                                            <td className="!text-right font-semibold text-accent-em text-[10px]">
                                              {formatRp(tx.total || 0)}
                                            </td>
                                            <td className="!text-center">
                                              <span className="inline-flex items-center gap-1 text-[9px] font-bold opacity-60">
                                                <span className="w-2 h-2 rounded-full bg-green-500" />
                                                Selesai
                                              </span>
                                            </td>
                                          </tr>
                                        ))}
                                      </tbody>
                                    </table>
                                  </div>
                                </td>
                              </tr>
                            );
                          } else {
                            rows.push(
                              <tr key={`empty-${r.date}`}>
                                <td colSpan={7} className="text-center py-4 text-xs opacity-50">
                                  No transactions
                                </td>
                              </tr>
                            );
                          }
                        }

                        return rows;
                      })}
                  </tbody>
                </table>
              </div>
            </div>
          )}

          {tab === 'products' && (
            <div className="card overflow-hidden">
              <table className="tbl text-sm">
                <thead>
                  <tr>
                    <th className="w-[30%]">Nama Produk</th>
                    <th className="w-[15%]">SKU</th>
                    <th className="!text-right w-[10%]">Terjual</th>
                    <th className="!text-right w-[15%]">Revenue</th>
                    <th className="!text-right w-[15%]">Profit</th>
                    <th className="!text-center w-[15%]">Margin</th>
                  </tr>
                </thead>
                <tbody>
                  {byProduct.map((r, i) => (
                    <tr key={`${r.product_id}-${i}`}>
                      <td className="font-bold">{r.product_name}</td>
                      <td className="text-xs font-mono opacity-50">{r.sku}</td>
                      <td className="!text-right font-bold">{r.total_quantity}</td>
                      <td className="!text-right text-accent-em font-bold">
                        {formatRp(r.total_revenue)}
                      </td>
                      <td className="!text-right text-warning font-semibold">
                        {formatRp(r.gross_profit)}
                      </td>
                      <td className="!text-center">
                        <ProfitMarginBadge margin={r.profit_margin} />
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}

          {tab === 'profit' && (
            <div className="flex flex-col gap-5">
              <div className="flex items-center gap-2">
                {(['day', 'week', 'month'] as GroupBy[]).map(g => (
                  <button
                    key={g}
                    onClick={() => setGroupBy(g)}
                    className={`btn btn-xs px-4 rounded-full ${
                      groupBy === g ? 'btn-primary' : 'btn-outline'
                    }`}
                  >
                    {{ day: 'Harian', week: 'Mingguan', month: 'Bulanan' }[g]}
                  </button>
                ))}
              </div>
              <div className="card p-6">
                <h3 className="text-sm font-bold mb-6">📉 Profitability vs Expenses</h3>
                <div className="w-full h-80">
                  {isMounted && (
                    <ResponsiveContainer>
                      <ComposedChart data={profitChartData}>
                        <CartesianGrid strokeDasharray="3 3" vertical={false} />
                        <XAxis dataKey="period" tick={{ fontSize: 10 }} />
                        <YAxis
                          tick={{ fontSize: 10 }}
                          tickFormatter={(v: number) =>
                            v >= 1000000 ? `${v / 1000000}jt` : `${v / 1000}rb`
                          }
                        />
                        <Tooltip contentStyle={TOOLTIP_STYLE} />
                        <Legend
                          verticalAlign="top"
                          wrapperStyle={{ fontSize: 10, paddingBottom: 10 }}
                        />
                        <Area
                          type="monotone"
                          dataKey="total_sales"
                          fill="#10b98122"
                          stroke="#10b981"
                          name="Revenue"
                        />
                        <Bar
                          dataKey="gross_profit"
                          fill="#f59e0b"
                          radius={[4, 4, 0, 0]}
                          name="Gross"
                        />
                        <Bar dataKey="net_profit" fill="#3b82f6" radius={[4, 4, 0, 0]} name="Net" />
                        <Line
                          type="monotone"
                          dataKey="total_expense"
                          stroke="#ef4444"
                          name="Expenses"
                          strokeWidth={2}
                        />
                      </ComposedChart>
                    </ResponsiveContainer>
                  )}
                </div>
              </div>
              <div className="card overflow-hidden">
                <table className="tbl text-sm">
                  <thead>
                    <tr>
                      <th className="w-[120px]">Periode</th>
                      <th className="!text-right w-[150px]">Revenue</th>
                      <th className="!text-right w-[150px]">Cost</th>
                      <th className="!text-right w-[150px]">Expense</th>
                      <th className="!text-right w-[150px]">Net Profit</th>
                      <th className="!text-center w-[100px]">Margin</th>
                    </tr>
                  </thead>
                  <tbody>
                    {profitChartData
                      .slice()
                      .reverse()
                      .map((r, i) => (
                        <tr key={`${r.period}-${i}`}>
                          <td className="font-bold">{formatDate(r.period)}</td>
                          <td className="!text-right font-bold text-accent-em">
                            {formatRp(r.total_sales)}
                          </td>
                          <td className="!text-right text-accent-rd opacity-60">
                            {formatRp(r.total_cost)}
                          </td>
                          <td className="!text-right text-accent-rd">
                            {formatRp(r.total_expense)}
                          </td>
                          <td
                            className={`!text-right font-black ${
                              r.net_profit >= 0 ? 'text-blue-600' : 'text-accent-rd'
                            }`}
                          >
                            {formatRp(r.net_profit)}
                          </td>
                          <td className="!text-center">
                            <ProfitMarginBadge margin={r.profit_margin} />
                          </td>
                        </tr>
                      ))}
                  </tbody>
                </table>
              </div>
            </div>
          )}

          {tab === 'cashflow' && (
            <div className="flex flex-col gap-5">
              {cfMethodEntries.length > 0 && (
                <div className="flex flex-wrap gap-2">
                  {cfMethodEntries.map(([m, a]) => (
                    <div key={m} className="card p-3 px-4 flex items-center gap-3">
                      <div className="text-accent-em">{methodIcon(m)}</div>
                      <div>
                        <div className="text-[10px] font-bold uppercase text-3">
                          {methodLabel(m)}
                        </div>
                        <div className="font-bold">{formatRp(a)}</div>
                      </div>
                    </div>
                  ))}
                </div>
              )}
              <div className="card p-6">
                <h3 className="text-sm font-bold mb-6">📉 Tren Arus Kas Aktual</h3>
                <div className="w-full h-80">
                  {isMounted && (
                    <ResponsiveContainer>
                      <ComposedChart data={cfChartData}>
                        <CartesianGrid strokeDasharray="3 3" vertical={false} />
                        <XAxis dataKey="date" tick={{ fontSize: 10 }} />
                        <YAxis
                          tick={{ fontSize: 10 }}
                          tickFormatter={(v: number) =>
                            v >= 1000000 ? `${v / 1000000}jt` : `${v / 1000}rb`
                          }
                        />
                        <Tooltip content={<CfTooltip />} />
                        <Legend
                          verticalAlign="top"
                          wrapperStyle={{ fontSize: 10, paddingBottom: 15 }}
                        />
                        <ReferenceLine y={0} stroke="var(--border)" />
                        <Bar
                          dataKey="Uang Masuk"
                          fill="#10b981"
                          radius={[4, 4, 0, 0]}
                          maxBarSize={30}
                        />
                        <Bar
                          dataKey="Uang Keluar"
                          fill="#ef4444"
                          radius={[4, 4, 0, 0]}
                          maxBarSize={30}
                        />
                        <Line
                          type="monotone"
                          dataKey="Net Cash"
                          stroke="#3b82f6"
                          strokeWidth={3}
                          dot={{ r: 4 }}
                        />
                      </ComposedChart>
                    </ResponsiveContainer>
                  )}
                </div>
              </div>
              <div className="card overflow-hidden">
                <table className="tbl text-sm">
                  <thead>
                    <tr>
                      <th className="w-[120px]">Tanggal</th>
                      <th className="!text-right w-[150px]">Masuk</th>
                      <th className="!text-right w-[150px]">Keluar</th>
                      <th className="!text-right w-[150px]">Net</th>
                      <th>Metode Kas Masuk</th>
                    </tr>
                  </thead>
                  <tbody>
                    {(cfData?.rows ?? [])
                      .slice()
                      .reverse()
                      .map((r, i) => (
                        <tr key={`${r.date}-${i}`}>
                          <td className="font-bold">{formatDate(r.date)}</td>
                          <td className="!text-right text-accent-em font-bold">
                            {formatRp(r.cash_in)}
                          </td>
                          <td className="!text-right text-accent-rd">{formatRp(r.cash_out)}</td>
                          <td
                            className={`!text-right font-black ${
                              r.net_cash >= 0 ? 'text-blue-500' : 'text-accent-rd'
                            }`}
                          >
                            {formatRp(r.net_cash)}
                          </td>
                          <td className="text-[10px] opacity-60">
                            {Object.entries(r.cash_in_by_method)
                              .map(([m, a]) => `${methodLabel(m)}: ${formatRp(a)}`)
                              .join(' · ')}
                          </td>
                        </tr>
                      ))}
                  </tbody>
                </table>
              </div>
            </div>
          )}

          {tab === 'valuation' && (
            <div className="card overflow-hidden">
              <div className="p-5 border-b bg-indigo-500/5 flex justify-between items-end">
                <h3 className="font-black text-lg">💰 Valuasi Persediaan</h3>
                <div className="text-right">
                  <div className="text-[10px] font-bold uppercase text-3">Grand Total</div>
                  <div className="text-2xl font-black text-indigo-600">
                    {formatRp(valuation?.grand_total ?? 0)}
                  </div>
                </div>
              </div>
              <table className="tbl text-sm">
                <thead>
                  <tr>
                    <th className="w-[30%]">Produk</th>
                    <th className="w-[15%]">SKU</th>
                    <th className="!text-right w-[15%]">Stok</th>
                    <th className="!text-right w-[20%]">Harga Beli</th>
                    <th className="!text-right w-[20%]">Subtotal</th>
                  </tr>
                </thead>
                <tbody>
                  {(valuation?.rows ?? []).map((r, i) => (
                    <tr key={`${r.product_id}-${i}`}>
                      <td className="font-bold">{r.product_name}</td>
                      <td className="text-xs font-mono opacity-50">{r.sku}</td>
                      <td className="!text-right font-bold">
                        {r.quantity} {r.unit}
                      </td>
                      <td className="!text-right opacity-60">{formatRp(r.cost_price)}</td>
                      <td className="!text-right font-black text-indigo-500">
                        {formatRp(r.total_value)}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
