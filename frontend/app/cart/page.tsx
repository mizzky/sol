"use client";
import React, { useEffect } from "react";
import { useRouter } from "next/navigation";
import useCartStore from "../../store/useCartStore";
import { createOrder } from "../../lib/api";
import Button from "../components/ui/Button";
import Card from "../components/ui/Card";
import { FieldMessage } from "../components/ui/Field";
import QuantityStepper from "../components/ui/QuantityStepper";

export default function CartPage() {
  const router = useRouter();
  const {
    items,
    totalPrice,
    totalQuantity,
    loading,
    error,
    updateItem,
    removeItem,
    clearCart,
    syncCart,
  } = useCartStore();
  const [checkoutLoading, setCheckoutLoading] = React.useState(false);
  const [checkoutMessage, setCheckoutMessage] = React.useState<string | null>(
    null,
  );
  const [checkoutError, setCheckoutError] = React.useState<string | null>(null);

  useEffect(() => {
    void syncCart();
  }, [syncCart]);

  const handleCheckout = async () => {
    setCheckoutError(null);
    setCheckoutMessage(null);
    setCheckoutLoading(true);
    try {
      await createOrder();
      await syncCart();
      setCheckoutMessage("注文を作成しました。注文履歴を確認してください。");
      router.push("/orders");
    } catch (err: unknown) {
      const status = (err as { status?: number } | null)?.status;
      if (status === 400) {
        setCheckoutError("カートが空です");
      } else if (status === 401) {
        setCheckoutError("認証が必要です");
      } else if (status === 409) {
        setCheckoutError("在庫不足のため注文を作成できません");
      } else {
        setCheckoutError("注文作成に失敗しました");
      }
    } finally {
      setCheckoutLoading(false);
    }
  };

  if (loading) {
    return (
      <main className="mx-auto max-w-6xl px-4 py-10 sm:px-6 lg:px-8">
        読み込み中...
      </main>
    );
  }

  const resolveUnitPrice = (item: (typeof items)[number]) => {
    if (typeof item.price === "number" && Number.isFinite(item.price)) {
      return item.price;
    }
    if (
      typeof item.product_price === "number" &&
      Number.isFinite(item.product_price)
    ) {
      return item.product_price;
    }
    return 0;
  };

  return (
    <main className="mx-auto max-w-6xl px-4 py-10 sm:px-6 lg:px-8">
      <div className="mb-8 flex flex-col gap-3">
        <p className="text-sm uppercase tracking-[0.28em] text-indigo-600">
          Cart
        </p>
        <h1 className="text-4xl font-semibold tracking-tight text-zinc-900">
          数量操作を縦ステッパーへ統一したカート画面。
        </h1>
        <p className="max-w-3xl text-sm leading-7 text-zinc-600 sm:text-base">
          商品ごとの編集、削除、チェックアウトの導線を整理し、入力欄の代わりに同じ操作感のステッパーへ揃えています。
        </p>
      </div>

      <div className="grid gap-4">
        {error && <FieldMessage tone="error">{error}</FieldMessage>}
        {checkoutError && (
          <FieldMessage tone="error">{checkoutError}</FieldMessage>
        )}
        {checkoutMessage && (
          <FieldMessage tone="success">{checkoutMessage}</FieldMessage>
        )}
      </div>

      {items.length === 0 ? (
        <Card className="mt-6 text-zinc-600">カートに商品がありません</Card>
      ) : (
        <div className="mt-6 grid gap-6 lg:grid-cols-[1.2fr_0.8fr]">
          <div className="grid gap-4">
            {items.map((it) => {
              const unitPrice = resolveUnitPrice(it);
              const lineTotal = unitPrice * it.quantity;
              return (
                <Card
                  key={it.id}
                  className="grid gap-5 rounded-3xl p-5 md:grid-cols-[minmax(0,1fr)_auto_auto] md:items-center"
                >
                  <div>
                    <div className="text-lg font-semibold text-zinc-900">
                      {it.product_name ?? `Product #${it.product_id}`}
                    </div>
                    <div className="mt-2 text-sm text-zinc-600">
                      単価: ¥{unitPrice}
                    </div>
                    {typeof it.product_stock === "number" && (
                      <div className="mt-1 text-sm text-zinc-500">
                        在庫: {it.product_stock}
                      </div>
                    )}
                  </div>
                  <div className="flex justify-start md:justify-center">
                    <QuantityStepper
                      value={it.quantity}
                      onChange={(next) => void updateItem(it.id, next)}
                    />
                  </div>
                  <div className="flex items-center justify-between gap-4 md:flex-col md:items-end">
                    <div className="text-right">
                      <div className="text-xs uppercase tracking-[0.24em] text-zinc-500">
                        Line Total
                      </div>
                      <div className="mt-1 text-xl font-semibold text-indigo-600">
                        ¥{lineTotal}
                      </div>
                      <div className="mt-1 text-sm text-zinc-500">
                        小計: ¥{lineTotal}
                      </div>
                    </div>
                    <Button
                      variant="danger"
                      onClick={() => void removeItem(it.id)}
                    >
                      削除
                    </Button>
                  </div>
                </Card>
              );
            })}
          </div>

          <Card className="h-fit rounded-3xl p-6">
            <div className="text-sm uppercase tracking-[0.28em] text-zinc-500">
              Summary
            </div>
            <div className="mt-6 grid gap-4 rounded-2xl bg-zinc-50 p-5 ring-1 ring-zinc-200">
              <div className="flex items-center justify-between text-sm text-zinc-600">
                <span>合計数量</span>
                <span className="font-semibold text-zinc-900">
                  {totalQuantity}
                </span>
              </div>
              <div className="text-sm text-zinc-500">
                合計数量: {totalQuantity}
              </div>
              <div className="flex items-center justify-between text-sm text-zinc-600">
                <span>合計金額</span>
                <span className="text-2xl font-semibold text-indigo-600">
                  ¥{totalPrice}
                </span>
              </div>
              <div className="text-sm text-zinc-500">
                合計金額: ¥{totalPrice}
              </div>
            </div>
            <div className="mt-5 grid gap-3">
              <Button variant="secondary" onClick={() => void clearCart()}>
                カートを空にする
              </Button>
              <Button
                onClick={() => void handleCheckout()}
                disabled={checkoutLoading}
              >
                {checkoutLoading ? "注文処理中..." : "チェックアウトへ進む"}
              </Button>
            </div>
          </Card>
        </div>
      )}
    </main>
  );
}
