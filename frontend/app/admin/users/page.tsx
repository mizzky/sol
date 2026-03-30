"use client";

import React, { useState } from "react";
import AdminRoute from "../../components/AdminRoute";
import AdminNav from "../../components/AdminNav";
import { setUserRole } from "../../../lib/api";

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
      <main style={{ padding: "2rem", maxWidth: "760px", margin: "0 auto" }}>
        <h1>ユーザー権限管理</h1>
        <p style={{ color: "#57534e", marginBottom: "0.75rem" }}>
          現在はユーザー一覧APIが未提供のため、対象ユーザーIDを指定して権限を更新する暫定UIです。
        </p>
        <AdminNav />

        {error && (
          <div style={{ color: "crimson", marginBottom: "1rem" }}>{error}</div>
        )}
        {success && (
          <div style={{ color: "green", marginBottom: "1rem" }}>{success}</div>
        )}

        <section
          style={{
            border: "1px solid #d6d3d1",
            borderRadius: 12,
            padding: "1.25rem",
            background: "#fffdf8",
          }}
        >
          <h2 style={{ marginTop: 0 }}>権限を更新する</h2>
          <form
            onSubmit={handleSubmit}
            noValidate
            style={{ display: "grid", gap: "0.75rem" }}
          >
            <input
              type="number"
              placeholder="対象ユーザーID"
              value={userId}
              onChange={(e) => setUserId(e.target.value)}
              min={1}
              required
            />
            <select
              value={role}
              onChange={(e) => setRole(e.target.value as RoleValue)}
            >
              <option value="member">member</option>
              <option value="admin">admin</option>
            </select>
            <button type="submit" disabled={submitting}>
              {submitting ? "更新中..." : "権限を更新する"}
            </button>
          </form>
        </section>

        {lastUpdated && (
          <section
            style={{
              marginTop: "1rem",
              border: "1px solid #e7e5e4",
              borderRadius: 12,
              padding: "1rem",
            }}
          >
            <h2 style={{ marginTop: 0 }}>直近の更新内容</h2>
            <div>ユーザーID: {lastUpdated.userId}</div>
            <div>適用ロール: {lastUpdated.role}</div>
          </section>
        )}
      </main>
    </AdminRoute>
  );
}
