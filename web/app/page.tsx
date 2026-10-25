'use client';

import React, { useState, useEffect } from 'react';
import {
  api,
  authStorage,
  User,
  UserRole,
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
  Shield,
  ShieldCheck,
  LogOut,
  KeyRound,
  UserCheck,
  Lock,
  ShieldAlert,
} from 'lucide-react';

export default function MissionControlERP() {
  const [activeTab, setActiveTab] = useState<'dashboard' | 'products' | 'suppliers' | 'procurement' | 'inventory' | 'users'>('dashboard');
  const [backendReady, setBackendReady] = useState<boolean | null>(null);
  const [loading, setLoading] = useState(false);
  const [notification, setNotification] = useState<{ type: 'success' | 'error'; message: string } | null>(null);

  // Auth & RBAC State (Phase 7)
  const [currentUser, setCurrentUser] = useState<User | null>(null);
  const [authInitialized, setAuthInitialized] = useState(false);
  const [loginLoading, setLoginLoading] = useState(false);
  const [loginError, setLoginError] = useState<string | null>(null);
  const [credentials, setCredentials] = useState({ username: 'admin', password: 'password123' });

  // User Management State (Admin only)
  const [userList, setUserList] = useState<User[]>([]);
  const [showUserModal, setShowUserModal] = useState(false);
  const [userForm, setUserForm] = useState({
    username: '',
    email: '',
    password: '',
    full_name: '',
    role: 'PURCHASING' as UserRole,
  });

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
  const [loadingPOItems, setLoadingPOItems] = useState(false);

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

  const getRoleColor = (role?: UserRole) => {
    switch (role) {
      case 'ADMIN': return '#7c3aed';
      case 'PURCHASING': return '#3398e1';
      case 'MANAGER': return '#b45309';
      case 'WAREHOUSE': return '#15803d';
      default: return '#78716c';
    }
  };

  const getRoleBg = (role?: UserRole) => {
    switch (role) {
      case 'ADMIN': return '#f5f3ff';
      case 'PURCHASING': return '#eff9ff';
      case 'MANAGER': return '#fffbeb';
      case 'WAREHOUSE': return '#f0fdf4';
      default: return '#fafaf9';
    }
  };

  const getRoleBorder = (role?: UserRole) => {
    switch (role) {
      case 'ADMIN': return '#ddd6fe';
      case 'PURCHASING': return '#c1e1f7';
      case 'MANAGER': return '#fde68a';
      case 'WAREHOUSE': return '#bbf7d0';
      default: return '#e8e6e5';
    }
  };

  const loadUsers = async () => {
    try {
      const list = await api.getUsers();
      setUserList(list);
    } catch {
      // ignore if unauthorized
    }
  };

  const loadAllData = async () => {
    setLoading(true);
    try {
      const isReady = await api.checkHealth();
      setBackendReady(isReady);

      if (isReady && authStorage.getToken()) {
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
    const savedUser = authStorage.getUser();
    const token = authStorage.getToken();
    if (savedUser && token) {
      setCurrentUser(savedUser);
      loadAllData();
      if (savedUser.role === 'ADMIN') {
        loadUsers();
      }
    } else {
      api.checkHealth().then(setBackendReady);
    }
    setAuthInitialized(true);

    const onUnauthorized = () => {
      setCurrentUser(null);
      notify('Sesi autentikasi telah berakhir. Silakan login kembali.', 'error');
    };
    window.addEventListener('auth:unauthorized', onUnauthorized);
    return () => window.removeEventListener('auth:unauthorized', onUnauthorized);
  }, []);

  const handleQuickDemoLogin = async (username: string) => {
    setLoginLoading(true);
    setLoginError(null);
    try {
      const res = await api.login({ username, password: 'password123' });
      authStorage.setToken(res.token);
      authStorage.setUser(res.user);
      setCurrentUser(res.user);
      notify(`Berhasil masuk sebagai ${res.user.full_name} (${res.user.role})`);
      setLoading(true);
      await loadAllData();
      if (res.user.role === 'ADMIN') {
        await loadUsers();
      }
    } catch (err: any) {
      setLoginError(err.message || 'Gagal login demo');
    } finally {
      setLoginLoading(false);
    }
  };

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoginLoading(true);
    setLoginError(null);
    try {
      const res = await api.login(credentials);
      authStorage.setToken(res.token);
      authStorage.setUser(res.user);
      setCurrentUser(res.user);
      notify(`Selamat datang, ${res.user.full_name}!`);
      await loadAllData();
      if (res.user.role === 'ADMIN') {
        await loadUsers();
      }
    } catch (err: any) {
      setLoginError(err.message || 'Login gagal, periksa username dan password.');
    } finally {
      setLoginLoading(false);
    }
  };

  const handleLogout = () => {
    authStorage.clear();
    setCurrentUser(null);
    notify('Anda telah keluar dari sesi.');
  };

  const handleCreateUser = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await api.registerUser(userForm);
      notify(`Pengguna ${userForm.username} (${userForm.role}) berhasil didaftarkan`);
      setShowUserModal(false);
      setUserForm({ username: '', email: '', password: '', full_name: '', role: 'PURCHASING' });
      loadUsers();
    } catch (err: any) {
      notify(err.message, 'error');
    }
  };

  const handleToggleUserStatus = async (user: User) => {
    try {
      await api.updateUserStatus(user.id, !user.is_active);
      notify(`Status pengguna ${user.username} diubah menjadi ${!user.is_active ? 'Aktif' : 'Nonaktif'}`);
      loadUsers();
    } catch (err: any) {
      notify(err.message, 'error');
    }
  };

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

  const handleSelectPOForGR = async (poId: number) => {
    if (!poId || poId <= 0) {
      setGRForm(prev => ({
        ...prev,
        purchase_order_id: 0,
        items: [],
      }));
      return;
    }

    setLoadingPOItems(true);
    try {
      const poDetail = await api.getPO(poId);
      if (!poDetail || !poDetail.items || poDetail.items.length === 0) {
        notify('PO ini tidak memiliki daftar item barang yang valid', 'error');
        setGRForm(prev => ({
          ...prev,
          purchase_order_id: poId,
          items: [],
        }));
        return;
      }

      setGRForm(prev => ({
        ...prev,
        purchase_order_id: poId,
        items: poDetail.items!.map(item => ({
          purchase_order_item_id: item.id,
          product_id: item.product_id,
          product_name: item.product_name,
          received_quantity: item.quantity,
          accepted_quantity: item.quantity,
          rejected_quantity: 0,
          rejection_reason: '',
        })),
      }));
    } catch (err: any) {
      notify(`Gagal memuat item PO: ${err.message}`, 'error');
    } finally {
      setLoadingPOItems(false);
    }
  };

  const handleCreateGR = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!grForm.purchase_order_id || grForm.purchase_order_id <= 0) {
      notify('Silakan pilih Purchase Order yang valid terlebih dahulu', 'error');
      return;
    }
    if (!grForm.items || grForm.items.length === 0) {
      notify('Daftar item barang dari PO belum dimuat atau kosong. Silakan pilih ulang PO.', 'error');
      return;
    }
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
    <div style={{ minHeight: '100vh', display: 'flex', flexDirection: 'column', backgroundColor: 'var(--color-stone-canvas)' }}>
      {/* Top Notification Toast */}
      {notification && (
        <div
          style={{
            position: 'fixed',
            top: 24,
            right: 24,
            zIndex: 1000,
            padding: '12px 20px',
            borderRadius: 'var(--radius-buttons)',
            backgroundColor: notification.type === 'success' ? '#f0fdf4' : '#fef2f2',
            color: notification.type === 'success' ? '#15803d' : '#b91c1c',
            boxShadow: 'var(--shadow-xl)',
            display: 'flex',
            alignItems: 'center',
            gap: 10,
            fontSize: 14,
            border: notification.type === 'success' ? '1px solid #bbf7d0' : '1px solid #fecaca',
          }}
        >
          {notification.type === 'success' ? <CheckCircle2 size={18} /> : <AlertCircle size={18} />}
          {notification.message}
        </div>
      )}

      {/* Top Bar Header Navigation (Seline Analytics â€” light editorial aesthetic) */}
      <header
        style={{
          height: 64,
          borderBottom: '1px solid var(--color-stone-border)',
          backgroundColor: 'var(--color-pure-white)',
          boxShadow: 'var(--shadow-subtle)',
          position: 'sticky',
          top: 0,
          zIndex: 50,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          padding: '0 32px',
        }}
      >
        {/* Logo Mark: AgroSupply wordmark */}
        <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
          <div
            style={{
              width: 28,
              height: 28,
              borderRadius: 6,
              backgroundColor: 'var(--color-cyan-signal)',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              fontFamily: 'var(--font-display)',
              fontWeight: 600,
              fontSize: 14,
              color: '#fff',
              flexShrink: 0,
            }}
          >
            A
          </div>
          <div style={{ fontFamily: 'var(--font-display)', fontWeight: 500, fontSize: 14, color: 'var(--color-ink-black)', letterSpacing: '-0.01em' }}>
            Agro-Supply <span style={{ color: 'var(--color-warm-gray)', fontWeight: 400 }}>ERP</span>
          </div>
        </div>

        {/* Center Nav Links */}
        {currentUser && (
          <nav style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
            {[
              { id: 'dashboard', label: 'Dashboard' },
              { id: 'products', label: 'Master Produk' },
              { id: 'suppliers', label: 'Master Supplier' },
              { id: 'procurement', label: 'Procurement' },
              { id: 'inventory', label: 'Inventory' },
              ...(currentUser.role === 'ADMIN' ? [{ id: 'users', label: 'Pengguna' }] : []),
            ].map(tab => (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id as any)}
                className={`nav-link ${activeTab === tab.id ? 'active' : ''}`}
                style={{
                  backgroundColor: activeTab === tab.id ? 'var(--color-stone-canvas)' : 'transparent',
                }}
              >
                {tab.label}
              </button>
            ))}
          </nav>
        )}

        {/* Right: User Profile & Status */}
        <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
          {currentUser && (
            <div
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: 8,
                backgroundColor: getRoleBg(currentUser.role),
                border: `1px solid ${getRoleBorder(currentUser.role)}`,
                padding: '5px 12px',
                borderRadius: 'var(--radius-full)',
              }}
            >
              <div
                style={{
                  width: 22,
                  height: 22,
                  borderRadius: '50%',
                  backgroundColor: getRoleColor(currentUser.role),
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  color: '#fff',
                  fontSize: 10,
                  fontWeight: 600,
                  flexShrink: 0,
                }}
              >
                {currentUser.full_name ? currentUser.full_name.charAt(0).toUpperCase() : 'U'}
              </div>
              <div style={{ textAlign: 'left' }}>
                <div style={{ fontSize: 13, fontWeight: 500, color: 'var(--color-ink-black)', lineHeight: 1.2 }}>
                  {currentUser.full_name}
                </div>
                <div
                  style={{
                    fontSize: 10,
                    fontWeight: 500,
                    color: getRoleColor(currentUser.role),
                    display: 'flex',
                    alignItems: 'center',
                    gap: 3,
                  }}
                >
                  <Shield size={9} /> {currentUser.role}
                </div>
              </div>
              <button
                onClick={handleLogout}
                title="Keluar (Logout)"
                style={{
                  marginLeft: 4,
                  background: 'transparent',
                  border: 'none',
                  color: 'var(--color-ash-gray)',
                  cursor: 'pointer',
                  padding: 3,
                  display: 'flex',
                  alignItems: 'center',
                }}
              >
                <LogOut size={13} />
              </button>
            </div>
          )}

          <div
            className={`status-pill ${backendReady ? 'approved' : 'cancelled'}`}
            style={{ gap: 6 }}
          >
            <span
              style={{
                width: 6,
                height: 6,
                borderRadius: '50%',
                backgroundColor: backendReady ? '#15803d' : '#b91c1c',
                flexShrink: 0,
              }}
            />
            {backendReady ? 'API :8080' : 'Disconnected'}
          </div>

          {currentUser && (
            <button
              onClick={loadAllData}
              className="btn-secondary"
              title="Refresh Data"
              style={{ padding: '7px 12px' }}
            >
              <RefreshCw size={14} className={loading ? 'animate-spin' : ''} />
            </button>
          )}
        </div>
      </header>

      {/* Main Body */}
      <main style={{ flex: 1, padding: '40px 32px', maxWidth: 'var(--page-max-width)', margin: '0 auto', width: '100%' }}>
        {!currentUser ? (
          <div style={{ maxWidth: 820, margin: '32px auto', padding: '0 16px' }}>
            <div className="card-panel" style={{ padding: '48px 40px' }}>
              <div style={{ textAlign: 'center', marginBottom: 40 }}>
                <div
                  style={{
                    width: 48,
                    height: 48,
                    borderRadius: 12,
                    backgroundColor: '#eff9ff',
                    border: '1px solid var(--color-sky-wash)',
                    color: 'var(--color-cyan-signal)',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    margin: '0 auto 20px',
                  }}
                >
                  <ShieldCheck size={26} />
                </div>
                <h1 style={{ fontFamily: 'var(--font-display)', fontSize: 28, fontWeight: 400, letterSpacing: '-0.6px', marginBottom: 10, color: 'var(--color-ink-black)' }}>
                  Masuk ke <span className="highlight-span">Agro-Supply</span> ERP
                </h1>
                <p style={{ color: 'var(--color-warm-gray)', fontSize: 14, lineHeight: 1.69, maxWidth: 500, margin: '0 auto' }}>
                  Sistem pengadaan terproteksi dengan Role-Based Access Control (RBAC).
                  Pilih peran demonstrasi atau masukkan kredensial Anda.
                </p>
              </div>

              {loginError && (
                <div style={{ backgroundColor: '#fef2f2', border: '1px solid #fecaca', color: '#b91c1c', padding: '12px 16px', borderRadius: 'var(--radius-cards)', marginBottom: 24, fontSize: 13, display: 'flex', alignItems: 'center', gap: 10 }}>
                  <AlertCircle size={16} />
                  <span>{loginError}</span>
                </div>
              )}

              {/* Quick Demo Role Switcher */}
              <div style={{ marginBottom: 32 }}>
                <div style={{ fontSize: 11, fontWeight: 600, color: 'var(--color-ash-gray)', textTransform: 'uppercase', letterSpacing: '0.06em', marginBottom: 12 }}>
                  Masuk Cepat â€” Pilih Peran Demo
                </div>
                <div style={{ display: 'grid', gridTemplateColumns: 'repeat(2, 1fr)', gap: 12 }}>
                  <button
                    type="button"
                    onClick={() => handleQuickDemoLogin('admin')}
                    disabled={loginLoading}
                    style={{
                      background: '#f5f3ff',
                      border: '1px solid #ddd6fe',
                      padding: '16px',
                      borderRadius: 'var(--radius-cards)',
                      textAlign: 'left',
                      cursor: 'pointer',
                      transition: 'border-color 0.15s ease, box-shadow 0.15s ease',
                    }}
                  >
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 6 }}>
                      <span style={{ fontWeight: 500, color: '#7c3aed', fontSize: 14 }}>System Admin</span>
                      <span className="status-pill" style={{ backgroundColor: '#f5f3ff', borderColor: '#ddd6fe', color: '#7c3aed', fontSize: 10 }}>ADMIN</span>
                    </div>
                    <div style={{ fontSize: 12, color: 'var(--color-warm-gray)', lineHeight: 1.5 }}>
                      Akses penuh konfigurasi sistem, data master, alur pengadaan, dan manajemen akun pengguna.
                    </div>
                  </button>

                  <button
                    type="button"
                    onClick={() => handleQuickDemoLogin('purchasing')}
                    disabled={loginLoading}
                    style={{
                      background: '#eff9ff',
                      border: '1px solid var(--color-sky-wash)',
                      padding: '16px',
                      borderRadius: 'var(--radius-cards)',
                      textAlign: 'left',
                      cursor: 'pointer',
                      transition: 'border-color 0.15s ease, box-shadow 0.15s ease',
                    }}
                  >
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 6 }}>
                      <span style={{ fontWeight: 500, color: 'var(--color-cyan-edge)', fontSize: 14 }}>Purchasing Staff</span>
                      <span className="status-pill submitted" style={{ fontSize: 10 }}>PURCHASING</span>
                    </div>
                    <div style={{ fontSize: 12, color: 'var(--color-warm-gray)', lineHeight: 1.5 }}>
                      Membuat Purchase Request (PR), menerbitkan Purchase Order (PO), dan mendaftarkan rekanan supplier.
                    </div>
                  </button>

                  <button
                    type="button"
                    onClick={() => handleQuickDemoLogin('manager')}
                    disabled={loginLoading}
                    style={{
                      background: '#fffbeb',
                      border: '1px solid #fde68a',
                      padding: '16px',
                      borderRadius: 'var(--radius-cards)',
                      textAlign: 'left',
                      cursor: 'pointer',
                      transition: 'border-color 0.15s ease, box-shadow 0.15s ease',
                    }}
                  >
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 6 }}>
                      <span style={{ fontWeight: 500, color: '#b45309', fontSize: 14 }}>Procurement Manager</span>
                      <span className="status-pill" style={{ backgroundColor: '#fffbeb', borderColor: '#fde68a', color: '#b45309', fontSize: 10 }}>MANAGER</span>
                    </div>
                    <div style={{ fontSize: 12, color: 'var(--color-warm-gray)', lineHeight: 1.5 }}>
                      Otorisasi dokumen pengadaan: Menyetujui atau menolak PR & PO dari tim Purchasing.
                    </div>
                  </button>

                  <button
                    type="button"
                    onClick={() => handleQuickDemoLogin('warehouse')}
                    disabled={loginLoading}
                    style={{
                      background: '#f0fdf4',
                      border: '1px solid #bbf7d0',
                      padding: '16px',
                      borderRadius: 'var(--radius-cards)',
                      textAlign: 'left',
                      cursor: 'pointer',
                      transition: 'border-color 0.15s ease, box-shadow 0.15s ease',
                    }}
                  >
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 6 }}>
                      <span style={{ fontWeight: 500, color: '#15803d', fontSize: 14 }}>Warehouse & QC Officer</span>
                      <span className="status-pill approved" style={{ fontSize: 10 }}>WAREHOUSE</span>
                    </div>
                    <div style={{ fontSize: 12, color: 'var(--color-warm-gray)', lineHeight: 1.5 }}>
                      Mencatat fisik truk datang, melakukan inspeksi mutu (QC), dan membukukan mutasi buku besar stok.
                    </div>
                  </button>
                </div>
              </div>

              {/* Divider */}
              <div style={{ display: 'flex', alignItems: 'center', gap: 16, marginBottom: 28 }}>
                <div style={{ flex: 1, height: 1, backgroundColor: 'var(--color-stone-border)' }} />
                <span style={{ fontSize: 11, color: 'var(--color-ash-gray)', textTransform: 'uppercase', letterSpacing: '0.06em', whiteSpace: 'nowrap' }}>
                  atau masuk dengan kredensial
                </span>
                <div style={{ flex: 1, height: 1, backgroundColor: 'var(--color-stone-border)' }} />
              </div>

              {/* Manual Form */}
              <form onSubmit={handleLogin}>
                <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 16, marginBottom: 20 }}>
                  <div>
                    <label style={{ display: 'block', fontSize: 12, fontWeight: 500, color: 'var(--color-warm-gray)', marginBottom: 6, letterSpacing: '0.03em', textTransform: 'uppercase' }}>
                      Username
                    </label>
                    <input
                      type="text"
                      required
                      placeholder="Username Anda"
                      value={credentials.username}
                      onChange={e => setCredentials({ ...credentials, username: e.target.value })}
                      style={{ width: '100%' }}
                    />
                  </div>
                  <div>
                    <label style={{ display: 'block', fontSize: 12, fontWeight: 500, color: 'var(--color-warm-gray)', marginBottom: 6, letterSpacing: '0.03em', textTransform: 'uppercase' }}>
                      Password
                    </label>
                    <input
                      type="password"
                      required
                      placeholder="â€¢â€¢â€¢â€¢â€¢â€¢â€¢â€¢"
                      value={credentials.password}
                      onChange={e => setCredentials({ ...credentials, password: e.target.value })}
                      style={{ width: '100%' }}
                    />
                  </div>
                </div>

                <button
                  type="submit"
                  disabled={loginLoading}
                  className="btn-primary"
                  style={{ width: '100%', padding: '10px', justifyContent: 'center' }}
                >
                  {loginLoading ? 'Memverifikasi...' : 'Masuk'}
                </button>
              </form>
            </div>
          </div>
        ) : (
          <>
            {/* ============================================================== */}
            {/* TAB 1: DASHBOARD VIEW                                           */}
            {/* ============================================================== */}
            {activeTab === 'dashboard' && (
          <div>
            <div style={{ marginBottom: 32 }}>
              <h1 style={{ fontFamily: 'var(--font-display)', fontSize: 28, fontWeight: 400, letterSpacing: 'var(--tracking-heading-sm)', color: 'var(--color-ink-black)', marginBottom: 6 }}>
                Dashboard <span className="highlight-span">Pengadaan</span>
              </h1>
              <p style={{ color: 'var(--color-warm-gray)', fontSize: 14, lineHeight: 1.69 }}>
                Pemantauan rantai pasok bahan baku batok kelapa dari pengadaan hingga saldo stok gudang.
              </p>
            </div>

            {/* 4 Metric Cards */}
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: 16, marginBottom: 32 }}>
              <div className="card-panel">
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: 12 }}>
                  <span style={{ fontSize: 12, fontWeight: 500, color: 'var(--color-warm-gray)', textTransform: 'uppercase', letterSpacing: '0.04em' }}>Total Stok Fisik</span>
                  <Boxes size={16} color="var(--color-cyan-signal)" />
                </div>
                <div style={{ fontSize: 28, fontWeight: 400, letterSpacing: '-0.5px', color: 'var(--color-ink-black)', fontFamily: 'var(--font-display)' }}>
                  {totalStockKg.toLocaleString()} <span style={{ fontSize: 14, color: 'var(--color-warm-gray)' }}>kg</span>
                </div>
                <div style={{ fontSize: 12, color: 'var(--color-ash-gray)', marginTop: 6 }}>
                  Akumulasi terverifikasi QC
                </div>
              </div>

              <div className="card-panel">
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: 12 }}>
                  <span style={{ fontSize: 12, fontWeight: 500, color: 'var(--color-warm-gray)', textTransform: 'uppercase', letterSpacing: '0.04em' }}>PO Dalam Proses</span>
                  <Truck size={16} color="var(--color-cyan-signal)" />
                </div>
                <div style={{ fontSize: 28, fontWeight: 400, letterSpacing: '-0.5px', color: 'var(--color-ink-black)', fontFamily: 'var(--font-display)' }}>
                  {approvedPOCount}
                </div>
                <div style={{ fontSize: 12, color: 'var(--color-ash-gray)', marginTop: 6 }}>
                  Menunggu atau sebagian dikirim
                </div>
              </div>

              <div className="card-panel">
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: 12 }}>
                  <span style={{ fontSize: 12, fontWeight: 500, color: 'var(--color-warm-gray)', textTransform: 'uppercase', letterSpacing: '0.04em' }}>Supplier Aktif</span>
                  <Users size={16} color="var(--color-cyan-signal)" />
                </div>
                <div style={{ fontSize: 28, fontWeight: 400, letterSpacing: '-0.5px', color: 'var(--color-ink-black)', fontFamily: 'var(--font-display)' }}>
                  {activeSuppliersCount} <span style={{ fontSize: 14, color: 'var(--color-warm-gray)' }}>/ {suppliers.length}</span>
                </div>
                <div style={{ fontSize: 12, color: 'var(--color-ash-gray)', marginTop: 6 }}>
                  Pemasok siap terima PO
                </div>
              </div>

              <div className="card-panel">
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: 12 }}>
                  <span style={{ fontSize: 12, fontWeight: 500, color: 'var(--color-warm-gray)', textTransform: 'uppercase', letterSpacing: '0.04em' }}>Katalog Produk</span>
                  <Package size={16} color="var(--color-cyan-signal)" />
                </div>
                <div style={{ fontSize: 28, fontWeight: 400, letterSpacing: '-0.5px', color: 'var(--color-ink-black)', fontFamily: 'var(--font-display)' }}>
                  {products.length}
                </div>
                <div style={{ fontSize: 12, color: 'var(--color-ash-gray)', marginTop: 6 }}>
                  Bahan baku & varian terdaftar
                </div>
              </div>
            </div>

            {/* Stok & Workflow side by side */}
            <div style={{ display: 'grid', gridTemplateColumns: '2fr 1fr', gap: 20, marginBottom: 32 }}>
              <div className="card-panel">
                <h3 style={{ fontFamily: 'var(--font-display)', fontSize: 18, fontWeight: 400, letterSpacing: '-0.3px', marginBottom: 16, color: 'var(--color-ink-black)' }}>
                  Ringkasan Saldo Stok Produk
                </h3>
                {stock.length === 0 ? (
                  <div style={{ color: 'var(--color-warm-gray)', fontSize: 14, padding: '32px 0', textAlign: 'center' }}>
                    Belum ada saldo stok. Buat Goods Receipt untuk menambah stok.
                  </div>
                ) : (
                  <table className="data-table">
                    <thead>
                      <tr>
                        <th>SKU</th>
                        <th>Nama Produk</th>
                        <th>Total Masuk</th>
                        <th>Total Keluar</th>
                        <th>Saldo Saat Ini</th>
                      </tr>
                    </thead>
                    <tbody>
                      {stock.map(s => (
                        <tr key={s.product_id}>
                          <td className="mono-cell" style={{ color: 'var(--color-cyan-edge)', fontWeight: 500 }}>{s.product_sku}</td>
                          <td style={{ fontWeight: 500, color: 'var(--color-ink-black)' }}>{s.product_name}</td>
                          <td className="mono-cell">{s.total_in.toLocaleString()} {s.product_unit}</td>
                          <td className="mono-cell">{s.total_out.toLocaleString()} {s.product_unit}</td>
                          <td className="mono-cell" style={{ fontWeight: 600, color: '#15803d' }}>
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
                  <h3 style={{ fontFamily: 'var(--font-display)', fontSize: 18, fontWeight: 400, letterSpacing: '-0.3px', marginBottom: 8, color: 'var(--color-ink-black)' }}>
                    Alur Kerja
                  </h3>
                  <p style={{ color: 'var(--color-warm-gray)', fontSize: 13, lineHeight: 1.6, marginBottom: 20 }}>
                    Validasi bisnis ditegakkan di setiap langkah pengadaan.
                  </p>

                  <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
                    {[
                      { step: '1', title: 'Purchase Request (PR)', desc: 'Kebutuhan diajukan & disetujui pimpinan' },
                      { step: '2', title: 'Purchase Order (PO)', desc: 'Hanya ke supplier berstatus aktif' },
                      { step: '3', title: 'Goods Receipt (QC)', desc: 'Baris lolos QC otomatis tambah stok' },
                    ].map(st => (
                      <div key={st.step} style={{ display: 'flex', gap: 12, alignItems: 'flex-start' }}>
                        <div
                          style={{
                            width: 22,
                            height: 22,
                            borderRadius: '50%',
                            backgroundColor: '#eff9ff',
                            border: '1px solid var(--color-sky-wash)',
                            color: 'var(--color-cyan-edge)',
                            display: 'flex',
                            alignItems: 'center',
                            justifyContent: 'center',
                            fontSize: 11,
                            fontWeight: 600,
                            flexShrink: 0,
                          }}
                        >
                          {st.step}
                        </div>
                        <div>
                          <div style={{ fontSize: 13, fontWeight: 500, color: 'var(--color-ink-black)' }}>{st.title}</div>
                          <div style={{ fontSize: 12, color: 'var(--color-warm-gray)', lineHeight: 1.5 }}>{st.desc}</div>
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
                    Buka Panel Procurement <ChevronRight size={15} />
                  </button>
                </div>
              </div>
            </div>

            {/* Recent Audit Ledger */}
            <div className="card-panel">
              <h3 style={{ fontFamily: 'var(--font-display)', fontSize: 18, fontWeight: 400, letterSpacing: '-0.3px', marginBottom: 16, color: 'var(--color-ink-black)' }}>
                Riwayat Mutasi Buku Besar
              </h3>
              {ledger.length === 0 ? (
                <div style={{ color: 'var(--color-warm-gray)', fontSize: 14, padding: '24px 0', textAlign: 'center' }}>
                  Belum ada transaksi stok tercatat.
                </div>
              ) : (
                <table className="data-table">
                  <thead>
                    <tr>
                      <th>Waktu</th>
                      <th>Referensi</th>
                      <th>Produk</th>
                      <th>Jenis</th>
                      <th>Kuantitas</th>
                      <th>Catatan</th>
                    </tr>
                  </thead>
                  <tbody>
                    {ledger.slice(0, 5).map(tx => (
                      <tr key={tx.id}>
                        <td className="mono-cell">
                          {new Date(tx.created_at).toLocaleString('id-ID')}
                        </td>
                        <td className="mono-cell" style={{ color: 'var(--color-cyan-edge)', fontWeight: 500 }}>{tx.reference_number}</td>
                        <td style={{ color: 'var(--color-ink-black)' }}>{tx.product_name} ({tx.product_sku})</td>
                        <td>
                          <span className={`status-pill ${tx.transaction_type === 'IN' ? 'completed' : 'cancelled'}`}>
                            {tx.transaction_type}
                          </span>
                        </td>
                        <td className="mono-cell" style={{ fontWeight: 600, color: tx.transaction_type === 'IN' ? '#15803d' : '#b91c1c' }}>
                          {tx.transaction_type === 'IN' ? '+' : '-'}{tx.quantity.toLocaleString()} {tx.product_unit}
                        </td>
                        <td style={{ color: 'var(--color-warm-gray)' }}>{tx.notes}</td>
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
                <h1 style={{ fontFamily: 'var(--font-display)', fontSize: 24, fontWeight: 400, letterSpacing: '-0.5px', color: 'var(--color-ink-black)', marginBottom: 4 }}>Katalog Produk & Bahan Baku</h1>
                <p style={{ color: 'var(--color-warm-gray)', fontSize: 14 }}>Kelola master data produk seperti batok kelapa dan arang briket.</p>
              </div>
              {(currentUser?.role === 'ADMIN' || currentUser?.role === 'PURCHASING') && (
                <button onClick={() => setShowProductModal(true)} className="btn-primary">
                  <Plus size={15} /> Tambah Produk
                </button>
              )}
            </div>

            <div className="card-panel">
              <table className="data-table">
                <thead>
                  <tr>
                    <th>ID</th>
                    <th>SKU / Kode</th>
                    <th>Nama Produk</th>
                    <th>Satuan</th>
                    <th>Status</th>
                    <th>Tanggal Registrasi</th>
                  </tr>
                </thead>
                <tbody>
                  {products.map(p => (
                    <tr key={p.id}>
                      <td className="mono-cell">#{p.id}</td>
                      <td className="mono-cell" style={{ color: 'var(--color-cyan-edge)', fontWeight: 500 }}>{p.sku}</td>
                      <td style={{ fontWeight: 500, color: 'var(--color-ink-black)' }}>{p.name}</td>
                      <td style={{ color: 'var(--color-warm-gray)' }}>{p.unit}</td>
                      <td>
                        <span className={`status-pill ${p.is_active ? 'active' : 'inactive'}`}>
                          {p.is_active ? 'Aktif' : 'Nonaktif'}
                        </span>
                      </td>
                      <td className="mono-cell">
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
                <h1 style={{ fontFamily: 'var(--font-display)', fontSize: 24, fontWeight: 400, letterSpacing: '-0.5px', color: 'var(--color-ink-black)', marginBottom: 4 }}>Direktori Rekanan Supplier</h1>
                <p style={{ color: 'var(--color-warm-gray)', fontSize: 14 }}>Kelola data supplier penyedia bahan baku PT AgroSupply.</p>
              </div>
              {(currentUser?.role === 'ADMIN' || currentUser?.role === 'PURCHASING') && (
                <button onClick={() => setShowSupplierModal(true)} className="btn-primary">
                  <Plus size={15} /> Daftarkan Supplier
                </button>
              )}
            </div>

            <div className="card-panel">
              <table className="data-table">
                <thead>
                  <tr>
                    <th>Nama Perusahaan</th>
                    <th>Kontak PIC</th>
                    <th>Email</th>
                    <th>No. Telepon</th>
                    <th>Alamat</th>
                    <th>Status</th>
                  </tr>
                </thead>
                <tbody>
                  {suppliers.map(s => (
                    <tr key={s.id}>
                      <td style={{ fontWeight: 600 }}>{s.name}</td>
                      <td>{s.contact_name || '-'}</td>
                      <td style={{ color: 'var(--color-warm-gray)' }}>{s.email || '-'}</td>
                      <td className="mono-cell">{s.phone || '-'}</td>
                      <td style={{ color: 'var(--color-warm-gray)' }}>{s.address || '-'}</td>
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
                <h1 style={{ fontFamily: 'var(--font-display)', fontSize: 24, fontWeight: 400, letterSpacing: '-0.5px', color: 'var(--color-ink-black)' }}>Pengadaan (Procurement)</h1>
                <p style={{ color: 'var(--color-warm-gray)', fontSize: 14 }}>Alur Purchase Request (PR) dan Purchase Order (PO).</p>
              </div>
              {(currentUser?.role === 'ADMIN' || currentUser?.role === 'PURCHASING') && (
                <div style={{ display: 'flex', gap: 12 }}>
                  <button onClick={() => setShowPRModal(true)} className="btn-secondary">
                    <Plus size={16} /> Buat PR Baru
                  </button>
                  <button onClick={() => setShowPOModal(true)} className="btn-primary">
                    <Plus size={16} /> Terbitkan PO Baru
                  </button>
                </div>
              )}
            </div>

            {/* Purchase Requests Section */}
            <div className="card-panel" style={{ marginBottom: 40 }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
                <h3 style={{ fontFamily: 'var(--font-display)', fontSize: 18, fontWeight: 400, letterSpacing: '-0.3px', color: 'var(--color-ink-black)' }}>1. Dokumen Purchase Request (PR)</h3>
                <span style={{ fontSize: 13, color: 'var(--color-warm-gray)' }}>Total: {prs.length} pengajuan</span>
              </div>

              {prs.length === 0 ? (
                <div style={{ color: 'var(--color-warm-gray)', padding: '20px 0', textAlign: 'center' }}>Belum ada Purchase Request.</div>
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
                        <td className="mono-cell" style={{ color: 'var(--color-cyan-edge)', fontWeight: 600 }}>{pr.pr_number}</td>
                        <td>
                          <span className={`status-pill ${pr.status.toLowerCase()}`}>{pr.status}</span>
                        </td>
                        <td>{pr.notes || '-'}</td>
                        <td style={{ color: 'var(--color-warm-gray)' }}>{new Date(pr.created_at).toLocaleDateString('id-ID')}</td>
                        <td>
                          <div style={{ display: 'flex', gap: 8 }}>
                            {pr.status === 'DRAFT' && (currentUser?.role === 'ADMIN' || currentUser?.role === 'PURCHASING') && (
                              <button onClick={() => handlePRAction(pr.id, 'submit')} className="btn-secondary" style={{ padding: '6px 12px', fontSize: 12 }}>
                                Submit PR
                              </button>
                            )}
                            {pr.status === 'SUBMITTED' && (currentUser?.role === 'ADMIN' || currentUser?.role === 'MANAGER') && (
                              <>
                                <button onClick={() => handlePRAction(pr.id, 'approve')} className="btn-primary" style={{ padding: '6px 12px', fontSize: 12 }}>
                                  Approve
                                </button>
                                <button onClick={() => handlePRAction(pr.id, 'reject')} className="btn-secondary" style={{ padding: '6px 12px', fontSize: 12, color: '#b91c1c' }}>
                                  Reject
                                </button>
                              </>
                            )}
                            {pr.status === 'SUBMITTED' && currentUser?.role !== 'ADMIN' && currentUser?.role !== 'MANAGER' && (
                              <span style={{ fontSize: 12, color: 'var(--color-warm-gray)' }}>Menunggu Approval Manager</span>
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
                <h3 style={{ fontFamily: 'var(--font-display)', fontSize: 18, fontWeight: 400, letterSpacing: '-0.3px', color: 'var(--color-ink-black)' }}>2. Dokumen Purchase Order (PO)</h3>
                <span style={{ fontSize: 13, color: 'var(--color-warm-gray)' }}>Total: {pos.length} pesanan</span>
              </div>

              {pos.length === 0 ? (
                <div style={{ color: 'var(--color-warm-gray)', padding: '20px 0', textAlign: 'center' }}>Belum ada Purchase Order.</div>
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
                        <td className="mono-cell" style={{ color: 'var(--color-cyan-edge)', fontWeight: 600 }}>{po.po_number}</td>
                        <td style={{ fontWeight: 500 }}>{po.supplier_name}</td>
                        <td>
                          <span className={`status-pill ${po.status.toLowerCase()}`}>{po.status}</span>
                        </td>
                        <td className="mono-cell" style={{ fontWeight: 600 }}>
                          Rp {po.total_amount.toLocaleString('id-ID')}
                        </td>
                        <td style={{ color: 'var(--color-warm-gray)' }}>{new Date(po.created_at).toLocaleDateString('id-ID')}</td>
                        <td>
                          <div style={{ display: 'flex', gap: 8 }}>
                            {po.status === 'DRAFT' && (currentUser?.role === 'ADMIN' || currentUser?.role === 'MANAGER') && (
                              <button onClick={() => handlePOAction(po.id, 'approve')} className="btn-primary" style={{ padding: '6px 12px', fontSize: 12 }}>
                                Approve PO
                              </button>
                            )}
                            {po.status === 'DRAFT' && currentUser?.role !== 'ADMIN' && currentUser?.role !== 'MANAGER' && (
                              <span style={{ fontSize: 12, color: 'var(--color-warm-gray)' }}>Menunggu Approval Manager</span>
                            )}
                            {(po.status === 'APPROVED' || po.status === 'PARTIALLY_RECEIVED') && (currentUser?.role === 'ADMIN' || currentUser?.role === 'WAREHOUSE') && (
                              <button
                                onClick={() => {
                                  handleSelectPOForGR(po.id);
                                  setShowGRModal(true);
                                }}
                                className="btn-secondary"
                                style={{ padding: '6px 12px', fontSize: 12, borderColor: 'var(--color-cyan-edge)' }}
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
                <h1 style={{ fontFamily: 'var(--font-display)', fontSize: 24, fontWeight: 400, letterSpacing: '-0.5px', color: 'var(--color-ink-black)' }}>Inventaris & Buku Besar Stok</h1>
                <p style={{ color: 'var(--color-warm-gray)', fontSize: 14 }}>
                  Pencatatan Goods Receipt (inspeksi QC fisik) dan audit mutasi saldo stok.
                </p>
              </div>
              {(currentUser?.role === 'ADMIN' || currentUser?.role === 'WAREHOUSE') && (
                <button
                  onClick={() => {
                    const validPOs = pos.filter(p => p.status === 'APPROVED' || p.status === 'PARTIALLY_RECEIVED');
                    if (validPOs.length > 0) {
                      const firstValidPO = validPOs[0];
                      setGRForm({
                        purchase_order_id: firstValidPO.id,
                        delivery_order_number: '',
                        notes: '',
                        items: [],
                      });
                      setShowGRModal(true);
                      handleSelectPOForGR(firstValidPO.id);
                    } else {
                      notify('Tidak ada PO berstatus APPROVED atau PARTIALLY_RECEIVED untuk diterima', 'error');
                    }
                  }}
                  className="btn-primary"
                >
                  <Plus size={16} /> Catat Goods Receipt Baru
                </button>
              )}
            </div>

            {/* Saldo Stok Cards */}
            <div className="card-panel" style={{ marginBottom: 40 }}>
              <h3 style={{ fontFamily: 'var(--font-display)', fontSize: 18, fontWeight: 400, letterSpacing: '-0.3px', marginBottom: 16, color: 'var(--color-ink-black)' }}>Saldo Stok Terkini Gudang</h3>
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
                      <td className="mono-cell" style={{ color: 'var(--color-cyan-edge)', fontWeight: 600 }}>{s.product_sku}</td>
                      <td style={{ fontWeight: 600 }}>{s.product_name}</td>
                      <td className="mono-cell">{s.total_in.toLocaleString()} {s.product_unit}</td>
                      <td className="mono-cell">{s.total_out.toLocaleString()} {s.product_unit}</td>
                      <td className="mono-cell" style={{ fontWeight: 700, fontSize: 16, color: '#15803d' }}>
                        {s.current_stock.toLocaleString()} {s.product_unit}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>

            {/* Riwayat Dokumen Goods Receipt */}
            <div className="card-panel" style={{ marginBottom: 40 }}>
              <h3 style={{ fontFamily: 'var(--font-display)', fontSize: 18, fontWeight: 400, letterSpacing: '-0.3px', marginBottom: 16, color: 'var(--color-ink-black)' }}>Dokumen Penerimaan Barang (Goods Receipt)</h3>
              {goodsReceipts.length === 0 ? (
                <div style={{ color: 'var(--color-warm-gray)', padding: '20px 0', textAlign: 'center' }}>Belum ada Goods Receipt tercatat.</div>
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
                        <td className="mono-cell" style={{ color: 'var(--color-cyan-edge)', fontWeight: 600 }}>{gr.gr_number}</td>
                        <td className="mono-cell">{gr.po_number || `#${gr.purchase_order_id}`}</td>
                        <td className="mono-cell">{gr.delivery_order_number}</td>
                        <td style={{ color: 'var(--color-warm-gray)' }}>{new Date(gr.receipt_date).toLocaleString('id-ID')}</td>
                        <td style={{ color: 'var(--color-warm-gray)' }}>{gr.notes || '-'}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              )}
            </div>

            {/* Inventory Ledger Timeline */}
            <div className="card-panel">
              <h3 style={{ fontFamily: 'var(--font-display)', fontSize: 18, fontWeight: 400, letterSpacing: '-0.3px', marginBottom: 16, color: 'var(--color-ink-black)' }}>Buku Besar Transaksi Stok (Audit Trail)</h3>
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
                      <td className="mono-cell" style={{ color: 'var(--color-warm-gray)' }}>
                        {new Date(tx.created_at).toLocaleString('id-ID')}
                      </td>
                      <td className="mono-cell" style={{ color: 'var(--color-cyan-edge)' }}>{tx.reference_number}</td>
                      <td>{tx.product_name} ({tx.product_sku})</td>
                      <td>
                        <span className={`status-pill ${tx.transaction_type === 'IN' ? 'completed' : 'cancelled'}`}>
                          {tx.transaction_type}
                        </span>
                      </td>
                      <td className="mono-cell" style={{ fontWeight: 600, color: tx.transaction_type === 'IN' ? '#15803d' : '#b91c1c' }}>
                        {tx.transaction_type === 'IN' ? '+' : '-'}{tx.quantity.toLocaleString()} {tx.product_unit}
                      </td>
                      <td style={{ color: 'var(--color-warm-gray)' }}>{tx.notes}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        )}

        {/* ============================================================== */}
        {/* TAB 6: PENGGUNA & ROLE-BASED ACCESS CONTROL (ADMIN ONLY)       */}
        {/* ============================================================== */}
        {activeTab === 'users' && currentUser?.role === 'ADMIN' && (
          <div>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 28 }}>
              <div>
                <h1 style={{ fontFamily: 'var(--font-display)', fontSize: 24, fontWeight: 400, letterSpacing: '-0.5px', color: 'var(--color-ink-black)' }}>Pengguna & Hak Akses (RBAC)</h1>
                <p style={{ color: 'var(--color-warm-gray)', fontSize: 14 }}>
                  Manajemen akun karyawan, penetapan peran (Admin, Purchasing, Manager, Warehouse), dan kontrol keaktifan.
                </p>
              </div>
              <button onClick={() => setShowUserModal(true)} className="btn-primary">
                <Plus size={16} /> Daftarkan Pengguna Baru
              </button>
            </div>

            <div className="card-panel">
              <table className="data-table">
                <thead>
                  <tr>
                    <th>ID</th>
                    <th>NAMA LENGKAP</th>
                    <th>USERNAME</th>
                    <th>EMAIL</th>
                    <th>PERAN (ROLE)</th>
                    <th>STATUS</th>
                    <th>TANGGAL TERDAFTAR</th>
                    <th>KONTROL STATUS</th>
                  </tr>
                </thead>
                <tbody>
                  {userList.map(u => (
                    <tr key={u.id}>
                      <td className="mono-cell">#{u.id}</td>
                      <td style={{ fontWeight: 600 }}>{u.full_name}</td>
                      <td className="mono-cell" style={{ color: 'var(--color-cyan-edge)' }}>{u.username}</td>
                      <td style={{ color: 'var(--color-warm-gray)' }}>{u.email}</td>
                      <td>
                        <span
                          className="status-pill"
                          style={{
                            backgroundColor: `${getRoleColor(u.role)}20`,
                            color: getRoleColor(u.role),
                            fontWeight: 600,
                          }}
                        >
                          {u.role}
                        </span>
                      </td>
                      <td>
                        <span className={`status-pill ${u.is_active ? 'active' : 'inactive'}`}>
                          {u.is_active ? 'Aktif' : 'Nonaktif'}
                        </span>
                      </td>
                      <td style={{ color: 'var(--color-warm-gray)' }}>
                        {new Date(u.created_at).toLocaleDateString('id-ID')}
                      </td>
                      <td>
                        {u.username !== 'admin' ? (
                          <button
                            onClick={() => handleToggleUserStatus(u)}
                            className="btn-secondary"
                            style={{ padding: '4px 10px', fontSize: 12 }}
                          >
                            {u.is_active ? 'Nonaktifkan' : 'Aktifkan'}
                          </button>
                        ) : (
                          <span style={{ fontSize: 12, color: 'var(--color-warm-gray)' }}>Primary Superadmin</span>
                        )}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        )}
      </>
    )}
  </main>

      {/* ============================================================== */}
      {/* MODAL 1: TAMBAH PRODUK                                         */}
      {/* ============================================================== */}
      {showProductModal && (
        <div style={{ position: 'fixed', inset: 0, backgroundColor: 'rgba(12,10,9,0.3)', backdropFilter: 'blur(6px)', zIndex: 100, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
          <div className="card-elevated" style={{ width: 480, maxWidth: '90%' }}>
            <h3 style={{ fontFamily: 'var(--font-display)', fontSize: 20, fontWeight: 400, letterSpacing: '-0.4px', marginBottom: 16, color: 'var(--color-ink-black)' }}>Tambah Produk Baru</h3>
            <form onSubmit={handleCreateProduct}>
              <div style={{ marginBottom: 16 }}>
                <label style={{ display: 'block', fontSize: 12, fontWeight: 500, color: 'var(--color-warm-gray)', marginBottom: 6, letterSpacing: '0.03em' }}>KODE SKU *</label>
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
                <label style={{ display: 'block', fontSize: 12, fontWeight: 500, color: 'var(--color-warm-gray)', marginBottom: 6, letterSpacing: '0.03em' }}>NAMA PRODUK *</label>
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
                <label style={{ display: 'block', fontSize: 12, fontWeight: 500, color: 'var(--color-warm-gray)', marginBottom: 6, letterSpacing: '0.03em' }}>SATUAN / UNIT *</label>
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
        <div style={{ position: 'fixed', inset: 0, backgroundColor: 'rgba(12,10,9,0.3)', backdropFilter: 'blur(6px)', zIndex: 100, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
          <div className="card-elevated" style={{ width: 520, maxWidth: '90%' }}>
            <h3 style={{ fontFamily: 'var(--font-display)', fontSize: 20, fontWeight: 400, letterSpacing: '-0.4px', marginBottom: 16, color: 'var(--color-ink-black)' }}>Daftarkan Supplier Baru</h3>
            <form onSubmit={handleCreateSupplier}>
              <div style={{ marginBottom: 14 }}>
                <label style={{ display: 'block', fontSize: 12, fontWeight: 500, color: 'var(--color-warm-gray)', marginBottom: 6, letterSpacing: '0.03em' }}>NAMA PERUSAHAAN / CV *</label>
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
                  <label style={{ display: 'block', fontSize: 12, fontWeight: 500, color: 'var(--color-warm-gray)', marginBottom: 6, letterSpacing: '0.03em' }}>NAMA KONTAK PIC</label>
                  <input
                    type="text"
                    placeholder="Misal: Pak Budi"
                    value={supplierForm.contact_name}
                    onChange={e => setSupplierForm({ ...supplierForm, contact_name: e.target.value })}
                    style={{ width: '100%' }}
                  />
                </div>
                <div>
                  <label style={{ display: 'block', fontSize: 12, fontWeight: 500, color: 'var(--color-warm-gray)', marginBottom: 6, letterSpacing: '0.03em' }}>NO. TELEPON / WA</label>
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
                <label style={{ display: 'block', fontSize: 12, fontWeight: 500, color: 'var(--color-warm-gray)', marginBottom: 6, letterSpacing: '0.03em' }}>EMAIL</label>
                <input
                  type="email"
                  placeholder="budi@agrolestari.com"
                  value={supplierForm.email}
                  onChange={e => setSupplierForm({ ...supplierForm, email: e.target.value })}
                  style={{ width: '100%' }}
                />
              </div>
              <div style={{ marginBottom: 24 }}>
                <label style={{ display: 'block', fontSize: 12, fontWeight: 500, color: 'var(--color-warm-gray)', marginBottom: 6, letterSpacing: '0.03em' }}>ALAMAT GUDANG / KANTOR</label>
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
        <div style={{ position: 'fixed', inset: 0, backgroundColor: 'rgba(12,10,9,0.3)', backdropFilter: 'blur(6px)', zIndex: 100, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
          <div className="card-elevated" style={{ width: 560, maxWidth: '90%' }}>
            <h3 style={{ fontFamily: 'var(--font-display)', fontSize: 20, fontWeight: 400, letterSpacing: '-0.4px', marginBottom: 16, color: 'var(--color-ink-black)' }}>Buat Purchase Request (PR) Baru</h3>
            <form onSubmit={handleCreatePR}>
              <div style={{ marginBottom: 16 }}>
                <label style={{ display: 'block', fontSize: 12, fontWeight: 500, color: 'var(--color-warm-gray)', marginBottom: 6, letterSpacing: '0.03em' }}>CATATAN KEBUTUHAN PENGADAAN</label>
                <input
                  type="text"
                  placeholder="Misal: Stok produksi batok kelapa minggu ke-4"
                  value={prForm.notes}
                  onChange={e => setPRForm({ ...prForm, notes: e.target.value })}
                  style={{ width: '100%' }}
                />
              </div>

              <div style={{ marginBottom: 20 }}>
                <label style={{ display: 'block', fontSize: 13, color: 'var(--color-warm-gray)', marginBottom: 8 }}>BARANG YANG DIBUTUHKAN</label>
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
        <div style={{ position: 'fixed', inset: 0, backgroundColor: 'rgba(12,10,9,0.3)', backdropFilter: 'blur(6px)', zIndex: 100, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
          <div className="card-elevated" style={{ width: 600, maxWidth: '90%' }}>
            <h3 style={{ fontFamily: 'var(--font-display)', fontSize: 20, fontWeight: 400, letterSpacing: '-0.4px', marginBottom: 16, color: 'var(--color-ink-black)' }}>Terbitkan Purchase Order (PO) Baru</h3>
            <form onSubmit={handleCreatePO}>
              <div style={{ marginBottom: 14 }}>
                <label style={{ display: 'block', fontSize: 12, fontWeight: 500, color: 'var(--color-warm-gray)', marginBottom: 6, letterSpacing: '0.03em' }}>PILIH SUPPLIER AKTIF *</label>
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
                <label style={{ display: 'block', fontSize: 12, fontWeight: 500, color: 'var(--color-warm-gray)', marginBottom: 6, letterSpacing: '0.03em' }}>RUJUKAN PR TERSETUJUI (OPSIONAL)</label>
                <select
                  value={poForm.purchase_request_id}
                  onChange={async (e) => {
                    const prId = Number(e.target.value);
                    if (prId > 0) {
                      try {
                        const prDetail = await api.getPR(prId);
                        if (prDetail && prDetail.items && prDetail.items.length > 0) {
                          setPOForm(prev => ({
                            ...prev,
                            purchase_request_id: prId,
                            items: prDetail.items!.map(it => ({
                              product_id: it.product_id,
                              quantity: it.quantity,
                              unit_price: 2500,
                            })),
                          }));
                          return;
                        }
                      } catch (err: any) {
                        notify(`Gagal memuat item PR: ${err.message}`, 'error');
                      }
                    }
                    setPOForm(prev => ({
                      ...prev,
                      purchase_request_id: prId,
                    }));
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
                <label style={{ display: 'block', fontSize: 13, color: 'var(--color-warm-gray)', marginBottom: 8 }}>BARIS PESANAN & HARGA BELI</label>
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
                <label style={{ display: 'block', fontSize: 12, fontWeight: 500, color: 'var(--color-warm-gray)', marginBottom: 6, letterSpacing: '0.03em' }}>CATATAN PENGIRIMAN</label>
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
        <div style={{ position: 'fixed', inset: 0, backgroundColor: 'rgba(12,10,9,0.3)', backdropFilter: 'blur(6px)', zIndex: 100, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
          <div className="card-elevated" style={{ width: 680, maxWidth: '90%' }}>
            <h3 style={{ fontSize: 20, fontWeight: 600, marginBottom: 8 }}>Penerimaan Barang Fisik (Goods Receipt)</h3>
            <p style={{ color: 'var(--color-warm-gray)', fontSize: 13, marginBottom: 20 }}>
              Pemeriksaan QC fisik truk. Kuantitas yang lolos (accepted) akan otomatis masuk ke Buku Besar Stok.
            </p>

            <form onSubmit={handleCreateGR}>
              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12, marginBottom: 16 }}>
                <div>
                  <label style={{ display: 'block', fontSize: 12, fontWeight: 500, color: 'var(--color-warm-gray)', marginBottom: 6, letterSpacing: '0.03em' }}>PILIH PURCHASE ORDER (PO)</label>
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
                  <label style={{ display: 'block', fontSize: 12, fontWeight: 500, color: 'var(--color-warm-gray)', marginBottom: 6, letterSpacing: '0.03em' }}>NO. SURAT JALAN SUPPLIER *</label>
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
                <label style={{ display: 'block', fontSize: 13, color: 'var(--color-warm-gray)', marginBottom: 8 }}>HASIL INSPEKSI MUTU & TIMBANGAN (QC)</label>
                {loadingPOItems ? (
                  <div style={{ padding: '24px', textAlign: 'center', color: 'var(--color-warm-gray)', backgroundColor: 'var(--color-carbon)', borderRadius: 8, fontSize: 13 }}>
                    Sedang memuat daftar barang dari PO...
                  </div>
                ) : grForm.items.length === 0 ? (
                  <div style={{ padding: '24px', textAlign: 'center', color: 'var(--color-warm-gray)', backgroundColor: 'var(--color-carbon)', borderRadius: 8, fontSize: 13 }}>
                    {grForm.purchase_order_id <= 0 
                      ? 'Silakan pilih Purchase Order di atas untuk memuat daftar barang.' 
                      : 'Purchase Order ini tidak memiliki item barang yang dapat diterima.'}
                  </div>
                ) : (
                  grForm.items.map((item, idx) => (
                    <div key={idx} style={{ backgroundColor: 'var(--color-carbon)', padding: 14, borderRadius: 8, marginBottom: 12 }}>
                      <div style={{ fontWeight: 600, fontSize: 14, marginBottom: 10, color: 'var(--color-ink-black)' }}>
                        {item.product_name || `Produk ID #${item.product_id}`}
                      </div>
                    <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr 1fr', gap: 10, marginBottom: 10 }}>
                      <div>
                        <span style={{ fontSize: 11, color: 'var(--color-warm-gray)' }}>TOTAL TIBA (KG)</span>
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
                        <span style={{ fontSize: 11, color: '#15803d' }}>LOLOS QC (MASUK STOK)</span>
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
                        <span style={{ fontSize: 11, color: '#b91c1c' }}>DITOLAK (CACAT/KOTOR)</span>
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
                ))
              )}
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

      {/* ============================================================== */}
      {/* MODAL 6: DAFTARKAN PENGGUNA BARU (ADMIN ONLY)                   */}
      {/* ============================================================== */}
      {showUserModal && (
        <div style={{ position: 'fixed', inset: 0, backgroundColor: 'rgba(12,10,9,0.3)', backdropFilter: 'blur(6px)', zIndex: 100, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
          <div className="card-elevated" style={{ width: 520, maxWidth: '90%' }}>
            <h3 style={{ fontFamily: 'var(--font-display)', fontSize: 20, fontWeight: 400, letterSpacing: '-0.4px', marginBottom: 16, color: 'var(--color-ink-black)' }}>Daftarkan Pengguna Baru</h3>
            <form onSubmit={handleCreateUser}>
              <div style={{ marginBottom: 16 }}>
                <label style={{ display: 'block', fontSize: 12, fontWeight: 500, color: 'var(--color-warm-gray)', marginBottom: 6, letterSpacing: '0.03em' }}>NAMA LENGKAP *</label>
                <input
                  type="text"
                  required
                  placeholder="Misal: Ahmad Purchasing Officer"
                  value={userForm.full_name}
                  onChange={e => setUserForm({ ...userForm, full_name: e.target.value })}
                  style={{ width: '100%' }}
                />
              </div>

              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 14, marginBottom: 16 }}>
                <div>
                  <label style={{ display: 'block', fontSize: 12, fontWeight: 500, color: 'var(--color-warm-gray)', marginBottom: 6, letterSpacing: '0.03em' }}>USERNAME *</label>
                  <input
                    type="text"
                    required
                    placeholder="Misal: ahmad_purchasing"
                    value={userForm.username}
                    onChange={e => setUserForm({ ...userForm, username: e.target.value })}
                    style={{ width: '100%' }}
                  />
                </div>
                <div>
                  <label style={{ display: 'block', fontSize: 12, fontWeight: 500, color: 'var(--color-warm-gray)', marginBottom: 6, letterSpacing: '0.03em' }}>PASSWORD *</label>
                  <input
                    type="password"
                    required
                    placeholder="Minimal 6 karakter"
                    value={userForm.password}
                    onChange={e => setUserForm({ ...userForm, password: e.target.value })}
                    style={{ width: '100%' }}
                  />
                </div>
              </div>

              <div style={{ display: 'grid', gridTemplateColumns: '1.2fr 1fr', gap: 14, marginBottom: 24 }}>
                <div>
                  <label style={{ display: 'block', fontSize: 12, fontWeight: 500, color: 'var(--color-warm-gray)', marginBottom: 6, letterSpacing: '0.03em' }}>EMAIL *</label>
                  <input
                    type="email"
                    required
                    placeholder="nama@perusahaan.com"
                    value={userForm.email}
                    onChange={e => setUserForm({ ...userForm, email: e.target.value })}
                    style={{ width: '100%' }}
                  />
                </div>
                <div>
                  <label style={{ display: 'block', fontSize: 12, fontWeight: 500, color: 'var(--color-warm-gray)', marginBottom: 6, letterSpacing: '0.03em' }}>PERAN (ROLE) *</label>
                  <select
                    value={userForm.role}
                    onChange={e => setUserForm({ ...userForm, role: e.target.value as UserRole })}
                    style={{ width: '100%' }}
                  >
                    <option value="PURCHASING">PURCHASING</option>
                    <option value="MANAGER">MANAGER</option>
                    <option value="WAREHOUSE">WAREHOUSE</option>
                    <option value="ADMIN">ADMIN</option>
                  </select>
                </div>
              </div>

              <div style={{ display: 'flex', justifyContent: 'flex-end', gap: 12 }}>
                <button type="button" onClick={() => setShowUserModal(false)} className="btn-secondary">
                  Batal
                </button>
                <button type="submit" className="btn-primary">
                  Daftarkan Akun
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
