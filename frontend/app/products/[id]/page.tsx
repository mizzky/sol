"use client";

import Link from "next/link";
import React, { useCallback, useEffect, useMemo, useState } from "react";
import { useParams } from "next/navigation";
import {
  getCategories,
  getProductById,
  type Category,
  type Product,
} from "../../../lib/api";
import useCartStore from "../../../store/useCartStore";

function getErrorMessage(error: unknown): string {
  const status = (error as { status?: number } | null)?.status;
  if (status === 404) return "商品が見つかりません";
  return "商品詳細の取得に失敗しました";
}

export default function ProductDetailPage() {
  const params = useParams<{ id: string }>();
  const [product, setProduct] = useState<Product | null>(null);
  const [categories, setCategories] = useState<Category[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [quantity, setQuantity] = useState(1);
  const [message, setMessage] = useState<string | null>(null);
  const addItem = useCartStore((state) => state.addItem);

  const productId = Number(params?.id ?? 0);

  const loadProduct = useCallback(async () => {
    if (!Number.isInteger(productId) || productId <= 0) {
      setError("商品IDが不正です");
      setLoading(false);
      return;
    }

    setLoading(true);
    setError(null);

    try {
      const [productResponse, categoryResponse] = await Promise.all([
        getProductById(productId),
        getCategories(),
      ]);
      setProduct(productResponse);
      setCategories(categoryResponse);
    } catch (err: unknown) {
      setError(getErrorMessage(err));
    } finally {
      setLoading(false);
    }
  }, [productId]);

  useEffect(() => {
    void loadProduct();
  }, [loadProduct]);

  const categoryName = useMemo(() => {
    if (!product) {
      return "-";
    }
    return (
      categories.find((category) => category.id === product.category_id)
        ?.name ?? `カテゴリ #${product.category_id}`
    );
  }, [categories, product]);

  const handleAddToCart = async () => {
    if (!product) {
      return;
    }

    try {
      await addItem(product.id, quantity);
      setMessage("カートに追加しました");
      setTimeout(() => setMessage(null), 1800);
    } catch {
      setMessage("カート追加に失敗しました");
      setTimeout(() => setMessage(null), 2200);
    }
  };

  if (loading) {
    return <main style={{ padding: "2rem" }}>読み込み中...</main>;
  }

  if (error) {
    return (
      <main style={{ padding: "2rem", maxWidth: "760px", margin: "0 auto" }}>
        <div style={{ color: "crimson", marginBottom: "1rem" }}>{error}</div>
        <Link href="/" style={{ color: "#0f766e" }}>
          商品一覧へ戻る
        </Link>
      </main>
    );
  }

  if (!product) {
    return null;
  }

  return (
    <main style={{ padding: "2rem", maxWidth: "760px", margin: "0 auto" }}>
      <Link href="/" style={{ color: "#0f766e" }}>
        商品一覧へ戻る
      </Link>
      <article
        style={{
          marginTop: "1rem",
          border: "1px solid #d6d3d1",
          borderRadius: 16,
          padding: "1.5rem",
          background: "#fffdf8",
        }}
      >
        <div
          style={{
            display: "flex",
            justifyContent: "space-between",
            gap: "1rem",
            flexWrap: "wrap",
          }}
        >
          <div>
            <h1 style={{ margin: "0 0 0.5rem" }}>{product.name}</h1>
            <div style={{ fontSize: "1.2rem", fontWeight: 700 }}>
              ¥{product.price}
            </div>
          </div>
          <span
            style={{
              padding: "0.35rem 0.75rem",
              borderRadius: 999,
              background: product.is_available ? "#dcfce7" : "#fee2e2",
              color: product.is_available ? "#166534" : "#991b1b",
              height: "fit-content",
            }}
          >
            {product.is_available ? "販売中" : "販売停止中"}
          </span>
        </div>

        <dl
          style={{
            display: "grid",
            gridTemplateColumns: "140px 1fr",
            gap: "0.75rem",
            marginTop: "1.5rem",
          }}
        >
          <dt>カテゴリ</dt>
          <dd style={{ margin: 0 }}>{categoryName}</dd>
          <dt>SKU</dt>
          <dd style={{ margin: 0 }}>{product.sku}</dd>
          <dt>在庫数</dt>
          <dd style={{ margin: 0 }}>{product.stock_quantity}</dd>
          <dt>説明</dt>
          <dd style={{ margin: 0 }}>
            {product.description || "説明はありません"}
          </dd>
        </dl>

        <div
          style={{
            marginTop: "1.5rem",
            display: "flex",
            gap: "0.75rem",
            alignItems: "center",
            flexWrap: "wrap",
          }}
        >
          <input
            type="number"
            min={1}
            value={quantity}
            onChange={(e) =>
              setQuantity(Math.max(1, Number(e.target.value) || 1))
            }
            style={{ width: 80 }}
          />
          <button
            type="button"
            onClick={() => void handleAddToCart()}
            disabled={!product.is_available}
          >
            カートに追加
          </button>
          {message && <span>{message}</span>}
        </div>
      </article>
    </main>
  );
}
