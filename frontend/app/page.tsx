"use client";
import React, { useCallback, useEffect, useState } from "react";
import { Product, getProducts } from "../lib/api";
import useAuthStore from "../store/useAuthStore";

export default function Home() {
  const [products, setProducts] = useState<Product[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const loadFromStorage = useAuthStore((s) => s.loadFromStorage);

  const fetchProducts = useCallback(async () => {
    setLoading(true);
    try {
      const list = await getProducts();
      setProducts(list);
    } catch (e: unknown) {
      console.error("getProducts error:", e);
      setError("商品一覧の取得に失敗しました");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    loadFromStorage();
    void fetchProducts();
  }, [loadFromStorage, fetchProducts]);

  if (loading) return <div style={{ padding: "2rem" }}>読み込み中...</div>;

  return (
    <main style={{ padding: "2rem", maxWidth: "600px", margin: "0 auto" }}>
      <h1 style={{ fontSize: "2rem", marginBottom: "1rem" }}>
        ☕ Sol Coffee System
      </h1>

      {error && (
        <div style={{ color: "crimson", marginBottom: "1rem" }}>{error}</div>
      )}

      <div
        style={{
          border: "1px solid #ccc",
          borderRadius: "8px",
          padding: "1rem",
        }}
      >
        <h2 style={{ borderBottom: "1px solid #eee", paddingBottom: "0.5rem" }}>
          本日のおすすめ
        </h2>
        <ul style={{ listStyle: "none", padding: 0 }}>
          {products.length === 0 ? (
            <li>商品がありません</li>
          ) : (
            products.map((p) => (
              <div
                key={p.id}
                style={{
                  padding: "1rem",
                  border: "1px solid #ddd",
                  borderRadius: "8px",
                  marginBottom: "0.5rem",
                }}
              >
                <h3 style={{ margin: 0 }}>{p.name}</h3>
                <p style={{ margin: "5px 0", color: "#666" }}>
                  価格: ¥{p.price}
                </p>
                <span
                  style={{
                    fontSize: "0.8rem",
                    padding: "2px 8px",
                    borderRadius: "4px",
                    backgroundColor: p.is_available ? "#e6fffa" : "#fff5f5",
                    color: p.is_available ? "#2c7a7b" : "#c53030",
                  }}
                >
                  {p.is_available ? "販売中" : "準備中"}
                </span>
              </div>
            ))
          )}
        </ul>
      </div>
    </main>
  );
}
