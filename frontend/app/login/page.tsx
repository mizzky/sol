"use client";
import React, { useState } from "react";
import { useRouter } from "next/navigation";
import useAuthStore from "../../store/useAuthStore";
import Button from "../components/ui/Button";
import Card from "../components/ui/Card";
import { FieldMessage, FieldWrapper, Input } from "../components/ui/Field";

export default function LoginPage() {
  const router = useRouter();

  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const onSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setLoading(true);
    try {
      await useAuthStore.getState().login(email, password);
      router.push("/");
    } catch (err: unknown) {
      console.error("login failed", err);
      setError(
        "ログインに失敗しました。メール／パスワードを確認してください。",
      );
    } finally {
      setLoading(false);
    }
  };

  return (
    <main className="mx-auto flex min-h-[calc(100vh-96px)] max-w-6xl items-center px-4 py-10 sm:px-6 lg:px-8">
      <div className="grid w-full gap-8 lg:grid-cols-[1.1fr_0.9fr]">
        <section className="rounded-4xl bg-linear-to-br from-indigo-600 to-indigo-500 p-8 text-white shadow-sm sm:p-10">
          <p className="text-sm uppercase tracking-[0.28em] text-indigo-100">
            Member Sign In
          </p>
          <h1 className="mt-4 text-4xl font-semibold tracking-tight">
            落ち着いた操作感で、すぐ業務に戻れるログイン画面へ。
          </h1>
          <p className="mt-4 max-w-xl text-sm leading-7 text-indigo-50/90 sm:text-base">
            入力欄の視認性を改善し、操作要素の余白と状態変化を整理しました。白いカード上で迷わず認証できます。
          </p>
        </section>

        <Card className="mx-auto w-full max-w-xl rounded-4xl p-8 sm:p-10">
          <div className="mb-6">
            <h1 className="text-3xl font-semibold tracking-tight text-zinc-900">
              ログイン
            </h1>
            <p className="mt-2 text-sm leading-6 text-zinc-600">
              登録済みメールアドレスとパスワードを入力してください。
            </p>
          </div>

          <form onSubmit={onSubmit} className="grid gap-5">
            <FieldWrapper htmlFor="login-email" label="メールアドレス">
              <Input
                id="login-email"
                type="email"
                placeholder="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                required
              />
            </FieldWrapper>
            <FieldWrapper htmlFor="login-password" label="パスワード">
              <Input
                id="login-password"
                type="password"
                placeholder="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
              />
            </FieldWrapper>
            {error && <FieldMessage tone="error">{error}</FieldMessage>}
            <Button
              type="submit"
              disabled={loading}
              className="w-full justify-center"
            >
              {loading ? "ログイン中..." : "ログイン"}
            </Button>
          </form>
        </Card>
      </div>
    </main>
  );
}
