"use client";
import React, { useEffect } from "react";
import { useRouter } from "next/navigation";
import useAuthStore from "../../store/useAuthStore";

type Props = { children: React.ReactNode };

export default function AdminRoute({ children }: Props) {
  const router = useRouter();
  const { token, user } = useAuthStore();

  useEffect(() => {
    if (!token || !user) {
      router.push("/login");
      return;
    }
    if (user.role !== "admin") {
      router.push("/");
      return;
    }
  }, [token, user, router]);

  // 管理者であれば子コンテンツを表示（redirect が非同期なので短時間は既にレンダリングされるが、テストは push 呼び出しを見ている）
  if (!token || !user || user.role !== "admin") {
    return null;
  }
  return <>{children}</>;
}
