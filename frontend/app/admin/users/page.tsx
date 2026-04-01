"use client";

import React, { useState } from "react";
import AdminRoute from "../../components/AdminRoute";
import AdminNav from "../../components/AdminNav";
import { setUserRole } from "../../../lib/api";
import Button from "../../components/ui/Button";
import Card from "../../components/ui/Card";
import {
  FieldMessage,
  FieldWrapper,
  Input,
  Select,
} from "../../components/ui/Field";

type RoleValue = "admin" | "member";

function getErrorMessage(error: unknown): string {
  const status = (error as { status?: number } | null)?.status;
  if (status === 400) return "入力内容を確認してください";
  if (status === 401) return "認証が必要です";
  if (status === 403) return "管理者権限が必要です";
  if (status === 404) return "対象ユーザーが見つかりません";
  return "権限更新に失敗しました";
}

export default function AdminUsersPage() {
  const [userId, setUserId] = useState("");
  const [role, setRole] = useState<RoleValue>("member");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);
  const [lastUpdated, setLastUpdated] = useState<{
    userId: number;
    role: RoleValue;
  } | null>(null);

  const handleSubmit = async (event: React.FormEvent) => {
    event.preventDefault();
    setSubmitting(true);
    setError(null);
    setSuccess(null);

    const parsedUserId = Number(userId);
    if (!Number.isInteger(parsedUserId) || parsedUserId <= 0) {
      setSubmitting(false);
      setError("ユーザーIDは正の整数で入力してください");
      return;
    }

    try {
      await setUserRole(parsedUserId, role);
      setLastUpdated({ userId: parsedUserId, role });
      setSuccess("ユーザー権限を更新しました");
    } catch (err: unknown) {
      setError(getErrorMessage(err));
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <AdminRoute>
      <main className="mx-auto max-w-5xl px-4 py-10 sm:px-6 lg:px-8">
        <h1 className="text-4xl font-semibold tracking-tight text-zinc-900">
          ユーザー権限管理
        </h1>
        <p className="mt-3 mb-6 max-w-3xl text-sm leading-7 text-zinc-600 sm:text-base">
          現在はユーザー一覧APIが未提供のため、対象ユーザーIDを指定して権限を更新する暫定UIです。
        </p>
        <AdminNav />

        <div className="grid gap-4">
          {error && <FieldMessage tone="error">{error}</FieldMessage>}
          {success && <FieldMessage tone="success">{success}</FieldMessage>}
        </div>

        <Card className="mt-6 rounded-4xl p-6 sm:p-7">
          <h2 className="text-2xl font-semibold tracking-tight text-zinc-900">
            権限を更新する
          </h2>
          <form onSubmit={handleSubmit} noValidate className="mt-6 grid gap-5">
            <FieldWrapper htmlFor="target-user-id" label="対象ユーザーID">
              <Input
                id="target-user-id"
                type="number"
                placeholder="対象ユーザーID"
                value={userId}
                onChange={(e) => setUserId(e.target.value)}
                min={1}
                required
              />
            </FieldWrapper>
            <FieldWrapper htmlFor="role-select" label="適用ロール">
              <Select
                id="role-select"
                value={role}
                onChange={(e) => setRole(e.target.value as RoleValue)}
              >
                <option value="member">member</option>
                <option value="admin">admin</option>
              </Select>
            </FieldWrapper>
            <Button type="submit" disabled={submitting} className="w-fit">
              {submitting ? "更新中..." : "権限を更新する"}
            </Button>
          </form>
        </Card>

        {lastUpdated && (
          <Card className="mt-4 rounded-3xl p-5">
            <h2 className="text-xl font-semibold text-zinc-900">
              直近の更新内容
            </h2>
            <div className="mt-3 grid gap-2 text-sm text-zinc-600">
              <div>ユーザーID: {lastUpdated.userId}</div>
              <div>適用ロール: {lastUpdated.role}</div>
            </div>
          </Card>
        )}
      </main>
    </AdminRoute>
  );
}
