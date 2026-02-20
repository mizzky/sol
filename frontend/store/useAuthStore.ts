import { create } from "zustand";
import {API_URL, login as apiLogin, register as apiRegister } from "../lib/api";

export interface User {
  id: number;
  name: string;
  email: string;
  role: "admin" | "member";
}

interface AuthState {
  token: string | null;
  user: User | null;
  setToken: (token: string | null) => void;
  setUser: (user: User | null) => void;
  login: (email: string, password: string) => Promise<void>;
  register: (name: string, email: string, password: string) => Promise<void>;
  logout: () => void;
  loadFromStorage: () => void;
}

const STORAGE_KEY = "auth_token";

export const useAuthStore = create<AuthState>((set) => ({
  token: null,
  user: null,
  setToken: (token) => {
    if (typeof window !== "undefined") {
      if (token) localStorage.setItem(STORAGE_KEY, token);
      else localStorage.removeItem(STORAGE_KEY);
    }
    set({ token: token ?? null });
  },
  setUser: (user) => {
    if (typeof window !== "undefined") {
      try {
        if (user) localStorage.setItem("auth_user", JSON.stringify(user));
        else localStorage.removeItem("auth_user");
      } catch {}
    }
    set({ user: user ?? null });
  },
  login: async (email: string, password: string) => {
    try {
      const response = await apiLogin(email, password);
      const token = response.token;
      const user = response.user;
      
      set({ token, user });
      
      if (typeof window !== "undefined") {
        localStorage.setItem(STORAGE_KEY, token);
        localStorage.setItem("auth_user", JSON.stringify(user));
      }
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
    if (typeof window !== "undefined") {
      localStorage.removeItem(STORAGE_KEY);
      localStorage.removeItem("auth_user");
    }
    set({ token: null, user: null });
  },
  loadFromStorage: async () => {
    if (typeof window === "undefined") return;
    const token = localStorage.getItem(STORAGE_KEY);
    if (!token) return;
    set({ token });
    // Try server-side restore via /api/me
    try {
      const res = await fetch(`${API_URL}/api/me`, {
        headers: { Authorization: `Bearer ${token}`, Accept: "application/json" },
      });
      if (res.ok) {
        const payload = await res.json();
        const user = payload.user ?? payload;
        set({ user });
        try {
          localStorage.setItem("auth_user", JSON.stringify(user));
        } catch {}
        return;
      }
    } catch {
      // ignore network errors and fallback to localStorage
    }
    // Fallback: restore user from localStorage if available
    try {
      const raw = localStorage.getItem("auth_user");
      if (raw) {
        const parsed = JSON.parse(raw);
        set({ user: parsed });
      }
    } catch {
      // ignore
    }
  },
}));


export default useAuthStore;