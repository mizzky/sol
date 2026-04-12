"use client";
import { useEffect } from "react";
import useAuthStore from "../../store/useAuthStore";
import useCartStore from "../../store/useCartStore";

export default function AuthLoader() {
  useEffect(() => {
    let active = true;

    const bootstrap = async () => {
      const authState = useAuthStore.getState();

      if (typeof authState.loadFromStorage !== "function") {
        return;
      }

      try {
        await authState.loadFromStorage();
        if (!active) {
          return;
        }

        const cartState = useCartStore.getState();
        if (useAuthStore.getState().isAuthenticated) {
          await cartState.syncCart();
          return;
        }

        cartState.resetCart();
      } catch {
        // Keep client state consistent when restore/sync fails unexpectedly.
        if (!active) {
          return;
        }
        useCartStore.getState().resetCart();
      }
    };

    void bootstrap();

    return () => {
      active = false;
    };
  }, []);
  return null;
}
