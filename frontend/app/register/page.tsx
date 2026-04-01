"use client";
import React, { useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import useAuthStore from "../../store/useAuthStore";
import Button from "../components/ui/Button";
import Card from "../components/ui/Card";
import { FieldMessage, FieldWrapper, Input } from "../components/ui/Field";

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
      const errorMessage =
        err && typeof err === "object" && "message" in err
          ? String(
              (err as { message?: unknown }).message ?? "登録に失敗しました",
            )
          : "登録に失敗しました";
      setError(errorMessage);
    } finally {
      setLoading(false);
    }
  };

  return (
    <main className="mx-auto flex min-h-[calc(100vh-96px)] max-w-6xl items-center px-4 py-10 sm:px-6 lg:px-8">
      <div className="grid w-full gap-8 lg:grid-cols-[0.95fr_1.05fr]">
        <Card className="order-2 rounded-4xl p-8 sm:order-1 sm:p-10">
          <div className="mb-6">
            <h1 className="text-3xl font-semibold tracking-tight text-zinc-900">
              ユーザー登録
            </h1>
            <p className="mt-2 text-sm leading-6 text-zinc-600">
              情報を入力するとすぐにアカウントを作成できます。
            </p>
          </div>

          <form onSubmit={onSubmit} className="grid gap-5">
            <FieldWrapper htmlFor="register-name" label="名前">
              <Input
                id="register-name"
                type="text"
                placeholder="名前"
                value={name}
                onChange={(e) => setName(e.target.value)}
                required
              />
            </FieldWrapper>
            <FieldWrapper htmlFor="register-email" label="メールアドレス">
              <Input
                id="register-email"
                type="email"
                placeholder="メールアドレス"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                required
              />
            </FieldWrapper>
            <FieldWrapper htmlFor="register-password" label="パスワード">
              <Input
                id="register-password"
                type="password"
                placeholder="パスワード"
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
              {loading ? "登録中..." : "登録"}
            </Button>
            <p className="text-center text-sm text-zinc-600">
              既にアカウントをお持ちですか？{" "}
              <Link
                href="/login"
                className="font-medium text-indigo-600 hover:text-indigo-500"
              >
                ログインページへ
              </Link>
            </p>
          </form>
        </Card>

        <section className="order-1 rounded-4xl bg-white p-8 shadow-sm ring-1 ring-zinc-200 sm:order-2 sm:p-10">
          <p className="text-sm uppercase tracking-[0.28em] text-indigo-600">
            Create Account
          </p>
          <h2 className="mt-4 text-4xl font-semibold tracking-tight text-zinc-900">
            広めの余白と高コントラストで、入力に集中できる登録画面へ。
          </h2>
          <div className="mt-8 grid gap-4 text-sm text-zinc-600 sm:grid-cols-2">
            <div className="rounded-2xl bg-zinc-50 p-5 ring-1 ring-zinc-200">
              <div className="font-semibold text-zinc-900">
                読みやすいフォーム
              </div>
              <p className="mt-2 leading-6">
                白背景の入力欄と zinc 系テキストで、入力内容を見失いません。
              </p>
            </div>
            <div className="rounded-2xl bg-zinc-50 p-5 ring-1 ring-zinc-200">
              <div className="font-semibold text-zinc-900">軽いモーション</div>
              <p className="mt-2 leading-6">
                ホバー時のみ影を深め、操作可能な要素を自然に伝えます。
              </p>
            </div>
          </div>
        </section>
      </div>
    </main>
  );
}
