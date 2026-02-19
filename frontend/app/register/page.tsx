"use client";
import React, { useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import useAuthStore from "../../store/useAuthStore";

export default function RegisterPage() {
  const router = useRouter();

  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const onSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setLoading(true);

    try {
      await useAuthStore.getState().register(name, email, password);
      router.push("/login");
    } catch (err: unknown) {
      console.error("register failed", err);
      const errorMessage = (err as any)?.message || "登録に失敗しました";
      setError(errorMessage);
    } finally {
      setLoading(false);
    }
  };

  return (
    <main style={{ padding: "2rem", maxWidth: 480, margin: "0 auto" }}>
      <h1>ユーザー登録</h1>
      <form onSubmit={onSubmit} style={{ display: "grid", gap: 8 }}>
        <input
          type="text"
          placeholder="名前"
          value={name}
          onChange={(e) => setName(e.target.value)}
          required
        />
        <input
          type="email"
          placeholder="メールアドレス"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          required
        />
        <input
          type="password"
          placeholder="パスワード"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          required
        />
        {error && <div style={{ color: "crimson" }}>{error}</div>}
        <button type="submit" disabled={loading}>
          {loading ? "登録中..." : "登録"}
        </button>
        <p style={{ fontSize: "0.9rem", textAlign: "center" }}>
          既にアカウントをお持ちですか？{" "}
          <Link href="/login" style={{ color: "blue" }}>
            ログインページへ
          </Link>
        </p>
      </form>
    </main>
  );
}
