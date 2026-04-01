"use client";
import React from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import useAuthStore from "../../store/useAuthStore";
import useCartStore from "../../store/useCartStore";
import Badge from "./ui/Badge";
import Button from "./ui/Button";

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

  const navLinkClassName =
    "rounded-xl px-3 py-2 text-sm font-medium text-zinc-700 transition hover:bg-white hover:text-indigo-600";

  return (
    <header className="border-b border-zinc-200/80 bg-zinc-50/95 backdrop-blur">
      <div className="mx-auto flex max-w-6xl flex-col gap-4 px-4 py-4 sm:px-6 lg:flex-row lg:items-center lg:justify-between lg:px-8">
        <div className="flex flex-col gap-3 lg:flex-row lg:items-center">
          <Link
            href="/"
            className="inline-flex items-center rounded-2xl bg-white px-4 py-3 text-sm font-semibold text-zinc-900 shadow-sm ring-1 ring-zinc-200 transition hover:shadow-md"
          >
            Sol Coffee System
          </Link>

          <nav
            className="flex flex-wrap items-center gap-2"
            aria-label="primary"
          >
            <Link href="/" className={navLinkClassName}>
              Home
            </Link>

            {isLoggedIn ? (
              <>
                <Link href="/" className={navLinkClassName}>
                  Products
                </Link>
                <Link href="/orders" className={navLinkClassName}>
                  Orders
                </Link>
                {isAdmin && (
                  <Link href="/admin/products" className={navLinkClassName}>
                    Admin
                  </Link>
                )}
              </>
            ) : null}
          </nav>
        </div>

        <div className="flex flex-wrap items-center gap-3 lg:justify-end">
          {isLoggedIn ? (
            <>
              <Link
                href="/cart"
                aria-label="カート"
                className="inline-flex items-center gap-2 rounded-2xl bg-white px-4 py-3 text-sm font-medium text-zinc-700 shadow-sm ring-1 ring-zinc-200 transition hover:text-indigo-600 hover:shadow-md"
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
                <span>Cart</span>
                {totalQuantity > 0 && (
                  <Badge tone="info">{totalQuantity}</Badge>
                )}
              </Link>

              <div className="inline-flex items-center gap-3 rounded-2xl bg-white px-4 py-3 shadow-sm ring-1 ring-zinc-200 transition hover:shadow-md">
                <div className="flex h-9 w-9 items-center justify-center rounded-full bg-indigo-100 text-sm font-semibold text-indigo-700">
                  {user?.name?.slice(0, 1).toUpperCase() || "U"}
                </div>
                <div className="min-w-0">
                  <div className="truncate text-sm font-semibold text-zinc-900">
                    {user?.name}
                  </div>
                  <div className="text-xs text-zinc-500">{user?.role}</div>
                </div>
              </div>

              <Button onClick={handleLogout} variant="outline">
                Logout
              </Button>
            </>
          ) : (
            <>
              <Link href="/login" className={navLinkClassName}>
                Login
              </Link>
              <Link href="/register" className={navLinkClassName}>
                Sign Up
              </Link>
            </>
          )}
        </div>
      </div>
    </header>
  );
}
