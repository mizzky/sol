import { create } from "zustand";
import {
  API_URL,
  fetchWithAuth,
  login as apiLogin,
  register as apiRegister,
  revokeRefreshToken,
} from "../lib/api";
import { useCartStore } from "./useCartStore";

export interface User {
  id: number;
  name: string;
  email: string;
  role: "admin" | "member";
}

interface AuthState {
  isAuthenticated: boolean;
  user: User | null;
  setUser: (user: User | null) => void;
  login: (email: string, password: string) => Promise<void>;
  register: (name: string, email: string, password: string) => Promise<void>;
  logout: () => void;
  loadFromStorage: () => Promise<void>;
}

export const useAuthStore = create<AuthState>((set) => ({
  isAuthenticated: false,
  user: null,
  setUser: (user) => {
    set({ user: user ?? null, isAuthenticated: Boolean(user) });
  },
  login: async (email: string, password: string) => {
    try {
      const response = await apiLogin(email, password);
      const user = response.user;

      set({ isAuthenticated: true, user });
    } catch (error) {
      // エラーをそのまま再throw（呼び出し側でハンドリング）
      throw error;
    }
  },
  register: async (name: string, email: string, password: string) => {
    try {
      await apiRegister(name, email, password);
      // 登録後は自動ログインしない設計とする
    } catch (error) {
      throw error;
    }
  },
  logout: () => {
    useCartStore.getState().resetCart();
    set({ isAuthenticated: false, user: null });
    void revokeRefreshToken().catch(() => {
      // クライアント状態クリアを優先する
    });
  },
  loadFromStorage: async () => {
    try {
      const res = await fetchWithAuth(`${API_URL}/api/me`, {
        method: "GET",
        headers: { Accept: "application/json" },
      });

      const payload = await res.json();
      const user = payload.user ?? payload;
      set({ isAuthenticated: true, user });
    } catch {
      set({ isAuthenticated: false, user: null });
    }
  },
}));


export default useAuthStore;