import { create } from "zustand";
import {API_URL } from "../lib/api";

export interface User {
  id: number;
  name: string;
  email: string;
}

interface AuthState {
  token: string | null;
  user: User | null;
  setToken: (token: string | null) => void;
  setUser: (user: User | null) => void;
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