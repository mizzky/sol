"use client";
import React, { useCallback, useEffect, useState } from "react";
import Link from "next/link";
import { Product, getProducts } from "../lib/api";
import useCartStore from "../store/useCartStore";
import Badge from "./components/ui/Badge";
import Button from "./components/ui/Button";
import Card from "./components/ui/Card";
import { FieldMessage } from "./components/ui/Field";
import QuantityStepper from "./components/ui/QuantityStepper";

export default function Home() {
  const [products, setProducts] = useState<Product[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

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
    void fetchProducts();
  }, [fetchProducts]);

  if (loading) {
    return (
      <main className="mx-auto max-w-6xl px-4 py-10 sm:px-6 lg:px-8">
        読み込み中...
      </main>
    );
  }

  return (
    <main className="mx-auto max-w-6xl px-4 py-10 sm:px-6 lg:px-8">
      <section className="rounded-4xl bg-linear-to-br from-white via-zinc-50 to-indigo-50 p-8 shadow-sm ring-1 ring-zinc-200 sm:p-10">
        <div className="flex flex-col gap-6 lg:flex-row lg:items-end lg:justify-between">
          <div className="max-w-3xl">
            <p className="text-sm uppercase tracking-[0.32em] text-indigo-600">
              Daily Selection
            </p>
            <h1 className="mt-4 text-4xl font-semibold tracking-tight text-zinc-900 sm:text-5xl">
              淡いグレーの空気感に、静かな青を差したコーヒー商品一覧。
            </h1>
            <p className="mt-4 text-sm leading-7 text-zinc-600 sm:text-base">
              商品カード、数量操作、ボタン表現を統一しました。余白を広めに取り、価格と状態がひと目で分かる構成です。
            </p>
          </div>
          <div className="grid gap-3 sm:grid-cols-2">
            <div className="rounded-2xl bg-white px-5 py-4 shadow-sm ring-1 ring-zinc-200">
              <div className="text-sm text-zinc-500">掲載商品数</div>
              <div className="mt-1 text-2xl font-semibold text-zinc-900">
                {products.length}
              </div>
            </div>
            <div className="rounded-2xl bg-white px-5 py-4 shadow-sm ring-1 ring-zinc-200">
              <div className="text-sm text-zinc-500">テーマカラー</div>
              <div className="mt-1 text-2xl font-semibold text-indigo-600">
                Indigo 600
              </div>
            </div>
          </div>
        </div>
      </section>

      <section className="mt-8">
        <div className="mb-4 flex items-center justify-between gap-4">
          <h2 className="text-2xl font-semibold tracking-tight text-zinc-900">
            本日のおすすめ
          </h2>
          <span className="text-sm text-zinc-500">
            操作要素はすべて同じ影と角丸で統一
          </span>
        </div>

        {error && <FieldMessage tone="error">{error}</FieldMessage>}

        {products.length === 0 ? (
          <Card className="mt-4 text-zinc-600">商品がありません</Card>
        ) : (
          <div className="mt-4 grid gap-6 md:grid-cols-2 xl:grid-cols-3">
            {products.map((p) => (
              <ProductCard key={p.id} product={p} />
            ))}
          </div>
        )}
      </section>
    </main>
  );
}

function ProductCard({ product }: { product: Product }) {
  const [qty, setQty] = React.useState<number>(1);
  const [msg, setMsg] = React.useState<string | null>(null);
  const addItem = useCartStore((s) => s.addItem);

  const handleAdd = async () => {
    try {
      await addItem(product.id, qty);
      setMsg("カートに追加しました");
      setTimeout(() => setMsg(null), 1800);
    } catch {
      setMsg("カート追加に失敗しました");
      setTimeout(() => setMsg(null), 2500);
    }
  };

  return (
    <Card className="flex h-full flex-col gap-5 rounded-3xl p-5">
      <div className="aspect-4/3 rounded-2xl bg-linear-to-br from-zinc-100 via-white to-indigo-50" />
      <div className="flex items-start justify-between gap-4">
        <div>
          <h3 className="text-xl font-semibold text-zinc-900">
            {product.name}
          </h3>
          <p className="mt-2 text-sm text-zinc-500">商品ID: {product.id}</p>
        </div>
        <Badge tone={product.is_available ? "success" : "default"}>
          {product.is_available ? "販売中" : "準備中"}
        </Badge>
      </div>

      <div className="text-3xl font-semibold tracking-tight text-indigo-600">
        ¥{product.price}
      </div>

      <div className="flex items-center justify-between gap-4">
        <Link
          href={`/products/${product.id}`}
          className="text-sm font-medium text-indigo-600 hover:text-indigo-500"
        >
          詳細を見る
        </Link>
        <QuantityStepper value={qty} onChange={setQty} />
      </div>

      <div className="mt-auto grid gap-3">
        <Button
          onClick={() => void handleAdd()}
          disabled={!product.is_available}
          className="w-full justify-center"
        >
          カートに追加
        </Button>
        {msg && (
          <FieldMessage tone={msg.includes("失敗") ? "error" : "success"}>
            {msg}
          </FieldMessage>
        )}
      </div>
    </Card>
  );
}
