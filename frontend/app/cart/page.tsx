"use client";
import React, { useEffect } from "react";
import { useRouter } from "next/navigation";
import useCartStore from "../../store/useCartStore";
import { createOrder } from "../../lib/api";

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

  if (loading) return <div style={{ padding: 20 }}>読み込み中...</div>;

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
    <main style={{ padding: 20, maxWidth: 800, margin: "0 auto" }}>
      <h1>カート</h1>
      {error && <div style={{ color: "crimson" }}>{error}</div>}
      {checkoutError && <div style={{ color: "crimson" }}>{checkoutError}</div>}
      {checkoutMessage && (
        <div style={{ color: "green" }}>{checkoutMessage}</div>
      )}

      {items.length === 0 ? (
        <div>カートに商品がありません</div>
      ) : (
        <div>
          <ul style={{ listStyle: "none", padding: 0 }}>
            {items.map((it) => {
              const unitPrice = resolveUnitPrice(it);
              const lineTotal = unitPrice * it.quantity;
              return (
                <li
                  key={it.id}
                  style={{
                    display: "flex",
                    gap: 12,
                    alignItems: "center",
                    padding: 8,
                    borderBottom: "1px solid #eee",
                  }}
                >
                  <div style={{ flex: 1 }}>
                    <div style={{ fontWeight: 600 }}>
                      {it.product_name ?? `Product #${it.product_id}`}
                    </div>
                    <div style={{ color: "#666" }}>単価: ¥{unitPrice}</div>
                    {typeof it.product_stock === "number" && (
                      <div style={{ color: "#666" }}>
                        在庫: {it.product_stock}
                      </div>
                    )}
                  </div>
                  <div
                    style={{ display: "flex", alignItems: "center", gap: 8 }}
                  >
                    <button
                      onClick={() =>
                        void updateItem(it.id, Math.max(1, it.quantity - 1))
                      }
                    >
                      -
                    </button>
                    <input
                      type="number"
                      value={it.quantity}
                      min={1}
                      onChange={(e) =>
                        void updateItem(
                          it.id,
                          Math.max(1, Number(e.target.value) || 1),
                        )
                      }
                      style={{ width: 60 }}
                    />
                    <button
                      onClick={() => void updateItem(it.id, it.quantity + 1)}
                    >
                      +
                    </button>
                  </div>
                  <div style={{ width: 120, textAlign: "right" }}>
                    小計: ¥{lineTotal}
                  </div>
                  <div>
                    <button onClick={() => void removeItem(it.id)}>削除</button>
                  </div>
                </li>
              );
            })}
          </ul>

          <div style={{ marginTop: 16, textAlign: "right" }}>
            <div>合計数量: {totalQuantity}</div>
            <div style={{ fontSize: "1.25rem", fontWeight: 700 }}>
              合計金額: ¥{totalPrice}
            </div>
            <div style={{ marginTop: 12 }}>
              <button
                onClick={() => void clearCart()}
                style={{ marginRight: 8 }}
              >
                カートを空にする
              </button>
              <button
                onClick={() => void handleCheckout()}
                disabled={checkoutLoading}
              >
                {checkoutLoading ? "注文処理中..." : "チェックアウトへ進む"}
              </button>
            </div>
          </div>
        </div>
      )}
    </main>
  );
}
