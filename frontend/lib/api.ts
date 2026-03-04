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

export interface CartItem {
  id: number;
  cart_id: number;
  product_id: number;
  quantity: number;
  price: number;
  created_at?: string;
  updated_at?: string;
  // サーバーが埋めてくれる場合があるためオプショナルでプロダクト情報を含める
  product?: Product;
}

export interface CartResponse {
  items: CartItem[];
}


async function parseJsonSafe<T = Record<string, unknown>>(res: Response): Promise<T | Record<string, unknown>> {
  try {
    return await res.json();
  } catch {
    return {};
  }
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
    throw new Error("認証が必要です");
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

  const data = await parseJsonSafe<CartItem>(res);
  if (!res.ok) {
    const payload = data as Record<string, unknown>;
    throw { status: res.status, ...payload } as ApiError;
  }
  return data as CartItem;
}

export async function updateCartItem(itemId: number, quantity: number): Promise<CartItem> {
  const res = await fetchWithAuth(`${API_URL}/api/cart/items/${itemId}`, {
    method: "PUT",
    body: JSON.stringify({ quantity }),
  });

  const data = await parseJsonSafe<CartItem>(res);
  if (!res.ok) {
    const payload = data as Record<string, unknown>;
    throw { status: res.status, ...payload } as ApiError;
  }
  return data as CartItem;
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