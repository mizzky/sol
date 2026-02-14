"use client";
import React, { useCallback, useEffect, useState } from "react";
import { Product, getProducts, createProduct } from "../lib/api";
import useAuthStore from "../store/useAuthStore";

// Productsの型定義は ../lib/api の Product を利用

export default function Home() {
  const [products, setProducts] = useState<Product[]>([]);
  const [loading, setLoading] = useState(true);

  const [newName, setNewName] = useState("");
  const [newPrice, setNewPrice] = useState("");
  const [newCategoryId, setNewCategoryId] = useState("1");
  const [newSku, setNewSku] = useState("");
  const [newStock, setNewStock] = useState("0");
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  const token = useAuthStore((s) => s.token);
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

  const handleAddProduct = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setSuccess(null);

    const payload = {
      name: newName,
      price: Number(newPrice || 0),
      is_available: true,
      category_id: Number(newCategoryId || 1),
      sku: newSku || `SKU-${Date.now()}`,
      description: null,
      image_url: null,
      stock_quantity: Number(newStock || 0),
    };

    try {
      await createProduct(payload, token ?? undefined);
      setNewName("");
      setNewPrice("");
      setNewCategoryId("1");
      setNewSku("");
      setNewStock("0");
      setSuccess("商品を追加しました");
      await fetchProducts();
    } catch (err: unknown) {
      console.error("createProduct error:", err);
      const extractStatus = (e: unknown): number | undefined => {
        if (typeof e !== "object" || e === null) return undefined;
        const maybe = e as { status?: unknown };
        return typeof maybe.status === "number" ? maybe.status : undefined;
      };
      const status = extractStatus(err);
      if (status === 401) setError("認証が必要です（ログインしてください）");
      else if (status === 403) setError("管理者権限が必要です");
      else if (status === 404) setError("カテゴリが見つかりません");
      else if (status === 409) setError("SKUが既に存在します");
      else setError("商品追加に失敗しました");
    }
  };

  if (loading) return <div style={{ padding: "2rem" }}>読み込み中...</div>;

  return (
    <main style={{ padding: "2rem", maxWidth: "600px", margin: "0 auto" }}>
      <h1 style={{ fontSize: "2rem", marginBottom: "1rem" }}>
        ☕ Sol Coffee System
      </h1>

      {error && (
        <div style={{ color: "crimson", marginBottom: "1rem" }}>{error}</div>
      )}
      {success && (
        <div style={{ color: "green", marginBottom: "1rem" }}>{success}</div>
      )}

      {/* 登録フォーム */}
      <form
        onSubmit={handleAddProduct}
        style={{
          marginBottom: "2rem",
          padding: "1rem",
          background: "#f9f9f9",
          borderRadius: "8px",
        }}
      >
        <h3>新規商品登録（管理者）</h3>
        <input
          type="text"
          placeholder="コーヒー名"
          value={newName}
          onChange={(e) => setNewName(e.target.value)}
          style={{ marginRight: "10px", padding: "5px" }}
          required
        />
        <input
          type="number"
          placeholder="価格"
          value={newPrice}
          onChange={(e) => setNewPrice(e.target.value)}
          style={{ marginRight: "10px", padding: "5px" }}
          required
        />
        <input
          type="text"
          placeholder="SKU（未入力時は自動生成）"
          value={newSku}
          onChange={(e) => setNewSku(e.target.value)}
          style={{ marginRight: "10px", padding: "5px" }}
        />
        <input
          type="number"
          placeholder="カテゴリID"
          value={newCategoryId}
          onChange={(e) => setNewCategoryId(e.target.value)}
          style={{ marginRight: "10px", padding: "5px", width: "90px" }}
          required
        />
        <input
          type="number"
          placeholder="在庫数"
          value={newStock}
          onChange={(e) => setNewStock(e.target.value)}
          style={{ marginRight: "10px", padding: "5px", width: "90px" }}
        />
        <button
          type="submit"
          style={{ padding: "5px 15px", cursor: "pointer" }}
        >
          追加
        </button>
      </form>

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
