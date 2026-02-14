import { create } from "zustand";

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
      if (token) {
        localStorage.setItem(STORAGE_KEY, token);
      } else {
        localStorage.removeItem(STORAGE_KEY);
      }
    }
    set({ token: token ?? null });
  },
  setUser: (user) => set({ user: user ?? null }),
  logout: () => {
    if (typeof window !== "undefined") {
      localStorage.removeItem(STORAGE_KEY);
    }
    set({ token: null, user: null });
  },
  loadFromStorage: () => {
    if (typeof window === "undefined") return;
    const token = localStorage.getItem(STORAGE_KEY);
    if (token) set({ token });
  },
}));

export default useAuthStore;