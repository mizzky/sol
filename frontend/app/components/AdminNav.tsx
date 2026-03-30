"use client";

import Link from "next/link";

const linkStyle = {
  padding: "0.5rem 0.85rem",
  border: "1px solid #d6d3d1",
  borderRadius: "999px",
  color: "#44403c",
  textDecoration: "none",
  fontSize: "0.95rem",
} as const;

export default function AdminNav() {
  return (
    <nav
      style={{
        display: "flex",
        gap: "0.75rem",
        flexWrap: "wrap",
        marginBottom: "1.5rem",
      }}
    >
      <Link href="/admin/products" style={linkStyle}>
        商品管理
      </Link>
      <Link href="/admin/categories" style={linkStyle}>
        カテゴリ管理
      </Link>
      <Link href="/admin/users" style={linkStyle}>
        ユーザー権限
      </Link>
    </nav>
  );
}
