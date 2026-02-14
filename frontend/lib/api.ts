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

export interface LoginResponse {
  message?: string;
  token: string;
  user: {
    id: number;
    name: string;
    email: string;
  };
}

async function parseJsonSafe<T = Record<string, unknown>>(res: Response): Promise<T | Record<string, unknown>> {
  try {
    return await res.json();
  } catch {
    return {};
  }
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
    throw { status: res.status, ...payload } as ApiError;
  }
  return data as LoginResponse;
}

export async function createProduct(product: CreateProductRequest, token?: string): Promise<Product> {
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    Accept: "application/json",
  };
  if (token) headers["Authorization"] = `Bearer ${token}`;

  const res = await fetch(`${API_URL}/api/products`, {
    method: "POST",
    headers,
    body: JSON.stringify(product),
  });
  const data = await parseJsonSafe<Product>(res);
  if (!res.ok) {
    const payload = data as Record<string, unknown>;
    throw { status: res.status, ...payload } as ApiError;
  }
  return data as Product;
}