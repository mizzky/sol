"use client";

import React, { useCallback, useEffect, useState } from "react";
import { cancelOrder, getOrders, type OrderWithItems } from "../../lib/api";

export default function OrdersPage() {
  const [orders, setOrders] = useState<OrderWithItems[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [cancellingId, setCancellingId] = useState<number | null>(null);

  const loadOrders = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const fetched = await getOrders();
      setOrders(fetched);
    } catch (err: unknown) {
      const status = (err as { status?: number } | null)?.status;
      if (status === 401) {
        setError("認証が必要です");
      } else {
        setError("注文履歴の取得に失敗しました");
      }
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void loadOrders();
  }, [loadOrders]);

  const handleCancel = async (orderId: number) => {
    setCancellingId(orderId);
    setError(null);
    try {
      await cancelOrder(orderId);
      await loadOrders();
    } catch (err: unknown) {
      const status = (err as { status?: number } | null)?.status;
      if (status === 400) {
        setError("この注文はキャンセルできません");
      } else if (status === 404) {
        setError("注文が見つかりません");
      } else {
        setError("注文キャンセルに失敗しました");
      }
    } finally {
      setCancellingId(null);
    }
  };

  if (loading) {
    return <main style={{ padding: 20 }}>読み込み中...</main>;
  }

  return (
    <main style={{ padding: 20, maxWidth: 900, margin: "0 auto" }}>
      <h1>注文履歴</h1>
      {error && (
        <div style={{ color: "crimson", marginBottom: 12 }}>{error}</div>
      )}

      {orders.length === 0 ? (
        <div>注文履歴はありません</div>
      ) : (
        <div style={{ display: "grid", gap: 12 }}>
          {orders.map(({ order, items }) => (
            <section
              key={order.id}
              style={{ border: "1px solid #ddd", borderRadius: 8, padding: 12 }}
            >
              <div style={{ display: "flex", justifyContent: "space-between" }}>
                <strong>注文ID: {order.id}</strong>
                <span>ステータス: {order.status}</span>
              </div>
              <div style={{ marginTop: 6 }}>合計: ¥{order.total}</div>
              <div style={{ marginTop: 6 }}>
                作成日時: {order.created_at || "-"}
              </div>

              <div style={{ marginTop: 10 }}>
                <strong>明細</strong>
                <ul style={{ margin: "6px 0 0", paddingLeft: 18 }}>
                  {items.map((item, idx) => (
                    <li key={`${order.id}-${item.product_id}-${idx}`}>
                      {item.product_name_snapshot || `商品 #${item.product_id}`}{" "}
                      / 数量: {item.quantity} / 単価: ¥{item.unit_price}
                    </li>
                  ))}
                </ul>
              </div>

              {order.status === "pending" && (
                <div style={{ marginTop: 12 }}>
                  <button
                    onClick={() => void handleCancel(order.id)}
                    disabled={cancellingId === order.id}
                  >
                    {cancellingId === order.id
                      ? "キャンセル中..."
                      : "注文をキャンセル"}
                  </button>
                </div>
              )}
            </section>
          ))}
        </div>
      )}
    </main>
  );
}
