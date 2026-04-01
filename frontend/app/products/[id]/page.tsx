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
import Badge from "../../components/ui/Badge";
import Button from "../../components/ui/Button";
import Card from "../../components/ui/Card";
import { FieldMessage } from "../../components/ui/Field";
import QuantityStepper from "../../components/ui/QuantityStepper";

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
    return (
      <main className="mx-auto max-w-5xl px-4 py-10 sm:px-6 lg:px-8">
        読み込み中...
      </main>
    );
  }

  if (error) {
    return (
      <main className="mx-auto max-w-5xl px-4 py-10 sm:px-6 lg:px-8">
        <FieldMessage tone="error">{error}</FieldMessage>
        <Link
          href="/"
          className="mt-4 inline-flex text-sm font-medium text-indigo-600 hover:text-indigo-500"
        >
          商品一覧へ戻る
        </Link>
      </main>
    );
  }

  if (!product) {
    return null;
  }

  return (
    <main className="mx-auto max-w-5xl px-4 py-10 sm:px-6 lg:px-8">
      <Link
        href="/"
        className="inline-flex text-sm font-medium text-indigo-600 hover:text-indigo-500"
      >
        商品一覧へ戻る
      </Link>
      <Card className="mt-4 rounded-4xl p-6 sm:p-8">
        <div className="grid gap-8 lg:grid-cols-[1.05fr_0.95fr]">
          <div>
            <div className="aspect-4/3 rounded-3xl bg-linear-to-br from-zinc-100 via-white to-indigo-50" />
            <div className="mt-6 flex items-start justify-between gap-4">
              <div>
                <h1 className="text-4xl font-semibold tracking-tight text-zinc-900">
                  {product.name}
                </h1>
                <div className="mt-3 text-3xl font-semibold text-indigo-600">
                  ¥{product.price}
                </div>
              </div>
              <Badge tone={product.is_available ? "success" : "danger"}>
                {product.is_available ? "販売中" : "販売停止中"}
              </Badge>
            </div>
          </div>

          <div className="flex flex-col gap-6">
            <div className="grid gap-4 rounded-3xl bg-zinc-50 p-6 ring-1 ring-zinc-200 sm:grid-cols-[140px_1fr]">
              <div className="text-sm font-medium text-zinc-500">カテゴリ</div>
              <div className="text-sm text-zinc-900">{categoryName}</div>
              <div className="text-sm font-medium text-zinc-500">SKU</div>
              <div className="text-sm text-zinc-900">{product.sku}</div>
              <div className="text-sm font-medium text-zinc-500">在庫数</div>
              <div className="text-sm text-zinc-900">
                {product.stock_quantity}
              </div>
              <div className="text-sm font-medium text-zinc-500">説明</div>
              <div className="text-sm leading-7 text-zinc-700">
                {product.description || "説明はありません"}
              </div>
            </div>

            <div className="flex flex-wrap items-center gap-4">
              <input
                type="number"
                value={quantity}
                onChange={(e) =>
                  setQuantity(Math.max(1, Number(e.target.value) || 1))
                }
                min={1}
                className="sr-only"
                aria-hidden="true"
                tabIndex={-1}
              />
              <QuantityStepper value={quantity} onChange={setQuantity} />
              <Button
                type="button"
                onClick={() => void handleAddToCart()}
                disabled={!product.is_available}
              >
                カートに追加
              </Button>
            </div>

            {message && (
              <FieldMessage
                tone={message.includes("失敗") ? "error" : "success"}
              >
                {message}
              </FieldMessage>
            )}
          </div>
        </div>
      </Card>
    </main>
  );
}
