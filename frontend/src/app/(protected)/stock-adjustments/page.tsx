'use client';

import { useState, useEffect } from 'react';
import { useAuth } from '@/lib/auth/AuthContext';
import { api } from '@/lib/api/client';
import { stockAdjustmentApi, type StockAdjustment, type CreateAdjustmentInput } from '@/lib/api/stock-adjustments';
import { ClipboardList, Plus, AlertCircle, ArrowUpCircle, ArrowDownCircle, Info } from 'lucide-react';

interface Product {
  id: string;
  name: string;
  sku: string;
  unit: string;
}

export default function StockAdjustmentsPage() {
  const { selectedStore } = useAuth();
  const storeId = selectedStore?.store_id;
  const role = selectedStore?.role;
  const canUpdateStock = ['superadmin', 'admin', 'manager'].includes(role || '');

  const [adjustments, setAdjustments] = useState<StockAdjustment[]>([]);
  const [products, setProducts] = useState<Product[]>([]);
  const [loading, setLoading] = useState(true);
  const [errorHeader, setErrorHeader] = useState<string | null>(null);
  
  // Modal state
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [formData, setFormData] = useState<CreateAdjustmentInput>({
    product_id: '',
    type: 'OUT',
    reason: 'DAMAGED',
    quantity: 1,
    notes: '',
  });

  const fetchData = async () => {
    if (!storeId) return;
    try {
      setLoading(true);
      setErrorHeader(null);
      const [adjRes, prodRes] = await Promise.all([
        stockAdjustmentApi.getHistory(storeId),
        api.get<Product[]>(`/stores/${storeId}/products`),
      ]);
      setAdjustments(adjRes.data || []);
      setProducts(prodRes.data || []);
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Gagal memuat data';
      setErrorHeader(msg);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
  }, [storeId]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!storeId) return;
    try {
      setSubmitting(true);
      await stockAdjustmentApi.create(storeId, formData);
      setIsModalOpen(false);
      setFormData({
        product_id: '',
        type: 'OUT',
        reason: 'DAMAGED',
        quantity: 1,
        notes: '',
      });
      fetchData();
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Gagal menyimpan penyesuaian';
      alert(msg);
    } finally {
      setSubmitting(false);
    }
  };

  const getTypeLabel = (type: string) => {
    return type === 'IN' ? (
      <span style={{ color: 'var(--success)', display: 'flex', alignItems: 'center', gap: 4 }}>
        <ArrowUpCircle size={14} /> IN
      </span>
    ) : (
      <span style={{ color: 'var(--danger)', display: 'flex', alignItems: 'center', gap: 4 }}>
        <ArrowDownCircle size={14} /> OUT
      </span>
    );
  };

  const getReasonLabel = (reason: string) => {
    switch (reason) {
      case 'DAMAGED': return 'Rusak';
      case 'LOST': return 'Hilang';
      case 'MANUAL_CORRECTION': return 'Koreksi Manual';
      default: return reason;
    }
  };

  if (loading) {
    return <div className="page-container"><p>Memuat data...</p></div>;
  }

  return (
    <div className="page-container fade-in">
      {/* Header */}
      <header className="page-header">
        <div>
          <h1 className="page-title">
            <ClipboardList size={22} color="var(--brand)" />
            Penyesuaian Stok
          </h1>
          <p className="page-subtitle">
            Catat barang rusak, hilang, atau koreksi sistem.
          </p>
        </div>
        {canUpdateStock && (
          <button 
            className="btn btn-primary btn-sm"
            onClick={() => setIsModalOpen(true)}
          >
            <Plus size={16} /> Buat Penyesuaian
          </button>
        )}
      </header>

      {errorHeader && (
        <div style={{ padding: '12px 16px', background: 'rgba(239,68,68,0.1)', color: 'var(--danger)', borderRadius: 8, marginBottom: 20, display: 'flex', alignItems: 'center', gap: 8 }}>
          <AlertCircle size={18} />
          {errorHeader}
        </div>
      )}

      {/* Adjustments Table */}
      <div className="card" style={{ padding: 0, overflow: 'hidden' }}>
        {adjustments.length === 0 ? (
          <div style={{ padding: 40, textAlign: 'center', color: 'var(--text-3)' }}>
            <ClipboardList size={40} style={{ opacity: 0.3, margin: '0 auto 10px' }} />
            Belum ada riwayat penyesuaian stok.
          </div>
        ) : (
          <table className="table" style={{ width: '100%', borderCollapse: 'collapse' }}>
            <thead>
              <tr style={{ background: 'var(--bg-elevated)', borderBottom: '1px solid var(--border)' }}>
                <th style={{ padding: '12px 16px', textAlign: 'left', fontSize: '0.8rem', color: 'var(--text-2)', fontWeight: 600 }}>Tanggal</th>
                <th style={{ padding: '12px 16px', textAlign: 'left', fontSize: '0.8rem', color: 'var(--text-2)', fontWeight: 600 }}>Produk</th>
                <th style={{ padding: '12px 16px', textAlign: 'left', fontSize: '0.8rem', color: 'var(--text-2)', fontWeight: 600 }}>Tipe</th>
                <th style={{ padding: '12px 16px', textAlign: 'left', fontSize: '0.8rem', color: 'var(--text-2)', fontWeight: 600 }}>Alasan</th>
                <th style={{ padding: '12px 16px', textAlign: 'right', fontSize: '0.8rem', color: 'var(--text-2)', fontWeight: 600 }}>Kuantitas</th>
                <th style={{ padding: '12px 16px', textAlign: 'left', fontSize: '0.8rem', color: 'var(--text-2)', fontWeight: 600 }}>Staf</th>
              </tr>
            </thead>
            <tbody>
              {adjustments.map(adj => (
                <tr key={adj.id} style={{ borderBottom: '1px solid var(--border-light)' }}>
                  <td style={{ padding: '12px 16px', fontSize: '0.85rem' }}>
                    {new Intl.DateTimeFormat('id-ID', { day: '2-digit', month: 'short', year: 'numeric', hour: '2-digit', minute: '2-digit' }).format(new Date(adj.created_at))}
                  </td>
                  <td style={{ padding: '12px 16px' }}>
                    <div style={{ fontWeight: 500, color: 'var(--text-1)' }}>{adj.product_name}</div>
                    <div style={{ fontSize: '0.75rem', color: 'var(--text-3)' }}>SKU: {adj.product_sku}</div>
                  </td>
                  <td style={{ padding: '12px 16px', fontWeight: 600, fontSize: '0.8rem' }}>
                    {getTypeLabel(adj.type)}
                  </td>
                  <td style={{ padding: '12px 16px', fontSize: '0.85rem' }}>
                    <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
                      {getReasonLabel(adj.reason)}
                      {adj.notes && (
                        <div title={adj.notes} style={{ color: 'var(--text-3)', cursor: 'help' }}>
                          <Info size={14} />
                        </div>
                      )}
                    </div>
                  </td>
                  <td style={{ padding: '12px 16px', textAlign: 'right', fontWeight: 600, color: 'var(--text-1)' }}>
                    {adj.quantity} {adj.unit}
                  </td>
                  <td style={{ padding: '12px 16px', fontSize: '0.85rem', color: 'var(--text-2)' }}>
                    {adj.created_by_name}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>

      {/* Create Modal */}
      {isModalOpen && (
        <div className="modal-backdrop">
          <div className="modal-content" style={{ maxWidth: 500 }}>
            <h3 style={{ margin: '0 0 16px', color: 'var(--text-1)' }}>Catat Penyesuaian Stok</h3>
            
            <form onSubmit={handleSubmit} style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
              <div>
                <label style={{ display: 'block', marginBottom: '6px', fontSize: '0.8rem', fontWeight: 600 }}>
                  Produk <span style={{ color: 'var(--danger)' }}>*</span>
                </label>
                <select 
                  className="input" 
                  required
                  value={formData.product_id}
                  onChange={(e) => setFormData({ ...formData, product_id: e.target.value })}
                >
                  <option value="">-- Pilih Produk --</option>
                  {products.map(p => (
                    <option key={p.id} value={p.id}>{p.name} ({p.sku})</option>
                  ))}
                </select>
              </div>

              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '16px' }}>
                <div>
                  <label style={{ display: 'block', marginBottom: '6px', fontSize: '0.8rem', fontWeight: 600 }}>Tipe Permintaan</label>
                  <select 
                    className="input" 
                    value={formData.type}
                    onChange={(e) => {
                      const newType = e.target.value as 'IN' | 'OUT';
                      setFormData({ 
                        ...formData, 
                        type: newType,
                        reason: newType === 'IN' ? 'MANUAL_CORRECTION' : 'DAMAGED'
                      });
                    }}
                  >
                    <option value="OUT">Keluar (OUT)</option>
                    <option value="IN">Masuk (IN)</option>
                  </select>
                </div>
                
                <div>
                  <label style={{ display: 'block', marginBottom: '6px', fontSize: '0.8rem', fontWeight: 600 }}>Alasan</label>
                  <select 
                    className="input" 
                    value={formData.reason}
                    onChange={(e) => setFormData({ ...formData, reason: e.target.value as CreateAdjustmentInput['reason'] })}
                  >
                    {formData.type === 'OUT' && (
                      <>
                        <option value="DAMAGED">Barang Rusak</option>
                        <option value="LOST">Barang Hilang</option>
                      </>
                    )}
                    <option value="MANUAL_CORRECTION">Koreksi Manual</option>
                  </select>
                </div>
              </div>

              <div>
                <label style={{ display: 'block', marginBottom: '6px', fontSize: '0.8rem', fontWeight: 600 }}>
                  Kuantitas <span style={{ color: 'var(--danger)' }}>*</span>
                </label>
                <input 
                  type="number" 
                  min="0.1" 
                  step="0.1"
                  className="input" 
                  required
                  value={formData.quantity}
                  onChange={(e) => setFormData({ ...formData, quantity: parseFloat(e.target.value) })}
                />
              </div>

              <div>
                <label style={{ display: 'block', marginBottom: '6px', fontSize: '0.8rem', fontWeight: 600 }}>
                  Catatan {formData.reason === 'MANUAL_CORRECTION' && <span style={{ color: 'var(--danger)' }}>*</span>}
                </label>
                <textarea 
                  className="input" 
                  rows={3}
                  required={formData.reason === 'MANUAL_CORRECTION'}
                  value={formData.notes}
                  placeholder={formData.reason === 'MANUAL_CORRECTION' ? 'Wajib isi alasan detail' : 'Opsional'}
                  onChange={(e) => setFormData({ ...formData, notes: e.target.value })}
                />
              </div>

              <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '8px', marginTop: '8px' }}>
                <button 
                  type="button" 
                  className="btn btn-ghost" 
                  onClick={() => setIsModalOpen(false)}
                  disabled={submitting}
                >
                  Batal
                </button>
                <button 
                  type="submit" 
                  className="btn btn-primary"
                  disabled={submitting || !formData.product_id}
                >
                  {submitting ? 'Menyimpan...' : 'Simpan Penyesuaian'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
