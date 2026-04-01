"use client";

import Link from "next/link";

export default function AdminNav() {
  return (
    <nav className="mb-6 flex flex-wrap gap-3">
      <Link
        href="/admin/products"
        className="rounded-xl bg-white px-4 py-2.5 text-sm font-medium text-zinc-700 shadow-sm ring-1 ring-zinc-200 transition hover:text-indigo-600 hover:shadow-md"
      >
        商品管理
      </Link>
      <Link
        href="/admin/categories"
        className="rounded-xl bg-white px-4 py-2.5 text-sm font-medium text-zinc-700 shadow-sm ring-1 ring-zinc-200 transition hover:text-indigo-600 hover:shadow-md"
      >
        カテゴリ管理
      </Link>
      <Link
        href="/admin/users"
        className="rounded-xl bg-white px-4 py-2.5 text-sm font-medium text-zinc-700 shadow-sm ring-1 ring-zinc-200 transition hover:text-indigo-600 hover:shadow-md"
      >
        ユーザー権限
      </Link>
    </nav>
  );
}
