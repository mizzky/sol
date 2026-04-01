"use client";

import React, { useCallback, useEffect, useState } from "react";
import { cancelOrder, getOrders, type OrderWithItems } from "../../lib/api";
import Badge from "../components/ui/Badge";
import Button from "../components/ui/Button";
import Card from "../components/ui/Card";
import { FieldMessage } from "../components/ui/Field";

function resolveStatusTone(status: string) {
  if (status === "pending") return "info" as const;
  if (status === "cancelled") return "danger" as const;
  return "success" as const;
}

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
    return (
      <main className="mx-auto max-w-6xl px-4 py-10 sm:px-6 lg:px-8">
        読み込み中...
      </main>
    );
  }

  return (
    <main className="mx-auto max-w-6xl px-4 py-10 sm:px-6 lg:px-8">
      <div className="mb-8 flex flex-col gap-3">
        <p className="text-sm uppercase tracking-[0.28em] text-indigo-600">
          Orders
        </p>
        <h1 className="text-4xl font-semibold tracking-tight text-zinc-900">
          注文履歴
        </h1>
        <p className="max-w-3xl text-sm leading-7 text-zinc-600 sm:text-base">
          注文状況と明細をカード単位で確認できる履歴画面です。
        </p>
      </div>
      {error && <FieldMessage tone="error">{error}</FieldMessage>}

      {orders.length === 0 ? (
        <Card className="mt-6 text-zinc-600">注文履歴はありません</Card>
      ) : (
        <div className="mt-6 grid gap-4">
          {orders.map(({ order, items }) => (
            <Card key={order.id} className="rounded-3xl p-6">
              <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
                <div>
                  <strong className="text-lg text-zinc-900">
                    注文ID: {order.id}
                  </strong>
                  <div className="mt-2 text-sm text-zinc-600">
                    作成日時: {order.created_at || "-"}
                  </div>
                </div>
                <Badge tone={resolveStatusTone(order.status)}>
                  ステータス: {order.status}
                </Badge>
              </div>

              <div className="mt-5 rounded-2xl bg-zinc-50 p-5 ring-1 ring-zinc-200">
                <div className="text-sm text-zinc-500">合計</div>
                <div className="mt-1 text-3xl font-semibold text-indigo-600">
                  ¥{order.total}
                </div>
                <div className="mt-1 text-sm text-zinc-500">
                  合計: ¥{order.total}
                </div>
              </div>

              <div className="mt-6">
                <strong className="text-sm uppercase tracking-[0.24em] text-zinc-500">
                  明細
                </strong>
                <ul className="mt-4 grid gap-3">
                  {items.map((item, idx) => (
                    <li
                      key={`${order.id}-${item.product_id}-${idx}`}
                      className="rounded-2xl bg-white px-4 py-3 text-sm text-zinc-700 ring-1 ring-zinc-200"
                    >
                      <div className="font-medium text-zinc-900">
                        {item.product_name_snapshot ||
                          `商品 #${item.product_id}`}
                      </div>
                      <div className="mt-1 text-zinc-600">
                        数量: {item.quantity} / 単価: ¥{item.unit_price}
                      </div>
                    </li>
                  ))}
                </ul>
              </div>

              {order.status === "pending" && (
                <div className="mt-6">
                  <Button
                    onClick={() => void handleCancel(order.id)}
                    disabled={cancellingId === order.id}
                    variant="outline"
                  >
                    {cancellingId === order.id
                      ? "キャンセル中..."
                      : "注文をキャンセル"}
                  </Button>
                </div>
              )}
            </Card>
          ))}
        </div>
      )}
    </main>
  );
}
