'use client';

import React, { useState, useEffect } from 'react';
import {
  api,
  Product,
  Supplier,
  PurchaseRequest,
  PurchaseOrder,
  GoodsReceipt,
  StockSummary,
  InventoryTransaction,
} from '@/lib/api';
import {
  Package,
  Users,
  FileText,
  Boxes,
  Activity,
  Plus,
  CheckCircle2,
  XCircle,
  Clock,
  ArrowUpRight,
  RefreshCw,
  Search,
  Filter,
  AlertCircle,
  Truck,
  Check,
  ChevronRight,
} from 'lucide-react';

export default function MissionControlERP() {
  const [activeTab, setActiveTab] = useState<'dashboard' | 'products' | 'suppliers' | 'procurement' | 'inventory'>('dashboard');
  const [backendReady, setBackendReady] = useState<boolean | null>(null);
  const [loading, setLoading] = useState(false);
  const [notification, setNotification] = useState<{ type: 'success' | 'error'; message: string } | null>(null);

  // Data states
  const [products, setProducts] = useState<Product[]>([]);
  const [suppliers, setSuppliers] = useState<Supplier[]>([]);
  const [prs, setPRs] = useState<PurchaseRequest[]>([]);
  const [pos, setPOs] = useState<PurchaseOrder[]>([]);
  const [goodsReceipts, setGoodsReceipts] = useState<GoodsReceipt[]>([]);
  const [stock, setStock] = useState<StockSummary[]>([]);
  const [ledger, setLedger] = useState<InventoryTransaction[]>([]);

  // Modals state
  const [showProductModal, setShowProductModal] = useState(false);
  const [showSupplierModal, setShowSupplierModal] = useState(false);
  const [showPRModal, setShowPRModal] = useState(false);
  const [showPOModal, setShowPOModal] = useState(false);
  const [showGRModal, setShowGRModal] = useState(false);

  // Form states
  const [productForm, setProductForm] = useState({ sku: '', name: '', unit: 'kg' });
  const [supplierForm, setSupplierForm] = useState({ name: '', contact_name: '', email: '', phone: '', address: '' });
  const [prForm, setPRForm] = useState({ notes: '', items: [{ product_id: 0, quantity: 1000, notes: '' }] });
  const [poForm, setPOForm] = useState({ purchase_request_id: 0, supplier_id: 0, notes: '', items: [{ product_id: 0, quantity: 1000, unit_price: 2500 }] });
  const [grForm, setGRForm] = useState({
    purchase_order_id: 0,
    delivery_order_number: '',
    notes: '',
    items: [] as { purchase_order_item_id: number; product_id: number; product_name?: string; received_quantity: number; accepted_quantity: number; rejected_quantity: number; rejection_reason: string }[],
  });

  const notify = (message: string, type: 'success' | 'error' = 'success') => {
    setNotification({ message, type });
    setTimeout(() => setNotification(null), 4000);
  };

  const loadAllData = async () => {
    setLoading(true);
    try {
      const isReady = await api.checkHealth();
      setBackendReady(isReady);

      if (isReady) {
        const [p, s, prList, poList, grList, stockList, ledgerList] = await Promise.all([
          api.getProducts().catch(() => []),
          api.getSuppliers().catch(() => []),
          api.getPRs().catch(() => []),
          api.getPOs().catch(() => []),
          api.getGoodsReceipts().catch(() => []),
          api.getStockSummary().catch(() => []),
          api.getLedger().catch(() => []),
        ]);
        setProducts(p);
        setSuppliers(s);
        setPRs(prList);
        setPOs(poList);
        setGoodsReceipts(grList);
        setStock(stockList);
        setLedger(ledgerList);
      }
    } catch {
      setBackendReady(false);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadAllData();
  }, []);

  // Handlers
  const handleCreateProduct = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await api.createProduct(productForm);
      notify('Produk berhasil ditambahkan');
      setShowProductModal(false);
      setProductForm({ sku: '', name: '', unit: 'kg' });
      loadAllData();
    } catch (err: any) {
      notify(err.message, 'error');
    }
  };

  const handleCreateSupplier = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await api.createSupplier(supplierForm);
      notify('Supplier berhasil didaftarkan');
      setShowSupplierModal(false);
      setSupplierForm({ name: '', contact_name: '', email: '', phone: '', address: '' });
      loadAllData();
    } catch (err: any) {
      notify(err.message, 'error');
    }
  };

  const handleCreatePR = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      if (prForm.items.some(i => i.product_id <= 0 || i.quantity <= 0)) {
        notify('Pilih produk dan masukkan kuantitas valid', 'error');
        return;
      }
      await api.createPR(prForm);
      notify('Purchase Request (PR) berhasil dibuat');
      setShowPRModal(false);
      setPRForm({ notes: '', items: [{ product_id: 0, quantity: 1000, notes: '' }] });
      loadAllData();
    } catch (err: any) {
      notify(err.message, 'error');
    }
  };

  const handlePRAction = async (id: number, action: 'submit' | 'approve' | 'reject') => {
    try {
      if (action === 'submit') await api.submitPR(id);
      if (action === 'approve') await api.approvePR(id);
      if (action === 'reject') await api.rejectPR(id);
      notify(`PR #${id} berhasil di-${action}`);
      loadAllData();
    } catch (err: any) {
      notify(err.message, 'error');
    }
  };

  const handleCreatePO = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      if (poForm.supplier_id <= 0) {
        notify('Pilih supplier terlebih dahulu', 'error');
        return;
      }
      await api.createPO({
        purchase_request_id: poForm.purchase_request_id > 0 ? poForm.purchase_request_id : undefined,
        supplier_id: poForm.supplier_id,
        notes: poForm.notes,
        items: poForm.items,
      });
      notify('Purchase Order (PO) berhasil dibuat');
      setShowPOModal(false);
      loadAllData();
    } catch (err: any) {
      notify(err.message, 'error');
    }
  };

  const handlePOAction = async (id: number, action: 'approve' | 'cancel') => {
    try {
      if (action === 'approve') await api.approvePO(id);
      if (action === 'cancel') await api.cancelPO(id);
      notify(`PO #${id} berhasil di-${action}`);
      loadAllData();
    } catch (err: any) {
      notify(err.message, 'error');
    }
  };

  const handleSelectPOForGR = (poId: number) => {
    const selectedPO = pos.find(p => p.id === poId);
    if (!selectedPO || !selectedPO.items) return;
    setGRForm({
      purchase_order_id: poId,
      delivery_order_number: '',
      notes: '',
      items: selectedPO.items.map(item => ({
        purchase_order_item_id: item.id,
        product_id: item.product_id,
        product_name: item.product_name,
        received_quantity: item.quantity,
        accepted_quantity: item.quantity,
        rejected_quantity: 0,
        rejection_reason: '',
      })),
    });
  };

  const handleCreateGR = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await api.createGoodsReceipt(grForm);
      notify('Goods Receipt berhasil dicatat & stok ledger diperbarui');
      setShowGRModal(false);
      loadAllData();
    } catch (err: any) {
      notify(err.message, 'error');
    }
  };

  // Calculations for dashboard
  const totalStockKg = stock.reduce((sum, item) => sum + item.current_stock, 0);
  const activeSuppliersCount = suppliers.filter(s => s.is_active).length;
  const approvedPOCount = pos.filter(p => p.status === 'APPROVED' || p.status === 'PARTIALLY_RECEIVED').length;

  return (
    <div style={{ minHeight: '100vh', display: 'flex', flexDirection: 'column' }}>
      {/* Top Notification Toast */}
      {notification && (
        <div
          style={{
            position: 'fixed',
            top: 24,
            right: 24,
            zIndex: 1000,
            padding: '12px 24px',
            borderRadius: 'var(--radius-buttons)',
            backgroundColor: notification.type === 'success' ? '#064e3b' : '#7f1d1d',
            color: '#fff',
            boxShadow: '0 10px 25px rgba(0,0,0,0.5)',
            display: 'flex',
            alignItems: 'center',
            gap: 10,
            fontSize: 14,
            border: notification.type === 'success' ? '1px solid #10b981' : '1px solid #ef4444',
          }}
        >
          {notification.type === 'success' ? <CheckCircle2 size={18} /> : <AlertCircle size={18} />}
          {notification.message}
        </div>
      )}

      {/* Top Bar Header Navigation (REKKI Mission Control aesthetic) */}
      <header
        style={{
          height: 80,
          borderBottom: '1px solid rgba(255, 255, 255, 0.08)',
          backgroundColor: 'rgba(0, 0, 0, 0.85)',
          backdropFilter: 'blur(12px)',
          position: 'sticky',
          top: 0,
          zIndex: 50,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          padding: '0 32px',
        }}
      >
        {/* Logo Mark: REKKI Wordmark + π Symbol */}
        <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
          <div
            style={{
              width: 32,
              height: 32,
              borderRadius: 6,
              backgroundColor: 'var(--color-signal-blue)',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              fontWeight: 700,
              fontSize: 18,
              color: '#fff',
            }}
          >
            π
          </div>
          <div>
            <div style={{ fontWeight: 600, fontSize: 16, letterSpacing: '-0.02em', display: 'flex', alignItems: 'center', gap: 8 }}>
              AGRO-SUPPLY ERP
              <span style={{ width: 6, height: 6, borderRadius: '50%', backgroundColor: 'var(--color-signal-blue)' }}></span>
            </div>
            <div style={{ fontSize: 11, color: 'var(--color-ash)', letterSpacing: '0.04em', textTransform: 'uppercase' }}>
              Control Room
            </div>
          </div>
        </div>

        {/* Center Nav Links */}
        <nav style={{ display: 'flex', alignItems: 'center', gap: 32 }}>
          {[
            { id: 'dashboard', label: 'Dashboard' },
            { id: 'products', label: 'Master Produk' },
            { id: 'suppliers', label: 'Master Supplier' },
            { id: 'procurement', label: 'Procurement (PR & PO)' },
            { id: 'inventory', label: 'Inventory & Ledger' },
          ].map(tab => (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id as any)}
              style={{
                fontSize: 15,
                fontWeight: activeTab === tab.id ? 600 : 400,
                color: activeTab === tab.id ? 'var(--color-paper)' : 'var(--color-ash)',
                letterSpacing: '-0.007em',
                position: 'relative',
                padding: '6px 0',
                transition: 'color 0.15s ease',
              }}
            >
              {tab.label}
              {activeTab === tab.id && (
                <div
                  style={{
                    position: 'absolute',
                    bottom: -2,
                    left: 0,
                    right: 0,
                    height: 2,
                    backgroundColor: 'var(--color-signal-blue)',
                    borderRadius: 2,
                  }}
                />
              )}
            </button>
          ))}
        </nav>

        {/* Right Status Indicator & Action CTA */}
        <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
          <div
            className="status-pill"
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: 8,
              padding: '6px 14px',
            }}
          >
            <span
              style={{
                width: 8,
                height: 8,
                borderRadius: '50%',
                backgroundColor: backendReady ? '#22c55e' : '#ef4444',
                boxShadow: backendReady ? '0 0 10px #22c55e' : '0 0 10px #ef4444',
              }}
            />
            {backendReady ? 'GO API :8080 CONNECTED' : 'BACKEND DISCONNECTED'}
          </div>

          <button
            onClick={loadAllData}
            className="btn-secondary"
            title="Refresh Data"
            style={{ padding: '8px 14px' }}
          >
            <RefreshCw size={15} className={loading ? 'animate-spin' : ''} />
          </button>
        </div>
      </header>

      {/* Main Body */}
      <main style={{ flex: 1, padding: '40px 32px', maxWidth: 'var(--page-max-width)', margin: '0 auto', width: '100%' }}>
        {/* ============================================================== */}
        {/* TAB 1: DASHBOARD VIEW                                           */}
        {/* ============================================================== */}
        {activeTab === 'dashboard' && (
          <div>
            <div style={{ marginBottom: 32 }}>
              <h1 style={{ fontSize: 'var(--text-heading)', fontWeight: 600, letterSpacing: 'var(--tracking-heading)' }}>
                Mission Control Dashboard
              </h1>
              <p style={{ color: 'var(--color-ash)', fontSize: 14 }}>
                Pemantauan rantai pasok bahan baku batok kelapa dari pengadaan hingga saldo stok gudang.
              </p>
            </div>

            {/* 4 Metric Instrument Panels */}
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: 20, marginBottom: 40 }}>
              <div className="card-panel">
                <div style={{ display: 'flex', justifyContent: 'space-between', color: 'var(--color-ash)', marginBottom: 12 }}>
                  <span style={{ fontSize: 13, textTransform: 'uppercase', fontFamily: 'var(--font-mono)' }}>Total Stok Fisik</span>
                  <Boxes size={18} color="var(--color-signal-blue)" />
                </div>
                <div style={{ fontSize: 32, fontWeight: 600, letterSpacing: '-0.03em', color: 'var(--color-paper)' }}>
                  {totalStockKg.toLocaleString()} <span style={{ fontSize: 16, color: 'var(--color-ash)' }}>kg</span>
                </div>
                <div style={{ fontSize: 12, color: 'var(--color-ash)', marginTop: 6 }}>
                  Akumulasi terverifikasi QC di gudang
                </div>
              </div>

              <div className="card-panel">
                <div style={{ display: 'flex', justifyContent: 'space-between', color: 'var(--color-ash)', marginBottom: 12 }}>
                  <span style={{ fontSize: 13, textTransform: 'uppercase', fontFamily: 'var(--font-mono)' }}>PO Dalam Proses</span>
                  <Truck size={18} color="var(--color-signal-blue)" />
                </div>
                <div style={{ fontSize: 32, fontWeight: 600, letterSpacing: '-0.03em', color: 'var(--color-paper)' }}>
                  {approvedPOCount}
                </div>
                <div style={{ fontSize: 12, color: 'var(--color-ash)', marginTop: 6 }}>
                  Menunggu atau sebagian dikirim truk
                </div>
              </div>

              <div className="card-panel">
                <div style={{ display: 'flex', justifyContent: 'space-between', color: 'var(--color-ash)', marginBottom: 12 }}>
                  <span style={{ fontSize: 13, textTransform: 'uppercase', fontFamily: 'var(--font-mono)' }}>Supplier Aktif</span>
                  <Users size={18} color="var(--color-signal-blue)" />
                </div>
                <div style={{ fontSize: 32, fontWeight: 600, letterSpacing: '-0.03em', color: 'var(--color-paper)' }}>
                  {activeSuppliersCount} <span style={{ fontSize: 16, color: 'var(--color-ash)' }}>/ {suppliers.length}</span>
                </div>
                <div style={{ fontSize: 12, color: 'var(--color-ash)', marginTop: 6 }}>
                  Pemasok yang siap menerima PO
                </div>
              </div>

              <div className="card-panel">
                <div style={{ display: 'flex', justifyContent: 'space-between', color: 'var(--color-ash)', marginBottom: 12 }}>
                  <span style={{ fontSize: 13, textTransform: 'uppercase', fontFamily: 'var(--font-mono)' }}>Katalog Produk</span>
                  <Package size={18} color="var(--color-signal-blue)" />
                </div>
                <div style={{ fontSize: 32, fontWeight: 600, letterSpacing: '-0.03em', color: 'var(--color-paper)' }}>
                  {products.length}
                </div>
                <div style={{ fontSize: 12, color: 'var(--color-ash)', marginTop: 6 }}>
                  Bahan baku & varian terdaftar
                </div>
              </div>
            </div>

            {/* Quick Action Panels & Workflow Visualizer */}
            <div style={{ display: 'grid', gridTemplateColumns: '2fr 1fr', gap: 24, marginBottom: 40 }}>
              <div className="card-panel">
                <h3 style={{ fontSize: 18, fontWeight: 600, marginBottom: 20 }}>
                  Ringkasan Saldo Stok Produk
                </h3>
                {stock.length === 0 ? (
                  <div style={{ color: 'var(--color-ash)', fontSize: 14, padding: '24px 0', textAlign: 'center' }}>
                    Belum ada saldo stok tercatat. Buat Goods Receipt untuk menambah stok.
                  </div>
                ) : (
                  <table className="data-table">
                    <thead>
                      <tr>
                        <th>SKU</th>
                        <th>NAMA PRODUK</th>
                        <th>TOTAL MASUK (IN)</th>
                        <th>TOTAL KELUAR (OUT)</th>
                        <th>SALDO SAAT INI</th>
                      </tr>
                    </thead>
                    <tbody>
                      {stock.map(s => (
                        <tr key={s.product_id}>
                          <td className="mono-cell" style={{ color: 'var(--color-signal-blue)' }}>{s.product_sku}</td>
                          <td style={{ fontWeight: 500 }}>{s.product_name}</td>
                          <td className="mono-cell">{s.total_in.toLocaleString()} {s.product_unit}</td>
                          <td className="mono-cell">{s.total_out.toLocaleString()} {s.product_unit}</td>
                          <td className="mono-cell" style={{ fontWeight: 600, color: '#4ade80' }}>
                            {s.current_stock.toLocaleString()} {s.product_unit}
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                )}
              </div>

              {/* Workflow Pipeline */}
              <div className="card-panel" style={{ display: 'flex', flexDirection: 'column', justifyContent: 'space-between' }}>
                <div>
                  <h3 style={{ fontSize: 18, fontWeight: 600, marginBottom: 12 }}>
                    Alur Kerja Terintegrasi
                  </h3>
                  <p style={{ color: 'var(--color-ash)', fontSize: 13, marginBottom: 20 }}>
                    Sistem menegakkan validasi bisnis di setiap langkah pengadaan bahan baku.
                  </p>

                  <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
                    {[
                      { step: '1', title: 'Purchase Request (PR)', desc: 'Kebutuhan diajukan & disetujui pimpinan' },
                      { step: '2', title: 'Purchase Order (PO)', desc: 'Hanya dapat dipesan ke supplier berstatus aktif' },
                      { step: '3', title: 'Goods Receipt (QC)', desc: 'Baris lolos QC otomatis menambah ledger stok' },
                    ].map(st => (
                      <div key={st.step} style={{ display: 'flex', gap: 12, alignItems: 'flex-start' }}>
                        <div
                          style={{
                            width: 24,
                            height: 24,
                            borderRadius: '50%',
                            backgroundColor: 'var(--color-steel)',
                            color: '#fff',
                            display: 'flex',
                            alignItems: 'center',
                            justifyContent: 'center',
                            fontSize: 12,
                            fontWeight: 600,
                            flexShrink: 0,
                          }}
                        >
                          {st.step}
                        </div>
                        <div>
                          <div style={{ fontSize: 14, fontWeight: 500 }}>{st.title}</div>
                          <div style={{ fontSize: 12, color: 'var(--color-ash)' }}>{st.desc}</div>
                        </div>
                      </div>
                    ))}
                  </div>
                </div>

                <div style={{ marginTop: 24 }}>
                  <button
                    onClick={() => setActiveTab('procurement')}
                    className="btn-primary"
                    style={{ width: '100%' }}
                  >
                    Buka Panel Procurement <ChevronRight size={16} />
                  </button>
                </div>
              </div>
            </div>

            {/* Recent Audit Ledger */}
            <div className="card-panel">
              <h3 style={{ fontSize: 18, fontWeight: 600, marginBottom: 16 }}>
                Riwayat Mutasi Buku Besar (Audit-Proof Inventory Ledger)
              </h3>
              {ledger.length === 0 ? (
                <div style={{ color: 'var(--color-ash)', fontSize: 14, padding: '24px 0', textAlign: 'center' }}>
                  Belum ada transaksi stok tercatat.
                </div>
              ) : (
                <table className="data-table">
                  <thead>
                    <tr>
                      <th>WAKTU</th>
                      <th>REFERENSI</th>
                      <th>PRODUK</th>
                      <th>JENIS</th>
                      <th>KUANTITAS</th>
                      <th>CATATAN</th>
                    </tr>
                  </thead>
                  <tbody>
                    {ledger.slice(0, 5).map(tx => (
                      <tr key={tx.id}>
                        <td className="mono-cell" style={{ color: 'var(--color-ash)' }}>
                          {new Date(tx.created_at).toLocaleString('id-ID')}
                        </td>
                        <td className="mono-cell" style={{ color: 'var(--color-signal-blue)' }}>{tx.reference_number}</td>
                        <td>{tx.product_name} ({tx.product_sku})</td>
                        <td>
                          <span className={`status-pill ${tx.transaction_type === 'IN' ? 'completed' : 'cancelled'}`}>
                            {tx.transaction_type}
                          </span>
                        </td>
                        <td className="mono-cell" style={{ fontWeight: 600, color: tx.transaction_type === 'IN' ? '#4ade80' : '#f87171' }}>
                          {tx.transaction_type === 'IN' ? '+' : '-'}{tx.quantity.toLocaleString()} {tx.product_unit}
                        </td>
                        <td style={{ color: 'var(--color-ash)' }}>{tx.notes}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              )}
            </div>
          </div>
        )}

        {/* ============================================================== */}
        {/* TAB 2: MASTER PRODUCTS VIEW                                     */}
        {/* ============================================================== */}
        {activeTab === 'products' && (
          <div>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 28 }}>
              <div>
                <h1 style={{ fontSize: 'var(--text-heading)', fontWeight: 600 }}>Katalog Produk & Bahan Baku</h1>
                <p style={{ color: 'var(--color-ash)', fontSize: 14 }}>Kelola master data produk seperti batok kelapa dan arang briket.</p>
              </div>
              <button onClick={() => setShowProductModal(true)} className="btn-primary">
                <Plus size={16} /> Tambah Produk Baru
              </button>
            </div>

            <div className="card-panel">
              <table className="data-table">
                <thead>
                  <tr>
                    <th>ID</th>
                    <th>SKU / KODE</th>
                    <th>NAMA PRODUK</th>
                    <th>SATUAN</th>
                    <th>STATUS</th>
                    <th>TANGGAL REGISTRASI</th>
                  </tr>
                </thead>
                <tbody>
                  {products.map(p => (
                    <tr key={p.id}>
                      <td className="mono-cell">#{p.id}</td>
                      <td className="mono-cell" style={{ color: 'var(--color-signal-blue)', fontWeight: 500 }}>{p.sku}</td>
                      <td style={{ fontWeight: 500 }}>{p.name}</td>
                      <td>{p.unit}</td>
                      <td>
                        <span className={`status-pill ${p.is_active ? 'active' : 'inactive'}`}>
                          {p.is_active ? 'Aktif' : 'Nonaktif'}
                        </span>
                      </td>
                      <td style={{ color: 'var(--color-ash)' }}>
                        {new Date(p.created_at).toLocaleDateString('id-ID')}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        )}

        {/* ============================================================== */}
        {/* TAB 3: MASTER SUPPLIERS VIEW                                    */}
        {/* ============================================================== */}
        {activeTab === 'suppliers' && (
          <div>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 28 }}>
              <div>
                <h1 style={{ fontSize: 'var(--text-heading)', fontWeight: 600 }}>Direktori Rekanan Supplier</h1>
                <p style={{ color: 'var(--color-ash)', fontSize: 14 }}>Kelola data supplier penyedia bahan baku PT AgroSupply.</p>
              </div>
              <button onClick={() => setShowSupplierModal(true)} className="btn-primary">
                <Plus size={16} /> Daftarkan Supplier
              </button>
            </div>

            <div className="card-panel">
              <table className="data-table">
                <thead>
                  <tr>
                    <th>NAMA PERUSAHAAN</th>
                    <th>KONTAK PIC</th>
                    <th>EMAIL</th>
                    <th>NO. TELEPON</th>
                    <th>ALAMAT</th>
                    <th>STATUS</th>
                  </tr>
                </thead>
                <tbody>
                  {suppliers.map(s => (
                    <tr key={s.id}>
                      <td style={{ fontWeight: 600 }}>{s.name}</td>
                      <td>{s.contact_name || '-'}</td>
                      <td style={{ color: 'var(--color-ash)' }}>{s.email || '-'}</td>
                      <td className="mono-cell">{s.phone || '-'}</td>
                      <td style={{ color: 'var(--color-ash)' }}>{s.address || '-'}</td>
                      <td>
                        <span className={`status-pill ${s.is_active ? 'active' : 'inactive'}`}>
                          {s.is_active ? 'Aktif' : 'Nonaktif'}
                        </span>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        )}

        {/* ============================================================== */}
        {/* TAB 4: PROCUREMENT VIEW (PR & PO)                               */}
        {/* ============================================================== */}
        {activeTab === 'procurement' && (
          <div>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 28 }}>
              <div>
                <h1 style={{ fontSize: 'var(--text-heading)', fontWeight: 600 }}>Pengadaan (Procurement)</h1>
                <p style={{ color: 'var(--color-ash)', fontSize: 14 }}>Alur Purchase Request (PR) dan Purchase Order (PO).</p>
              </div>
              <div style={{ display: 'flex', gap: 12 }}>
                <button onClick={() => setShowPRModal(true)} className="btn-secondary">
                  <Plus size={16} /> Buat PR Baru
                </button>
                <button onClick={() => setShowPOModal(true)} className="btn-primary">
                  <Plus size={16} /> Terbitkan PO Baru
                </button>
              </div>
            </div>

            {/* Purchase Requests Section */}
            <div className="card-panel" style={{ marginBottom: 40 }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
                <h3 style={{ fontSize: 18, fontWeight: 600 }}>1. Dokumen Purchase Request (PR)</h3>
                <span style={{ fontSize: 13, color: 'var(--color-ash)' }}>Total: {prs.length} pengajuan</span>
              </div>

              {prs.length === 0 ? (
                <div style={{ color: 'var(--color-ash)', padding: '20px 0', textAlign: 'center' }}>Belum ada Purchase Request.</div>
              ) : (
                <table className="data-table">
                  <thead>
                    <tr>
                      <th>NO. PR</th>
                      <th>STATUS</th>
                      <th>CATATAN KEBUTUHAN</th>
                      <th>TANGGAL PENGAJUAN</th>
                      <th>AKSI WORKFLOW</th>
                    </tr>
                  </thead>
                  <tbody>
                    {prs.map(pr => (
                      <tr key={pr.id}>
                        <td className="mono-cell" style={{ color: 'var(--color-signal-blue)', fontWeight: 600 }}>{pr.pr_number}</td>
                        <td>
                          <span className={`status-pill ${pr.status.toLowerCase()}`}>{pr.status}</span>
                        </td>
                        <td>{pr.notes || '-'}</td>
                        <td style={{ color: 'var(--color-ash)' }}>{new Date(pr.created_at).toLocaleDateString('id-ID')}</td>
                        <td>
                          <div style={{ display: 'flex', gap: 8 }}>
                            {pr.status === 'DRAFT' && (
                              <button onClick={() => handlePRAction(pr.id, 'submit')} className="btn-secondary" style={{ padding: '6px 12px', fontSize: 12 }}>
                                Submit PR
                              </button>
                            )}
                            {pr.status === 'SUBMITTED' && (
                              <>
                                <button onClick={() => handlePRAction(pr.id, 'approve')} className="btn-primary" style={{ padding: '6px 12px', fontSize: 12 }}>
                                  Approve
                                </button>
                                <button onClick={() => handlePRAction(pr.id, 'reject')} className="btn-secondary" style={{ padding: '6px 12px', fontSize: 12, color: '#f87171' }}>
                                  Reject
                                </button>
                              </>
                            )}
                          </div>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              )}
            </div>

            {/* Purchase Orders Section */}
            <div className="card-panel">
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
                <h3 style={{ fontSize: 18, fontWeight: 600 }}>2. Dokumen Purchase Order (PO)</h3>
                <span style={{ fontSize: 13, color: 'var(--color-ash)' }}>Total: {pos.length} pesanan</span>
              </div>

              {pos.length === 0 ? (
                <div style={{ color: 'var(--color-ash)', padding: '20px 0', textAlign: 'center' }}>Belum ada Purchase Order.</div>
              ) : (
                <table className="data-table">
                  <thead>
                    <tr>
                      <th>NO. PO</th>
                      <th>SUPPLIER REKANAN</th>
                      <th>STATUS</th>
                      <th>TOTAL NILAI (RP)</th>
                      <th>TANGGAL TERBIT</th>
                      <th>AKSI WORKFLOW</th>
                    </tr>
                  </thead>
                  <tbody>
                    {pos.map(po => (
                      <tr key={po.id}>
                        <td className="mono-cell" style={{ color: 'var(--color-signal-blue)', fontWeight: 600 }}>{po.po_number}</td>
                        <td style={{ fontWeight: 500 }}>{po.supplier_name}</td>
                        <td>
                          <span className={`status-pill ${po.status.toLowerCase()}`}>{po.status}</span>
                        </td>
                        <td className="mono-cell" style={{ fontWeight: 600 }}>
                          Rp {po.total_amount.toLocaleString('id-ID')}
                        </td>
                        <td style={{ color: 'var(--color-ash)' }}>{new Date(po.created_at).toLocaleDateString('id-ID')}</td>
                        <td>
                          <div style={{ display: 'flex', gap: 8 }}>
                            {po.status === 'DRAFT' && (
                              <button onClick={() => handlePOAction(po.id, 'approve')} className="btn-primary" style={{ padding: '6px 12px', fontSize: 12 }}>
                                Approve PO
                              </button>
                            )}
                            {(po.status === 'APPROVED' || po.status === 'PARTIALLY_RECEIVED') && (
                              <button
                                onClick={() => {
                                  handleSelectPOForGR(po.id);
                                  setShowGRModal(true);
                                }}
                                className="btn-secondary"
                                style={{ padding: '6px 12px', fontSize: 12, borderColor: 'var(--color-signal-blue)' }}
                              >
                                Catat Penerimaan Truk (GR)
                              </button>
                            )}
                          </div>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              )}
            </div>
          </div>
        )}

        {/* ============================================================== */}
        {/* TAB 5: INVENTORY & STOCK LEDGER VIEW                            */}
        {/* ============================================================== */}
        {activeTab === 'inventory' && (
          <div>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 28 }}>
              <div>
                <h1 style={{ fontSize: 'var(--text-heading)', fontWeight: 600 }}>Inventaris & Buku Besar Stok</h1>
                <p style={{ color: 'var(--color-ash)', fontSize: 14 }}>
                  Pencatatan Goods Receipt (inspeksi QC fisik) dan audit mutasi saldo stok.
                </p>
              </div>
              <button
                onClick={() => {
                  if (pos.some(p => p.status === 'APPROVED' || p.status === 'PARTIALLY_RECEIVED')) {
                    const firstValidPO = pos.find(p => p.status === 'APPROVED' || p.status === 'PARTIALLY_RECEIVED');
                    if (firstValidPO) handleSelectPOForGR(firstValidPO.id);
                    setShowGRModal(true);
                  } else {
                    notify('Tidak ada PO berstatus APPROVED atau PARTIALLY_RECEIVED untuk diterima', 'error');
                  }
                }}
                className="btn-primary"
              >
                <Plus size={16} /> Catat Goods Receipt Baru
              </button>
            </div>

            {/* Saldo Stok Cards */}
            <div className="card-panel" style={{ marginBottom: 40 }}>
              <h3 style={{ fontSize: 18, fontWeight: 600, marginBottom: 16 }}>Saldo Stok Terkini Gudang</h3>
              <table className="data-table">
                <thead>
                  <tr>
                    <th>KODE SKU</th>
                    <th>NAMA BAHAN BAKU</th>
                    <th>AKUMULASI MASUK (IN)</th>
                    <th>AKUMULASI KELUAR (OUT)</th>
                    <th>SALDO STOK SAAT INI</th>
                  </tr>
                </thead>
                <tbody>
                  {stock.map(s => (
                    <tr key={s.product_id}>
                      <td className="mono-cell" style={{ color: 'var(--color-signal-blue)', fontWeight: 600 }}>{s.product_sku}</td>
                      <td style={{ fontWeight: 600 }}>{s.product_name}</td>
                      <td className="mono-cell">{s.total_in.toLocaleString()} {s.product_unit}</td>
                      <td className="mono-cell">{s.total_out.toLocaleString()} {s.product_unit}</td>
                      <td className="mono-cell" style={{ fontWeight: 700, fontSize: 16, color: '#4ade80' }}>
                        {s.current_stock.toLocaleString()} {s.product_unit}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>

            {/* Riwayat Dokumen Goods Receipt */}
            <div className="card-panel" style={{ marginBottom: 40 }}>
              <h3 style={{ fontSize: 18, fontWeight: 600, marginBottom: 16 }}>Dokumen Penerimaan Barang (Goods Receipt)</h3>
              {goodsReceipts.length === 0 ? (
                <div style={{ color: 'var(--color-ash)', padding: '20px 0', textAlign: 'center' }}>Belum ada Goods Receipt tercatat.</div>
              ) : (
                <table className="data-table">
                  <thead>
                    <tr>
                      <th>NO. GR</th>
                      <th>NO. PO INDUK</th>
                      <th>NO. SURAT JALAN</th>
                      <th>TANGGAL TERIMA</th>
                      <th>CATATAN</th>
                    </tr>
                  </thead>
                  <tbody>
                    {goodsReceipts.map(gr => (
                      <tr key={gr.id}>
                        <td className="mono-cell" style={{ color: 'var(--color-signal-blue)', fontWeight: 600 }}>{gr.gr_number}</td>
                        <td className="mono-cell">{gr.po_number || `#${gr.purchase_order_id}`}</td>
                        <td className="mono-cell">{gr.delivery_order_number}</td>
                        <td style={{ color: 'var(--color-ash)' }}>{new Date(gr.receipt_date).toLocaleString('id-ID')}</td>
                        <td style={{ color: 'var(--color-ash)' }}>{gr.notes || '-'}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              )}
            </div>

            {/* Inventory Ledger Timeline */}
            <div className="card-panel">
              <h3 style={{ fontSize: 18, fontWeight: 600, marginBottom: 16 }}>Buku Besar Transaksi Stok (Audit Trail)</h3>
              <table className="data-table">
                <thead>
                  <tr>
                    <th>WAKTU TRANSAKSI</th>
                    <th>DOKUMEN SUMBER</th>
                    <th>PRODUK</th>
                    <th>TIPE MUTASI</th>
                    <th>KUANTITAS</th>
                    <th>KETERANGAN AUDIT</th>
                  </tr>
                </thead>
                <tbody>
                  {ledger.map(tx => (
                    <tr key={tx.id}>
                      <td className="mono-cell" style={{ color: 'var(--color-ash)' }}>
                        {new Date(tx.created_at).toLocaleString('id-ID')}
                      </td>
                      <td className="mono-cell" style={{ color: 'var(--color-signal-blue)' }}>{tx.reference_number}</td>
                      <td>{tx.product_name} ({tx.product_sku})</td>
                      <td>
                        <span className={`status-pill ${tx.transaction_type === 'IN' ? 'completed' : 'cancelled'}`}>
                          {tx.transaction_type}
                        </span>
                      </td>
                      <td className="mono-cell" style={{ fontWeight: 600, color: tx.transaction_type === 'IN' ? '#4ade80' : '#f87171' }}>
                        {tx.transaction_type === 'IN' ? '+' : '-'}{tx.quantity.toLocaleString()} {tx.product_unit}
                      </td>
                      <td style={{ color: 'var(--color-ash)' }}>{tx.notes}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        )}
      </main>

      {/* ============================================================== */}
      {/* MODAL 1: TAMBAH PRODUK                                         */}
      {/* ============================================================== */}
      {showProductModal && (
        <div style={{ position: 'fixed', inset: 0, backgroundColor: 'rgba(0,0,0,0.85)', backdropFilter: 'blur(8px)', zIndex: 100, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
          <div className="card-elevated" style={{ width: 480, maxWidth: '90%' }}>
            <h3 style={{ fontSize: 20, fontWeight: 600, marginBottom: 16 }}>Tambah Produk Baru</h3>
            <form onSubmit={handleCreateProduct}>
              <div style={{ marginBottom: 16 }}>
                <label style={{ display: 'block', fontSize: 13, color: 'var(--color-ash)', marginBottom: 6 }}>KODE SKU *</label>
                <input
                  type="text"
                  required
                  placeholder="Misal: PROD-BATOK-001"
                  value={productForm.sku}
                  onChange={e => setProductForm({ ...productForm, sku: e.target.value })}
                  style={{ width: '100%' }}
                />
              </div>
              <div style={{ marginBottom: 16 }}>
                <label style={{ display: 'block', fontSize: 13, color: 'var(--color-ash)', marginBottom: 6 }}>NAMA PRODUK *</label>
                <input
                  type="text"
                  required
                  placeholder="Misal: Batok Kelapa Kering Grade A"
                  value={productForm.name}
                  onChange={e => setProductForm({ ...productForm, name: e.target.value })}
                  style={{ width: '100%' }}
                />
              </div>
              <div style={{ marginBottom: 24 }}>
                <label style={{ display: 'block', fontSize: 13, color: 'var(--color-ash)', marginBottom: 6 }}>SATUAN / UNIT *</label>
                <select
                  value={productForm.unit}
                  onChange={e => setProductForm({ ...productForm, unit: e.target.value })}
                  style={{ width: '100%' }}
                >
                  <option value="kg">kg (Kilogram)</option>
                  <option value="ton">ton (Ton)</option>
                  <option value="pcs">pcs (Pieces)</option>
                  <option value="karung">karung (Bag 50kg)</option>
                </select>
              </div>
              <div style={{ display: 'flex', justifyContent: 'flex-end', gap: 12 }}>
                <button type="button" onClick={() => setShowProductModal(false)} className="btn-secondary">
                  Batal
                </button>
                <button type="submit" className="btn-primary">
                  Simpan Produk
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* ============================================================== */}
      {/* MODAL 2: DAFTARKAN SUPPLIER                                    */}
      {/* ============================================================== */}
      {showSupplierModal && (
        <div style={{ position: 'fixed', inset: 0, backgroundColor: 'rgba(0,0,0,0.85)', backdropFilter: 'blur(8px)', zIndex: 100, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
          <div className="card-elevated" style={{ width: 520, maxWidth: '90%' }}>
            <h3 style={{ fontSize: 20, fontWeight: 600, marginBottom: 16 }}>Daftarkan Supplier Baru</h3>
            <form onSubmit={handleCreateSupplier}>
              <div style={{ marginBottom: 14 }}>
                <label style={{ display: 'block', fontSize: 13, color: 'var(--color-ash)', marginBottom: 6 }}>NAMA PERUSAHAAN / CV *</label>
                <input
                  type="text"
                  required
                  placeholder="Misal: CV Agro Lestari Mandiri"
                  value={supplierForm.name}
                  onChange={e => setSupplierForm({ ...supplierForm, name: e.target.value })}
                  style={{ width: '100%' }}
                />
              </div>
              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12, marginBottom: 14 }}>
                <div>
                  <label style={{ display: 'block', fontSize: 13, color: 'var(--color-ash)', marginBottom: 6 }}>NAMA KONTAK PIC</label>
                  <input
                    type="text"
                    placeholder="Misal: Pak Budi"
                    value={supplierForm.contact_name}
                    onChange={e => setSupplierForm({ ...supplierForm, contact_name: e.target.value })}
                    style={{ width: '100%' }}
                  />
                </div>
                <div>
                  <label style={{ display: 'block', fontSize: 13, color: 'var(--color-ash)', marginBottom: 6 }}>NO. TELEPON / WA</label>
                  <input
                    type="text"
                    placeholder="08123456789"
                    value={supplierForm.phone}
                    onChange={e => setSupplierForm({ ...supplierForm, phone: e.target.value })}
                    style={{ width: '100%' }}
                  />
                </div>
              </div>
              <div style={{ marginBottom: 14 }}>
                <label style={{ display: 'block', fontSize: 13, color: 'var(--color-ash)', marginBottom: 6 }}>EMAIL</label>
                <input
                  type="email"
                  placeholder="budi@agrolestari.com"
                  value={supplierForm.email}
                  onChange={e => setSupplierForm({ ...supplierForm, email: e.target.value })}
                  style={{ width: '100%' }}
                />
              </div>
              <div style={{ marginBottom: 24 }}>
                <label style={{ display: 'block', fontSize: 13, color: 'var(--color-ash)', marginBottom: 6 }}>ALAMAT GUDANG / KANTOR</label>
                <textarea
                  rows={2}
                  placeholder="Kabupaten Lampung Selatan"
                  value={supplierForm.address}
                  onChange={e => setSupplierForm({ ...supplierForm, address: e.target.value })}
                  style={{ width: '100%' }}
                />
              </div>
              <div style={{ display: 'flex', justifyContent: 'flex-end', gap: 12 }}>
                <button type="button" onClick={() => setShowSupplierModal(false)} className="btn-secondary">
                  Batal
                </button>
                <button type="submit" className="btn-primary">
                  Daftarkan Supplier
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* ============================================================== */}
      {/* MODAL 3: BUAT PURCHASE REQUEST (PR)                             */}
      {/* ============================================================== */}
      {showPRModal && (
        <div style={{ position: 'fixed', inset: 0, backgroundColor: 'rgba(0,0,0,0.85)', backdropFilter: 'blur(8px)', zIndex: 100, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
          <div className="card-elevated" style={{ width: 560, maxWidth: '90%' }}>
            <h3 style={{ fontSize: 20, fontWeight: 600, marginBottom: 16 }}>Buat Purchase Request (PR) Baru</h3>
            <form onSubmit={handleCreatePR}>
              <div style={{ marginBottom: 16 }}>
                <label style={{ display: 'block', fontSize: 13, color: 'var(--color-ash)', marginBottom: 6 }}>CATATAN KEBUTUHAN PENGADAAN</label>
                <input
                  type="text"
                  placeholder="Misal: Stok produksi batok kelapa minggu ke-4"
                  value={prForm.notes}
                  onChange={e => setPRForm({ ...prForm, notes: e.target.value })}
                  style={{ width: '100%' }}
                />
              </div>

              <div style={{ marginBottom: 20 }}>
                <label style={{ display: 'block', fontSize: 13, color: 'var(--color-ash)', marginBottom: 8 }}>BARANG YANG DIBUTUHKAN</label>
                <div style={{ display: 'grid', gridTemplateColumns: '2fr 1fr', gap: 12 }}>
                  <select
                    required
                    value={prForm.items[0]?.product_id}
                    onChange={e => {
                      const id = Number(e.target.value);
                      const items = [...prForm.items];
                      items[0].product_id = id;
                      setPRForm({ ...prForm, items });
                    }}
                  >
                    <option value={0}>-- Pilih Produk --</option>
                    {products.map(p => (
                      <option key={p.id} value={p.id}>{p.name} ({p.sku})</option>
                    ))}
                  </select>

                  <input
                    type="number"
                    required
                    min={1}
                    placeholder="Jumlah"
                    value={prForm.items[0]?.quantity}
                    onChange={e => {
                      const items = [...prForm.items];
                      items[0].quantity = Number(e.target.value);
                      setPRForm({ ...prForm, items });
                    }}
                  />
                </div>
              </div>

              <div style={{ display: 'flex', justifyContent: 'flex-end', gap: 12 }}>
                <button type="button" onClick={() => setShowPRModal(false)} className="btn-secondary">
                  Batal
                </button>
                <button type="submit" className="btn-primary">
                  Buat Dokumen PR
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* ============================================================== */}
      {/* MODAL 4: BUAT PURCHASE ORDER (PO)                              */}
      {/* ============================================================== */}
      {showPOModal && (
        <div style={{ position: 'fixed', inset: 0, backgroundColor: 'rgba(0,0,0,0.85)', backdropFilter: 'blur(8px)', zIndex: 100, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
          <div className="card-elevated" style={{ width: 600, maxWidth: '90%' }}>
            <h3 style={{ fontSize: 20, fontWeight: 600, marginBottom: 16 }}>Terbitkan Purchase Order (PO) Baru</h3>
            <form onSubmit={handleCreatePO}>
              <div style={{ marginBottom: 14 }}>
                <label style={{ display: 'block', fontSize: 13, color: 'var(--color-ash)', marginBottom: 6 }}>PILIH SUPPLIER AKTIF *</label>
                <select
                  required
                  value={poForm.supplier_id}
                  onChange={e => setPOForm({ ...poForm, supplier_id: Number(e.target.value) })}
                  style={{ width: '100%' }}
                >
                  <option value={0}>-- Pilih Supplier Aktif --</option>
                  {suppliers.filter(s => s.is_active).map(s => (
                    <option key={s.id} value={s.id}>{s.name} ({s.contact_name})</option>
                  ))}
                </select>
              </div>

              <div style={{ marginBottom: 14 }}>
                <label style={{ display: 'block', fontSize: 13, color: 'var(--color-ash)', marginBottom: 6 }}>RUJUKAN PR TERSETUJUI (OPSIONAL)</label>
                <select
                  value={poForm.purchase_request_id}
                  onChange={e => {
                    const prId = Number(e.target.value);
                    const selectedPR = prs.find(p => p.id === prId);
                    setPOForm({
                      ...poForm,
                      purchase_request_id: prId,
                      items: selectedPR?.items?.map(it => ({
                        product_id: it.product_id,
                        quantity: it.quantity,
                        unit_price: 2500,
                      })) || poForm.items,
                    });
                  }}
                  style={{ width: '100%' }}
                >
                  <option value={0}>-- PO Mandiri Tanpa PR --</option>
                  {prs.filter(p => p.status === 'APPROVED').map(p => (
                    <option key={p.id} value={p.id}>{p.pr_number} - {p.notes}</option>
                  ))}
                </select>
              </div>

              <div style={{ marginBottom: 14 }}>
                <label style={{ display: 'block', fontSize: 13, color: 'var(--color-ash)', marginBottom: 8 }}>BARIS PESANAN & HARGA BELI</label>
                <div style={{ display: 'grid', gridTemplateColumns: '2fr 1fr 1fr', gap: 10 }}>
                  <select
                    required
                    value={poForm.items[0]?.product_id}
                    onChange={e => {
                      const items = [...poForm.items];
                      items[0].product_id = Number(e.target.value);
                      setPOForm({ ...poForm, items });
                    }}
                  >
                    <option value={0}>-- Produk --</option>
                    {products.map(p => (
                      <option key={p.id} value={p.id}>{p.name}</option>
                    ))}
                  </select>

                  <input
                    type="number"
                    required
                    min={1}
                    placeholder="Qty"
                    value={poForm.items[0]?.quantity}
                    onChange={e => {
                      const items = [...poForm.items];
                      items[0].quantity = Number(e.target.value);
                      setPOForm({ ...poForm, items });
                    }}
                  />

                  <input
                    type="number"
                    required
                    min={0}
                    placeholder="Harga/kg"
                    value={poForm.items[0]?.unit_price}
                    onChange={e => {
                      const items = [...poForm.items];
                      items[0].unit_price = Number(e.target.value);
                      setPOForm({ ...poForm, items });
                    }}
                  />
                </div>
              </div>

              <div style={{ marginBottom: 24 }}>
                <label style={{ display: 'block', fontSize: 13, color: 'var(--color-ash)', marginBottom: 6 }}>CATATAN PENGIRIMAN</label>
                <input
                  type="text"
                  placeholder="Misal: Pengiriman Franco Pabrik, bayar tempo 14 hari"
                  value={poForm.notes}
                  onChange={e => setPOForm({ ...poForm, notes: e.target.value })}
                  style={{ width: '100%' }}
                />
              </div>

              <div style={{ display: 'flex', justifyContent: 'flex-end', gap: 12 }}>
                <button type="button" onClick={() => setShowPOModal(false)} className="btn-secondary">
                  Batal
                </button>
                <button type="submit" className="btn-primary">
                  Terbitkan Dokumen PO
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* ============================================================== */}
      {/* MODAL 5: CATAT GOODS RECEIPT DENGAN QC                          */}
      {/* ============================================================== */}
      {showGRModal && (
        <div style={{ position: 'fixed', inset: 0, backgroundColor: 'rgba(0,0,0,0.85)', backdropFilter: 'blur(8px)', zIndex: 100, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
          <div className="card-elevated" style={{ width: 680, maxWidth: '90%' }}>
            <h3 style={{ fontSize: 20, fontWeight: 600, marginBottom: 8 }}>Penerimaan Barang Fisik (Goods Receipt)</h3>
            <p style={{ color: 'var(--color-ash)', fontSize: 13, marginBottom: 20 }}>
              Pemeriksaan QC fisik truk. Kuantitas yang lolos (accepted) akan otomatis masuk ke Buku Besar Stok.
            </p>

            <form onSubmit={handleCreateGR}>
              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12, marginBottom: 16 }}>
                <div>
                  <label style={{ display: 'block', fontSize: 13, color: 'var(--color-ash)', marginBottom: 6 }}>PILIH PURCHASE ORDER (PO)</label>
                  <select
                    required
                    value={grForm.purchase_order_id}
                    onChange={e => handleSelectPOForGR(Number(e.target.value))}
                    style={{ width: '100%' }}
                  >
                    <option value={0}>-- Pilih PO Terbuka --</option>
                    {pos.filter(p => p.status === 'APPROVED' || p.status === 'PARTIALLY_RECEIVED').map(p => (
                      <option key={p.id} value={p.id}>{p.po_number} ({p.supplier_name})</option>
                    ))}
                  </select>
                </div>

                <div>
                  <label style={{ display: 'block', fontSize: 13, color: 'var(--color-ash)', marginBottom: 6 }}>NO. SURAT JALAN SUPPLIER *</label>
                  <input
                    type="text"
                    required
                    placeholder="Misal: SJ/2026/089"
                    value={grForm.delivery_order_number}
                    onChange={e => setGRForm({ ...grForm, delivery_order_number: e.target.value })}
                    style={{ width: '100%' }}
                  />
                </div>
              </div>

              {/* Items QC Inspection Table */}
              <div style={{ marginBottom: 20 }}>
                <label style={{ display: 'block', fontSize: 13, color: 'var(--color-ash)', marginBottom: 8 }}>HASIL INSPEKSI MUTU & TIMBANGAN (QC)</label>
                {grForm.items.map((item, idx) => (
                  <div key={idx} style={{ backgroundColor: 'var(--color-carbon)', padding: 14, borderRadius: 8, marginBottom: 12 }}>
                    <div style={{ fontWeight: 600, fontSize: 14, marginBottom: 10, color: 'var(--color-paper)' }}>
                      {item.product_name || `Produk ID #${item.product_id}`}
                    </div>
                    <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr 1fr', gap: 10, marginBottom: 10 }}>
                      <div>
                        <span style={{ fontSize: 11, color: 'var(--color-ash)' }}>TOTAL TIBA (KG)</span>
                        <input
                          type="number"
                          required
                          min={1}
                          value={item.received_quantity}
                          onChange={e => {
                            const val = Number(e.target.value);
                            const items = [...grForm.items];
                            items[idx].received_quantity = val;
                            items[idx].accepted_quantity = val - items[idx].rejected_quantity;
                            setGRForm({ ...grForm, items });
                          }}
                          style={{ width: '100%', marginTop: 4 }}
                        />
                      </div>
                      <div>
                        <span style={{ fontSize: 11, color: '#4ade80' }}>LOLOS QC (MASUK STOK)</span>
                        <input
                          type="number"
                          required
                          min={0}
                          value={item.accepted_quantity}
                          onChange={e => {
                            const val = Number(e.target.value);
                            const items = [...grForm.items];
                            items[idx].accepted_quantity = val;
                            items[idx].rejected_quantity = items[idx].received_quantity - val;
                            setGRForm({ ...grForm, items });
                          }}
                          style={{ width: '100%', marginTop: 4 }}
                        />
                      </div>
                      <div>
                        <span style={{ fontSize: 11, color: '#f87171' }}>DITOLAK (CACAT/KOTOR)</span>
                        <input
                          type="number"
                          min={0}
                          value={item.rejected_quantity}
                          onChange={e => {
                            const val = Number(e.target.value);
                            const items = [...grForm.items];
                            items[idx].rejected_quantity = val;
                            items[idx].accepted_quantity = items[idx].received_quantity - val;
                            setGRForm({ ...grForm, items });
                          }}
                          style={{ width: '100%', marginTop: 4 }}
                        />
                      </div>
                    </div>
                    {item.rejected_quantity > 0 && (
                      <input
                        type="text"
                        required
                        placeholder="Alasan penolakan (misal: basah, bercampur pasir, rapuh)"
                        value={item.rejection_reason}
                        onChange={e => {
                          const items = [...grForm.items];
                          items[idx].rejection_reason = e.target.value;
                          setGRForm({ ...grForm, items });
                        }}
                        style={{ width: '100%', fontSize: 13 }}
                      />
                    )}
                  </div>
                ))}
              </div>

              <div style={{ display: 'flex', justifyContent: 'flex-end', gap: 12 }}>
                <button type="button" onClick={() => setShowGRModal(false)} className="btn-secondary">
                  Batal
                </button>
                <button type="submit" className="btn-primary">
                  Posting ke Inventory Ledger
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
