import { create } from "zustand";
import type { CartItem } from "../lib/api";
import {
  getCart as apiGetCart,
  addToCart as apiAddToCart,
  updateCartItem as apiUpdateCartItem,
  removeFromCart as apiRemoveFromCart,
  clearCart as apiClearCart,
} from "../lib/api";

type Nullable<T> = T | null;

interface CartState {
  items: CartItem[];
  totalQuantity: number;
  totalPrice: number;
  loading: boolean;
  error: Nullable<string>;
  setCart: (items: CartItem[]) => void;
  syncCart: () => Promise<void>;
  addItem: (productId: number, quantity?: number) => Promise<void>;
  updateItem: (itemId: number, quantity: number) => Promise<void>;
  removeItem: (itemId: number) => Promise<void>;
  clearCart: () => Promise<void>;
}

function computeTotals(items: CartItem[]) {
  const totalQuantity = items.reduce((s, it) => s + (it.quantity || 0), 0);
  const totalPrice = items.reduce((s, it) => s + (it.quantity || 0) * (it.price || 0), 0);
  return { totalQuantity, totalPrice };
}

export const useCartStore = create<CartState>((set, get) => ({
  items: [],
  totalQuantity: 0,
  totalPrice: 0,
  loading: false,
  error: null,

  setCart(items: CartItem[]) {
    const { totalQuantity, totalPrice } = computeTotals(items);
    set(() => ({ items, totalQuantity, totalPrice, error: null }));
  },

  async syncCart() {
    set({ loading: true, error: null });
    try {
      const items = await apiGetCart();
      const { totalQuantity, totalPrice } = computeTotals(items);
      set(() => ({ items, totalQuantity, totalPrice }));
    } catch (err: any) {
      set({ error: err?.message || String(err) });
    } finally {
      set({ loading: false });
    }
  },

  async addItem(productId: number, quantity = 1) {
    set({ loading: true, error: null });
    try {
      const newItem = await apiAddToCart(productId, quantity);
      const items = get().items.slice();
      // Replace if same item id exists, otherwise push
      const idx = items.findIndex((i) => i.id === newItem.id);
      if (idx >= 0) items[idx] = newItem;
      else items.push(newItem);
      const { totalQuantity, totalPrice } = computeTotals(items);
      set(() => ({ items, totalQuantity, totalPrice }));
    } catch (err: any) {
      set({ error: err?.message || String(err) });
      throw err;
    } finally {
      set({ loading: false });
    }
  },

  async updateItem(itemId: number, quantity: number) {
    set({ loading: true, error: null });
    try {
      const updated = await apiUpdateCartItem(itemId, quantity);
      const items = get().items.slice();
      const idx = items.findIndex((i) => i.id === updated.id);
      if (idx >= 0) items[idx] = updated;
      const { totalQuantity, totalPrice } = computeTotals(items);
      set(() => ({ items, totalQuantity, totalPrice }));
    } catch (err: any) {
      set({ error: err?.message || String(err) });
      throw err;
    } finally {
      set({ loading: false });
    }
  },

  async removeItem(itemId: number) {
    set({ loading: true, error: null });
    try {
      await apiRemoveFromCart(itemId);
      const items = get().items.filter((i) => i.id !== itemId);
      const { totalQuantity, totalPrice } = computeTotals(items);
      set(() => ({ items, totalQuantity, totalPrice }));
    } catch (err: any) {
      set({ error: err?.message || String(err) });
      throw err;
    } finally {
      set({ loading: false });
    }
  },

  async clearCart() {
    set({ loading: true, error: null });
    try {
      await apiClearCart();
      set(() => ({ items: [], totalQuantity: 0, totalPrice: 0 }));
    } catch (err: any) {
      set({ error: err?.message || String(err) });
      throw err;
    } finally {
      set({ loading: false });
    }
  },
}));

export default useCartStore;
