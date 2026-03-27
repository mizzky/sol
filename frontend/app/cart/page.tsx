"use client";
import React, { useEffect } from "react";
import useCartStore from "../../store/useCartStore";

export default function CartPage() {
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

  useEffect(() => {
    void syncCart();
  }, [syncCart]);

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
              <button>チェックアウトへ進む</button>
            </div>
          </div>
        </div>
      )}
    </main>
  );
}
