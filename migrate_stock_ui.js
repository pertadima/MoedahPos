const fs = require('fs');
const path = require('path');

const srcPath = path.resolve(__dirname, 'frontend/src/app/(protected)/stock/page.tsx');
let currContent = fs.readFileSync(srcPath, 'utf8');

// 1. Imports
currContent = currContent.replace(
  `  Warehouse,\n} from 'lucide-react';`,
  `  Warehouse,\n  Search,\n  Check,\n  Plus,\n} from 'lucide-react';`
);
currContent = currContent.replace(
  `import { stockApi } from '@/lib/api/store-apis';`,
  `import { stockApi } from '@/lib/api/store-apis';\nimport { api } from '@/lib/api/client';\nimport { stockAdjustmentApi, type CreateAdjustmentInput } from '@/lib/api/stock-adjustments';`
);

// 2. Types
currContent = currContent.replace(
  `type Tab = 'stok' | 'movements';`,
  `type Tab = 'stok' | 'movements';\n\ninterface Product {\n  id: string;\n  name: string;\n  sku: string;\n  unit: string;\n}`
);

// 3. State & Logic
const stateCode = `
  const storeId = selectedStore?.store_id;
  const role = selectedStore?.role;
  const canUpdateStock = ['superadmin', 'admin', 'manager'].includes(role || '');

  // ── Modal state ─────────────────────────────────────────────────────────────
  const [products, setProducts] = useState<Product[]>([]);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [formData, setFormData] = useState<CreateAdjustmentInput>({
    product_id: '',
    type: 'OUT',
    reason: 'DAMAGED',
    quantity: 1,
    notes: '',
  });

  const [productSearch, setProductSearch] = useState('');
  const [debouncedSearch, setDebouncedSearch] = useState('');
  const [showDropdown, setShowDropdown] = useState(false);
  const [isSearchingProducts, setIsSearchingProducts] = useState(false);

  useEffect(() => {
    const handler = setTimeout(() => {
      setDebouncedSearch(productSearch);
    }, 300);
    return () => clearTimeout(handler);
  }, [productSearch]);

  useEffect(() => {
    if (!storeId) return;
    const fetchSearchedProducts = async () => {
      try {
        setIsSearchingProducts(true);
        const params = new URLSearchParams({ per_page: '20' });
        if (debouncedSearch && !formData.product_id) {
          params.append('search', debouncedSearch);
        }
        const res = await api.get<Product[]>(\`/stores/\${storeId}/products?\${params.toString()}\`);
        const prodData = Array.isArray(res?.data) ? res.data : (res?.data as any)?.data || [];
        setProducts(prodData);
      } catch (err) {
        console.error(err);
      } finally {
        setIsSearchingProducts(false);
      }
    };
    fetchSearchedProducts();
  }, [storeId, debouncedSearch, formData.product_id]);

  const handleAdjustmentSubmit = async (e: React.FormEvent) => {
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
      setProductSearch('');
      setShowDropdown(false);
      loadAll();
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Gagal menyimpan penyesuaian';
      alert(msg);
    } finally {
      setSubmitting(false);
    }
  };
`;
currContent = currContent.replace(`  const storeId = selectedStore?.store_id;`, stateCode);

// 4. Header Button
currContent = currContent.replace(
  `<div style={{ marginBottom: 20 }}>\n        <h1 className="page-title">Manajemen Stok</h1>\n        <p className="page-subtitle">\n          {selectedStore.store_name} · {levels.length} produk\n          {lowCount > 0 ? \` · ⚠ \${lowCount} stok menipis\` : ''}\n        </p>\n      </div>`,
  `<div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 20 }}>\n        <div>\n          <h1 className="page-title">Manajemen Stok</h1>\n          <p className="page-subtitle">\n            {selectedStore.store_name} · {levels.length} produk\n            {lowCount > 0 ? \` · ⚠ \${lowCount} stok menipis\` : ''}\n          </p>\n        </div>\n        {canUpdateStock && (\n          <button className="btn btn-primary btn-sm" onClick={() => setIsModalOpen(true)}>\n            <Plus size={16} /> Buat Penyesuaian\n          </button>\n        )}\n      </div>`
);

// 5. Read modal from stock-adjustments/page.tsx
const oldPagePath = path.resolve(__dirname, 'frontend/src/app/(protected)/stock-adjustments/page.tsx');
const oldPage = fs.readFileSync(oldPagePath, 'utf8');

// The modal starts from `{/* Create Modal */}` (let's find it).
// Wait, I didn't verify the exact marker. Let's just find `<div className="card"` under `isModalOpen && (`.
let modalJSX = oldPage.substring(oldPage.indexOf('{isModalOpen && ('), oldPage.lastIndexOf('</div>\n  );\n}'));

// Rename handleSubmit to handleAdjustmentSubmit
modalJSX = modalJSX.replace(/onSubmit=\{handleSubmit\}/g, "onSubmit={handleAdjustmentSubmit}");

// Insert modalJSX into currContent
currContent = currContent.replace(`      )}\n    </div>\n  );\n}`, `      )}\n\n      {/* ── Create Modal ─────────────────────────────────────────────────────── */}\n      ${modalJSX}\n    </div>\n  );\n}`);

fs.writeFileSync(srcPath, currContent, 'utf8');
console.log('Successfully injected stock/page.tsx');
