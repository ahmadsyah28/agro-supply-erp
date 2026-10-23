// API Client for PT AgroSupply ERP Backend (Go API :8080)

export interface Product {
  id: number;
  sku: string;
  name: string;
  unit: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface Supplier {
  id: number;
  name: string;
  contact_name: string;
  email: string;
  phone: string;
  address: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface PurchaseRequestItem {
  id: number;
  purchase_request_id: number;
  product_id: number;
  product_name?: string;
  product_sku?: string;
  quantity: number;
  notes: string;
  created_at: string;
}

export interface PurchaseRequest {
  id: number;
  pr_number: string;
  status: 'DRAFT' | 'SUBMITTED' | 'APPROVED' | 'REJECTED';
  notes: string;
  items?: PurchaseRequestItem[];
  created_at: string;
  updated_at: string;
}

export interface PurchaseOrderItem {
  id: number;
  purchase_order_id: number;
  product_id: number;
  product_name?: string;
  product_sku?: string;
  quantity: number;
  unit_price: number;
  subtotal: number;
  created_at: string;
}

export interface PurchaseOrder {
  id: number;
  po_number: string;
  purchase_request_id?: number;
  supplier_id: number;
  supplier_name?: string;
  status: 'DRAFT' | 'APPROVED' | 'PARTIALLY_RECEIVED' | 'COMPLETED' | 'CANCELLED';
  total_amount: number;
  notes: string;
  items?: PurchaseOrderItem[];
  created_at: string;
  updated_at: string;
}

export interface GoodsReceiptItem {
  id: number;
  goods_receipt_id: number;
  purchase_order_item_id: number;
  product_id: number;
  product_name?: string;
  product_sku?: string;
  product_unit?: string;
  received_quantity: number;
  accepted_quantity: number;
  rejected_quantity: number;
  rejection_reason?: string;
  created_at: string;
}

export interface GoodsReceipt {
  id: number;
  gr_number: string;
  purchase_order_id: number;
  po_number?: string;
  delivery_order_number: string;
  receipt_date: string;
  notes: string;
  items?: GoodsReceiptItem[];
  created_at: string;
}

export interface StockSummary {
  product_id: number;
  product_name: string;
  product_sku: string;
  product_unit: string;
  total_in: number;
  total_out: number;
  current_stock: number;
}

export interface InventoryTransaction {
  id: number;
  product_id: number;
  product_name: string;
  product_sku: string;
  product_unit: string;
  transaction_type: 'IN' | 'OUT' | 'ADJUSTMENT';
  reference_type: string;
  reference_id: number;
  reference_number: string;
  quantity: number;
  notes: string;
  created_at: string;
}

const API_BASE = '/api/v1';

async function request<T>(endpoint: string, options?: RequestInit): Promise<T> {
  const url = `${API_BASE}${endpoint}`;
  const response = await fetch(url, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
  });

  const json = await response.json();
  if (!response.ok || json.status === 'error') {
    const errorMsg = json.error?.message || `Request failed with status ${response.status}`;
    throw new Error(errorMsg);
  }

  return json.data as T;
}

// Products API
export const api = {
  // Health
  checkHealth: async () => {
    try {
      const res = await fetch('/health/ready');
      return res.ok;
    } catch {
      return false;
    }
  },

  // Products
  getProducts: () => request<Product[]>('/products'),
  getProduct: (id: number) => request<Product>(`/products/${id}`),
  createProduct: (data: { sku: string; name: string; unit: string }) =>
    request<Product>('/products', { method: 'POST', body: JSON.stringify(data) }),
  updateProduct: (id: number, data: { name: string; unit: string; is_active?: boolean }) =>
    request<Product>(`/products/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  deleteProduct: (id: number) => request<{ message: string }>(`/products/${id}`, { method: 'DELETE' }),

  // Suppliers
  getSuppliers: (activeOnly = false) =>
    request<Supplier[]>(`/suppliers${activeOnly ? '?active=true' : ''}`),
  getSupplier: (id: number) => request<Supplier>(`/suppliers/${id}`),
  createSupplier: (data: { name: string; contact_name: string; email: string; phone: string; address: string }) =>
    request<Supplier>('/suppliers', { method: 'POST', body: JSON.stringify(data) }),
  updateSupplier: (id: number, data: { name: string; contact_name: string; email: string; phone: string; address: string; is_active?: boolean }) =>
    request<Supplier>(`/suppliers/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  deleteSupplier: (id: number) => request<{ message: string }>(`/suppliers/${id}`, { method: 'DELETE' }),

  // Purchase Requests
  getPRs: () => request<PurchaseRequest[]>('/purchase-requests'),
  getPR: (id: number) => request<PurchaseRequest>(`/purchase-requests/${id}`),
  createPR: (data: { notes: string; items: { product_id: number; quantity: number; notes: string }[] }) =>
    request<PurchaseRequest>('/purchase-requests', { method: 'POST', body: JSON.stringify(data) }),
  submitPR: (id: number) => request<PurchaseRequest>(`/purchase-requests/${id}/submit`, { method: 'POST' }),
  approvePR: (id: number) => request<PurchaseRequest>(`/purchase-requests/${id}/approve`, { method: 'POST' }),
  rejectPR: (id: number) => request<PurchaseRequest>(`/purchase-requests/${id}/reject`, { method: 'POST' }),

  // Purchase Orders
  getPOs: () => request<PurchaseOrder[]>('/purchase-orders'),
  getPO: (id: number) => request<PurchaseOrder>(`/purchase-orders/${id}`),
  createPO: (data: { purchase_request_id?: number; supplier_id: number; notes: string; items: { product_id: number; quantity: number; unit_price: number }[] }) =>
    request<PurchaseOrder>('/purchase-orders', { method: 'POST', body: JSON.stringify(data) }),
  approvePO: (id: number) => request<PurchaseOrder>(`/purchase-orders/${id}/approve`, { method: 'POST' }),
  cancelPO: (id: number) => request<PurchaseOrder>(`/purchase-orders/${id}/cancel`, { method: 'POST' }),

  // Inventory & Stock
  getGoodsReceipts: () => request<GoodsReceipt[]>('/goods-receipts'),
  getGoodsReceipt: (id: number) => request<GoodsReceipt>(`/goods-receipts/${id}`),
  createGoodsReceipt: (data: {
    purchase_order_id: number;
    delivery_order_number: string;
    notes: string;
    items: {
      purchase_order_item_id: number;
      product_id: number;
      received_quantity: number;
      accepted_quantity: number;
      rejected_quantity: number;
      rejection_reason?: string;
    }[];
  }) => request<GoodsReceipt>('/goods-receipts', { method: 'POST', body: JSON.stringify(data) }),
  getStockSummary: () => request<StockSummary[]>('/inventory/stock'),
  getLedger: (productId?: number) =>
    request<InventoryTransaction[]>(`/inventory/ledger${productId ? `?product_id=${productId}` : ''}`),
};
