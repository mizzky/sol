"use client";
import React from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import useAuthStore from "../../store/useAuthStore";
import useCartStore from "../../store/useCartStore";

export default function Header() {
  const router = useRouter();
  const { token, user, logout } = useAuthStore();
  const isLoggedIn = !!token && !!user;
  const isAdmin = user?.role === "admin";
  const totalQuantity = useCartStore((s) => s.totalQuantity);

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
            <Link href="/orders">Orders</Link>
            {isAdmin && <Link href="/admin/products">Admin</Link>}
            <span>{user.name}</span>
            <button onClick={handleLogout}>Logout</button>
            <Link
              href="/cart"
              style={{
                marginLeft: "auto",
                display: "flex",
                alignItems: "center",
              }}
            >
              <svg
                width="20"
                height="20"
                viewBox="0 0 24 24"
                fill="none"
                xmlns="http://www.w3.org/2000/svg"
              >
                <path
                  d="M6 6h15l-1.5 9h-12L6 6z"
                  stroke="currentColor"
                  strokeWidth="1.2"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                />
                <circle cx="10" cy="20" r="1" fill="currentColor" />
                <circle cx="18" cy="20" r="1" fill="currentColor" />
              </svg>
              {totalQuantity > 0 && (
                <span
                  style={{
                    marginLeft: 6,
                    background: "#e53e3e",
                    color: "white",
                    borderRadius: 12,
                    padding: "2px 6px",
                    fontSize: "0.8rem",
                  }}
                >
                  {totalQuantity}
                </span>
              )}
            </Link>
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
