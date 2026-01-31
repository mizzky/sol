"use client";
import { useEffect, useState } from "react";

// Productsの型定義
interface Product {
  id: number;
  name: string;
  price: number; // 追加
  is_available: boolean; // 追加
}

export default function Home() {
  const [products, setProducts] = useState<Product[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const apiUrl = process.env.NEXT_PUBLIC_API_URL;

    // GoのAPIを叩く
    fetch(`${apiUrl}/api/products`)
      .then((res) => {
        if (!res.ok) throw new Error("API接続に失敗しました");
        return res.json();
      })
      .then((data: Product[]) => {
        setProducts(data);
        setLoading(false);
      })
      .catch((err) => {
        console.error(err);
        setLoading(false);
      });
  }, []);

  if (loading) return <div style={{ padding: "2rem" }}>読み込み中...</div>;

  return (
    <main style={{ padding: "2rem", maxWidth: "600px", margin: "0 auto" }}>
      <h1 style={{ fontSize: "2rem", marginBottom: "1rem" }}>
        ☕ Sol Coffee System
      </h1>

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
          {products.map((p) => (
            <div
              key={p.id}
              style={{
                padding: "1rem",
                border: "1px solid #ddd",
                borderRadius: "8px",
              }}
            >
              {/* 3. 変数を p に統一して、プロパティを呼び出す */}
              <h3 style={{ margin: 0 }}>{p.name}</h3>
              <p style={{ margin: "5px 0", color: "#666" }}>価格: ¥{p.price}</p>
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
          ))}
        </ul>
      </div>
    </main>
  );
}
