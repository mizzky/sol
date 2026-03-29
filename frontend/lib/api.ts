// Fallback to localhost:8080 if env is not available in dev client bundles
export const API_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

export type ApiError = { status: number } & Record<string, unknown>;

export interface LoginResponse {
    message?: string;
    token: string;
    user: {
        id: number;
        name: string;
        email: string;
        role: "admin" | "member";
    };
}

export interface Product {
  id: number;
  name: string;
  price: number;
  is_available: boolean;
  category_id: number;
  sku: string;
  description?: string | null;
  image_url?: string | null;
  stock_quantity: number;
  created_at: string;
  updated_at: string;
}

export interface CreateProductRequest {
  name: string;
  price: number;
  is_available: boolean;
  category_id: number;
  sku: string;
  description?: string | null;
  image_url?: string | null;
  stock_quantity: number;
}

export type UpdateProductRequest = CreateProductRequest;

export interface Category {
  id: number;
  name: string;
  description?: string | null;
}

export interface CreateCategoryRequest {
  name: string;
  description?: string | null;
}

export interface UpdateCategoryRequest {
  name: string;
  description?: string | null;
}

export interface OrderSummary {
  id: number;
  user_id?: number;
  total: number;
  status: string;
  created_at?: string;
  updated_at?: string;
  cancelled_at?: string | null;
}

export interface OrderItemDetail {
  id?: number;
  order_id?: number;
  product_id: number;
  quantity: number;
  unit_price: number;
  product_name_snapshot?: string;
}

export interface OrderWithItems {
  order: OrderSummary;
  items: OrderItemDetail[];
}

export interface CartItem {
  id: number;
  cart_id: number;
  product_id: number;
  quantity: number;
  price: number;
  created_at?: string;
  updated_at?: string;
  product_name?: string | null;
  product_price?: number | null;
  product_stock?: number | null;
}

export interface CartResponse {
  items: CartItem[];
}

interface CartItemMutationResponse {
  item: CartItem;
}

interface CategoriesResponse {
  categories?: Category[];
}

interface OrdersResponse {
  orders?: unknown[];
}


async function parseJsonSafe<T = Record<string, unknown>>(res: Response): Promise<T | Record<string, unknown>> {
  try {
    return await res.json();
  } catch {
    return {};
  }
}

function toNumber(value: unknown, fallback = 0): number {
  if (typeof value === "number" && Number.isFinite(value)) {
    return value;
  }
  if (typeof value === "string") {
    const parsed = Number(value);
    if (Number.isFinite(parsed)) {
      return parsed;
    }
  }
  return fallback;
}

function toStringValue(value: unknown, fallback = ""): string {
  return typeof value === "string" ? value : fallback;
}

function normalizeOrderSummary(raw: Record<string, unknown>): OrderSummary {
  return {
    id: toNumber(raw.id ?? raw.ID),
    user_id: toNumber(raw.user_id ?? raw.UserID),
    total: toNumber(raw.total ?? raw.Total),
    status: toStringValue(raw.status ?? raw.Status),
    created_at: toStringValue(raw.created_at ?? raw.CreatedAt),
    updated_at: toStringValue(raw.updated_at ?? raw.UpdatedAt),
    cancelled_at: (raw.cancelled_at ?? raw.CancelledAt ?? null) as string | null,
  };
}

function normalizeOrderItem(raw: Record<string, unknown>): OrderItemDetail {
  return {
    id: toNumber(raw.id ?? raw.ID),
    order_id: toNumber(raw.order_id ?? raw.OrderID),
    product_id: toNumber(raw.product_id ?? raw.ProductID),
    quantity: toNumber(raw.quantity ?? raw.Quantity),
    unit_price: toNumber(raw.unit_price ?? raw.UnitPrice),
    product_name_snapshot: toStringValue(raw.product_name_snapshot ?? raw.ProductNameSnapshot),
  };
}

function normalizeOrderWithItems(raw: unknown): OrderWithItems {
  const source = (raw ?? {}) as Record<string, unknown>;
  const orderRaw = (source.order ?? source.Order ?? {}) as Record<string, unknown>;
  const itemsRaw = Array.isArray(source.items ?? source.Items)
    ? ((source.items ?? source.Items) as unknown[])
    : [];
  return {
    order: normalizeOrderSummary(orderRaw),
    items: itemsRaw.map((item) => normalizeOrderItem((item ?? {}) as Record<string, unknown>)),
  };
}

/**
 * 認証付きfetch共通関数
 * localStorageからトークンを自動的に取得してAuthorizationヘッダーに付与
 * 401エラー時は自動的にログアウトを実行
 */
export async function fetchWithAuth(
  url: string,
  options: RequestInit = {}
): Promise<Response> {
  const token = typeof window !== "undefined" ? localStorage.getItem("auth_token") : null;

  const headers: Record<string, string> = {
    "Content-Type": "application/json",
  };

  // 既存のヘッダーをマージ
  if (options.headers) {
    const existingHeaders = options.headers as Record<string, string>;
    Object.assign(headers, existingHeaders);
  }

  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  const response = await fetch(url, {
    ...options,
    headers,
  });

  // 401エラー時は自動ログアウト
  if (response.status === 401) {
    // 動的インポートで循環参照を回避
    if (typeof window !== "undefined") {
      const { useAuthStore } = await import("../store/useAuthStore");
      useAuthStore.getState().logout();
    }
    const err = new Error("認証が必要です") as Error & { status?: number };
    err.status = 401;
    throw err;
  }

  return response;
}

export async function getProducts(): Promise<Product[]> {
  const res = await fetch(`${API_URL}/api/products`, {
    method: "GET",
    headers: { Accept: "application/json" },
  });
  const data = await parseJsonSafe<{ products?: Product[] }>(res);
  if (!res.ok) {
    const payload = data as Record<string, unknown>;
    throw { status: res.status, ...payload } as ApiError;
  }
  return Array.isArray(data?.products) ? data.products : [];
}

export async function login(email: string, password: string): Promise<LoginResponse> {
  const res = await fetch(`${API_URL}/api/login`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Accept: "application/json",
    },
    body: JSON.stringify({ email, password }),
  });
  const data = await parseJsonSafe<LoginResponse>(res);
  if (!res.ok) {
    const payload = data as Record<string, unknown>;
    const errorMessage = (payload?.error as string) || `Login failed with status ${res.status}`;
    const error = new Error(errorMessage) as Error & { status: number };
    error.status = res.status;
    throw error;
  }
  return data as LoginResponse;
}

export async function register(
  name: string,
  email: string,
  password: string
): Promise<void> {
  const res = await fetch(`${API_URL}/api/register`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ name, email, password }),
  });

  if (!res.ok) {
    const errorData = await parseJsonSafe<ApiError>(res);
    const errorMessage = (errorData as Record<string, unknown>)?.error || "登録に失敗しました";
    throw new Error(errorMessage as string);
  }
}

export async function createProduct(product: CreateProductRequest): Promise<Product> {
  const response = await fetchWithAuth(`${API_URL}/api/products`, {
    method: "POST",
    body: JSON.stringify(product),
  });

  const data = await parseJsonSafe<Product>(response);
  if (!response.ok) {
    const payload = data as Record<string, unknown>;
    throw { status: response.status, ...payload } as ApiError;
  }
  return data as Product;
}

export async function getProductById(productId: number): Promise<Product> {
  const response = await fetch(`${API_URL}/api/products/${productId}`, {
    method: "GET",
    headers: { Accept: "application/json" },
  });

  const data = await parseJsonSafe<Product>(response);
  if (!response.ok) {
    const payload = data as Record<string, unknown>;
    throw { status: response.status, ...payload } as ApiError;
  }
  return data as Product;
}

export async function updateProduct(productId: number, product: UpdateProductRequest): Promise<Product> {
  const response = await fetchWithAuth(`${API_URL}/api/products/${productId}`, {
    method: "PUT",
    body: JSON.stringify(product),
  });

  const data = await parseJsonSafe<Product>(response);
  if (!response.ok) {
    const payload = data as Record<string, unknown>;
    throw { status: response.status, ...payload } as ApiError;
  }
  return data as Product;
}

export async function deleteProduct(productId: number): Promise<void> {
  const response = await fetchWithAuth(`${API_URL}/api/products/${productId}`, {
    method: "DELETE",
  });

  if (!response.ok) {
    const data = await parseJsonSafe<Record<string, unknown>>(response);
    const payload = data as Record<string, unknown>;
    throw { status: response.status, ...payload } as ApiError;
  }
}

export async function getCategories(): Promise<Category[]> {
  const response = await fetch(`${API_URL}/api/categories`, {
    method: "GET",
    headers: { Accept: "application/json" },
  });

  const data = await parseJsonSafe<CategoriesResponse>(response);
  if (!response.ok) {
    const payload = data as Record<string, unknown>;
    throw { status: response.status, ...payload } as ApiError;
  }

  return Array.isArray((data as CategoriesResponse).categories)
    ? ((data as CategoriesResponse).categories as Category[])
    : [];
}

export async function createCategory(payload: CreateCategoryRequest): Promise<Category> {
  const response = await fetchWithAuth(`${API_URL}/api/categories`, {
    method: "POST",
    body: JSON.stringify(payload),
  });

  const data = await parseJsonSafe<Category>(response);
  if (!response.ok) {
    const errorPayload = data as Record<string, unknown>;
    throw { status: response.status, ...errorPayload } as ApiError;
  }

  return data as Category;
}

export async function updateCategory(categoryId: number, payload: UpdateCategoryRequest): Promise<Category> {
  const response = await fetchWithAuth(`${API_URL}/api/categories/${categoryId}`, {
    method: "PUT",
    body: JSON.stringify(payload),
  });

  const data = await parseJsonSafe<Category>(response);
  if (!response.ok) {
    const errorPayload = data as Record<string, unknown>;
    throw { status: response.status, ...errorPayload } as ApiError;
  }

  return data as Category;
}

export async function deleteCategory(categoryId: number): Promise<void> {
  const response = await fetchWithAuth(`${API_URL}/api/categories/${categoryId}`, {
    method: "DELETE",
  });

  if (!response.ok) {
    const data = await parseJsonSafe<Record<string, unknown>>(response);
    const payload = data as Record<string, unknown>;
    throw { status: response.status, ...payload } as ApiError;
  }
}

export async function setUserRole(userId: number, role: "admin" | "member"): Promise<Record<string, unknown>> {
  const response = await fetchWithAuth(`${API_URL}/api/users/${userId}/role`, {
    method: "PATCH",
    body: JSON.stringify({ role }),
  });

  const data = await parseJsonSafe<Record<string, unknown>>(response);
  if (!response.ok) {
    const payload = data as Record<string, unknown>;
    throw { status: response.status, ...payload } as ApiError;
  }

  return data as Record<string, unknown>;
}

export async function getOrders(status?: "pending" | "cancelled"): Promise<OrderWithItems[]> {
  const query = status ? `?status=${encodeURIComponent(status)}` : "";
  const response = await fetchWithAuth(`${API_URL}/api/orders${query}`, {
    method: "GET",
    headers: { Accept: "application/json" },
  });

  const data = await parseJsonSafe<OrdersResponse>(response);
  if (!response.ok) {
    const payload = data as Record<string, unknown>;
    throw { status: response.status, ...payload } as ApiError;
  }

  const orders = Array.isArray((data as OrdersResponse).orders)
    ? ((data as OrdersResponse).orders as unknown[])
    : [];

  return orders.map((order) => normalizeOrderWithItems(order));
}

export async function createOrder(): Promise<OrderSummary> {
  const response = await fetchWithAuth(`${API_URL}/api/orders`, {
    method: "POST",
    body: JSON.stringify({}),
  });

  const data = await parseJsonSafe<Record<string, unknown>>(response);
  if (!response.ok) {
    const payload = data as Record<string, unknown>;
    throw { status: response.status, ...payload } as ApiError;
  }

  const orderRaw = ((data as Record<string, unknown>).order ?? {}) as Record<string, unknown>;
  return normalizeOrderSummary(orderRaw);
}

export async function cancelOrder(orderId: number): Promise<OrderSummary> {
  const response = await fetchWithAuth(`${API_URL}/api/orders/${orderId}/cancel`, {
    method: "POST",
    body: JSON.stringify({}),
  });

  const data = await parseJsonSafe<Record<string, unknown>>(response);
  if (!response.ok) {
    const payload = data as Record<string, unknown>;
    throw { status: response.status, ...payload } as ApiError;
  }

  const orderRaw = ((data as Record<string, unknown>).order ?? {}) as Record<string, unknown>;
  return normalizeOrderSummary(orderRaw);
}

// ----------------------
// Cart API functions
// ----------------------

export async function getCart(): Promise<CartItem[]> {
  const res = await fetchWithAuth(`${API_URL}/api/cart`, {
    method: "GET",
    headers: { Accept: "application/json" },
  });

  const data = await parseJsonSafe<CartResponse>(res);
  if (!res.ok) {
    const payload = data as Record<string, unknown>;
    throw { status: res.status, ...payload } as ApiError;
  }
  return Array.isArray((data as CartResponse)?.items) ? (data as CartResponse).items : [];
}

export async function addToCart(productId: number, quantity: number): Promise<CartItem> {
  const res = await fetchWithAuth(`${API_URL}/api/cart/items`, {
    method: "POST",
    body: JSON.stringify({ product_id: productId, quantity }),
  });

  const data = await parseJsonSafe<CartItemMutationResponse>(res);
  if (!res.ok) {
    const payload = data as Record<string, unknown>;
    throw { status: res.status, ...payload } as ApiError;
  }
  return (data as CartItemMutationResponse).item;
}

export async function updateCartItem(itemId: number, quantity: number): Promise<CartItem> {
  const res = await fetchWithAuth(`${API_URL}/api/cart/items/${itemId}`, {
    method: "PUT",
    body: JSON.stringify({ quantity }),
  });

  const data = await parseJsonSafe<CartItemMutationResponse>(res);
  if (!res.ok) {
    const payload = data as Record<string, unknown>;
    throw { status: res.status, ...payload } as ApiError;
  }
  return (data as CartItemMutationResponse).item;
}

export async function removeFromCart(itemId: number): Promise<void> {
  const res = await fetchWithAuth(`${API_URL}/api/cart/items/${itemId}`, {
    method: "DELETE",
  });
  if (!res.ok) {
    const data = await parseJsonSafe<Record<string, unknown>>(res);
    const payload = data as Record<string, unknown>;
    throw { status: res.status, ...payload } as ApiError;
  }
}

export async function clearCart(): Promise<void> {
  const res = await fetchWithAuth(`${API_URL}/api/cart`, {
    method: "DELETE",
  });
  if (!res.ok) {
    const data = await parseJsonSafe<Record<string, unknown>>(res);
    const payload = data as Record<string, unknown>;
    throw { status: res.status, ...payload } as ApiError;
  }
}