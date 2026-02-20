"use client";
import React, { useState } from "react";
import AdminRoute from "../../../app/components/AdminRoute";
import { CreateProductRequest, createProduct } from "../../../lib/api";

export default function AdminProductsPage() {
  const [newName, setNewName] = useState("");
  const [newPrice, setNewPrice] = useState("");
  const [newCategoryId, setNewCategoryId] = useState("1");
  const [newSku, setNewSku] = useState("");
  const [newStock, setNewStock] = useState("0");
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  const handleAddProduct = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setSuccess(null);

    const payload: CreateProductRequest = {
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
      await createProduct(payload);
      setNewName("");
      setNewPrice("");
      setNewCategoryId("1");
      setNewSku("");
      setNewStock("0");
      setSuccess("商品を追加しました");
    } catch (err: unknown) {
      console.error("createProduct error:", err);
      const status = (err as any)?.status;
      if (status === 401) setError("認証が必要です（ログインしてください）");
      else if (status === 403) setError("管理者権限が必要です");
      else if (status === 404) setError("カテゴリが見つかりません");
      else if (status === 409) setError("SKUが既に存在します");
      else setError("商品追加に失敗しました");
    }
  };

  return (
    <AdminRoute>
      <main style={{ padding: "2rem", maxWidth: "600px", margin: "0 auto" }}>
        <h1>商品管理（管理者）</h1>

        {error && <div style={{ color: "crimson" }}>{error}</div>}
        {success && <div style={{ color: "green" }}>{success}</div>}

        <form onSubmit={handleAddProduct} style={{ marginTop: "1rem" }}>
          <h3>新規商品登録（管理者）</h3>
          <input
            type="text"
            placeholder="コーヒー名"
            value={newName}
            onChange={(e) => setNewName(e.target.value)}
            required
          />
          <input
            type="number"
            placeholder="価格"
            value={newPrice}
            onChange={(e) => setNewPrice(e.target.value)}
            required
          />
          <input
            type="text"
            placeholder="SKU（未入力時は自動生成）"
            value={newSku}
            onChange={(e) => setNewSku(e.target.value)}
          />
          <input
            type="number"
            placeholder="カテゴリID"
            value={newCategoryId}
            onChange={(e) => setNewCategoryId(e.target.value)}
            required
          />
          <input
            type="number"
            placeholder="在庫数"
            value={newStock}
            onChange={(e) => setNewStock(e.target.value)}
          />
          <button type="submit">追加</button>
        </form>
      </main>
    </AdminRoute>
  );
}
