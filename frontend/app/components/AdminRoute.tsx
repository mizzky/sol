"use client";
import React, { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import useAuthStore from "../../store/useAuthStore";

type Props = { children: React.ReactNode };

export default function AdminRoute({ children }: Props) {
  const router = useRouter();
  const { token, user, loadFromStorage, logout } = useAuthStore();
  const [restoring, setRestoring] = useState<boolean>(true);

  useEffect(() => {
    let mounted = true;
    (async () => {
      try {
        await loadFromStorage();
      } catch {
        // ignore - state check below will handle
      } finally {
        // defer to next microtask to avoid sync state update warnings in tests
        Promise.resolve().then(() => {
          if (mounted) setRestoring(false);
        });
      }
    })();
    return () => {
      mounted = false;
    };
  }, [loadFromStorage]);

  useEffect(() => {
    if (restoring) return;

    if (!token) {
      router.push("/login");
      return;
    }

    if (!user) {
      logout();
      router.push("/login");
      return;
    }

    if (user.role !== "admin") {
      router.push("/");
      return;
    }
  }, [restoring, token, user, router, logout]);

  if (restoring) return null;
  if (!token || !user || user.role !== "admin") return null;
  return <>{children}</>;
}
