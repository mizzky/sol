"use client";

import React, { useCallback, useEffect, useState } from "react";
import AdminRoute from "../../components/AdminRoute";
import AdminNav from "../../components/AdminNav";
import {
  type Category,
  createCategory,
  deleteCategory,
  getCategories,
  updateCategory,
} from "../../../lib/api";

type CategoryFormState = {
  name: string;
  description: string;
};

const initialFormState: CategoryFormState = {
  name: "",
  description: "",
};

function getErrorMessage(error: unknown, fallback: string): string {
  const status = (error as { status?: number } | null)?.status;
  if (status === 400) return "入力内容を確認してください";
  if (status === 401) return "認証が必要です";
  if (status === 403) return "管理者権限が必要です";
  if (status === 404) return "カテゴリが見つかりません";
  return fallback;
}

export default function AdminCategoriesPage() {
  const [categories, setCategories] = useState<Category[]>([]);
  const [form, setForm] = useState<CategoryFormState>(initialFormState);
  const [editingCategoryId, setEditingCategoryId] = useState<number | null>(
    null,
  );
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [deletingCategoryId, setDeletingCategoryId] = useState<number | null>(
    null,
  );
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  const loadCategories = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const response = await getCategories();
      setCategories(response);
    } catch (err: unknown) {
      setError(getErrorMessage(err, "カテゴリ一覧の取得に失敗しました"));
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void loadCategories();
  }, [loadCategories]);

  const resetForm = () => {
    setEditingCategoryId(null);
    setForm(initialFormState);
  };

  const handleSubmit = async (event: React.FormEvent) => {
    event.preventDefault();
    setSubmitting(true);
    setError(null);
    setSuccess(null);

    try {
      if (editingCategoryId === null) {
        await createCategory({
          name: form.name.trim(),
          description: form.description.trim() ? form.description.trim() : null,
        });
        setSuccess("カテゴリを作成しました");
      } else {
        await updateCategory(editingCategoryId, {
          name: form.name.trim(),
          description: form.description.trim() ? form.description.trim() : null,
        });
        setSuccess("カテゴリを更新しました");
      }
      resetForm();
      await loadCategories();
    } catch (err: unknown) {
      setError(
        getErrorMessage(
          err,
          editingCategoryId === null
            ? "カテゴリ作成に失敗しました"
            : "カテゴリ更新に失敗しました",
        ),
      );
    } finally {
      setSubmitting(false);
    }
  };

  const startEdit = (category: Category) => {
    setEditingCategoryId(category.id);
    setForm({ name: category.name, description: category.description ?? "" });
    setError(null);
    setSuccess(null);
  };

  const handleDelete = async (categoryId: number) => {
    if (
      typeof window !== "undefined" &&
      !window.confirm("このカテゴリを削除しますか？")
    ) {
      return;
    }

    setDeletingCategoryId(categoryId);
    setError(null);
    setSuccess(null);

    try {
      await deleteCategory(categoryId);
      setSuccess("カテゴリを削除しました");
      if (editingCategoryId === categoryId) {
        resetForm();
      }
      await loadCategories();
    } catch (err: unknown) {
      setError(getErrorMessage(err, "カテゴリ削除に失敗しました"));
    } finally {
      setDeletingCategoryId(null);
    }
  };

  return (
    <AdminRoute>
      <main style={{ padding: "2rem", maxWidth: "960px", margin: "0 auto" }}>
        <h1>カテゴリ管理</h1>
        <p style={{ color: "#57534e", marginBottom: "1rem" }}>
          商品登録で利用するカテゴリを管理します。
        </p>
        <AdminNav />

        {error && (
          <div style={{ color: "crimson", marginBottom: "1rem" }}>{error}</div>
        )}
        {success && (
          <div style={{ color: "green", marginBottom: "1rem" }}>{success}</div>
        )}

        <div
          style={{
            display: "grid",
            gridTemplateColumns: "minmax(280px, 340px) minmax(0, 1fr)",
            gap: "1.5rem",
          }}
        >
          <section
            style={{
              border: "1px solid #d6d3d1",
              borderRadius: 12,
              padding: "1.25rem",
              background: "#fffdf8",
            }}
          >
            <h2 style={{ marginTop: 0 }}>
              {editingCategoryId === null ? "新規カテゴリ" : "カテゴリ編集"}
            </h2>
            <form
              onSubmit={handleSubmit}
              style={{ display: "grid", gap: "0.75rem" }}
            >
              <input
                type="text"
                placeholder="カテゴリ名"
                value={form.name}
                onChange={(e) =>
                  setForm((current) => ({ ...current, name: e.target.value }))
                }
                required
              />
              <textarea
                placeholder="説明"
                rows={4}
                value={form.description}
                onChange={(e) =>
                  setForm((current) => ({
                    ...current,
                    description: e.target.value,
                  }))
                }
              />
              <div style={{ display: "flex", gap: "0.75rem" }}>
                <button type="submit" disabled={submitting}>
                  {submitting
                    ? "保存中..."
                    : editingCategoryId === null
                      ? "作成する"
                      : "更新する"}
                </button>
                {editingCategoryId !== null && (
                  <button type="button" onClick={resetForm}>
                    編集をキャンセル
                  </button>
                )}
              </div>
            </form>
          </section>

          <section>
            <h2 style={{ marginTop: 0 }}>カテゴリ一覧</h2>
            {loading ? (
              <div>読み込み中...</div>
            ) : categories.length === 0 ? (
              <div>カテゴリがありません</div>
            ) : (
              <div style={{ display: "grid", gap: "0.75rem" }}>
                {categories.map((category) => (
                  <article
                    key={category.id}
                    style={{
                      border: "1px solid #e7e5e4",
                      borderRadius: 12,
                      padding: "1rem",
                    }}
                  >
                    <div
                      style={{
                        display: "flex",
                        justifyContent: "space-between",
                        gap: "1rem",
                      }}
                    >
                      <div>
                        <h3 style={{ margin: "0 0 0.4rem" }}>
                          {category.name}
                        </h3>
                        <div>ID: {category.id}</div>
                        <p style={{ marginBottom: 0, color: "#57534e" }}>
                          {category.description || "説明なし"}
                        </p>
                      </div>
                      <div
                        style={{
                          display: "flex",
                          flexDirection: "column",
                          gap: "0.5rem",
                        }}
                      >
                        <button
                          type="button"
                          onClick={() => startEdit(category)}
                        >
                          編集
                        </button>
                        <button
                          type="button"
                          onClick={() => void handleDelete(category.id)}
                          disabled={deletingCategoryId === category.id}
                        >
                          {deletingCategoryId === category.id
                            ? "削除中..."
                            : "削除"}
                        </button>
                      </div>
                    </div>
                  </article>
                ))}
              </div>
            )}
          </section>
        </div>
      </main>
    </AdminRoute>
  );
}
