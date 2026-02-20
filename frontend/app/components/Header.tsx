"use client";
import React from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import useAuthStore from "../../store/useAuthStore";

export default function Header() {
  const router = useRouter();
  const { token, user, logout } = useAuthStore();
  const isLoggedIn = !!token && !!user;
  const isAdmin = user?.role === "admin";

  const handleLogout = () => {
    logout();
    router.push("/");
  };

  return (
    <header style={{ padding: "1rem", borderBottom: "1px solid #ccc" }}>
      <nav style={{ display: "flex", gap: "1rem", alignItems: "center" }}>
        {/* Home リンク（常に表示） */}
        <Link href="/">Home</Link>

        {isLoggedIn ? (
          <>
            {/* ログイン時のメニュー */}
            <Link href="/">Products</Link>
            {isAdmin && <Link href="/admin/products">Admin</Link>}
            <span>{user.name}</span>
            <button onClick={handleLogout}>Logout</button>
          </>
        ) : (
          <>
            {/* 未ログイン時のメニュー */}
            <Link href="/login">Login</Link>
            <Link href="/register">Sign Up</Link>
          </>
        )}
      </nav>
    </header>
  );
}
