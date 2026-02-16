"use client";
import { useEffect } from "react";
import useAuthStore from "../../store/useAuthStore";

export default function AuthLoader() {
  useEffect(() => {
    const load = useAuthStore.getState().loadFromStorage;
    if (typeof load === "function") {
      void load();
    }
  }, []);
  return null;
}
