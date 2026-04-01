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
import Button from "../../components/ui/Button";
import Card from "../../components/ui/Card";
import {
  FieldMessage,
  FieldWrapper,
  Input,
  Textarea,
} from "../../components/ui/Field";

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
      <main className="mx-auto max-w-6xl px-4 py-10 sm:px-6 lg:px-8">
        <h1 className="text-4xl font-semibold tracking-tight text-zinc-900">
          カテゴリ管理
        </h1>
        <p className="mt-3 mb-6 max-w-3xl text-sm leading-7 text-zinc-600 sm:text-base">
          商品登録で利用するカテゴリを管理します。
        </p>
        <AdminNav />

        <div className="grid gap-4">
          {error && <FieldMessage tone="error">{error}</FieldMessage>}
          {success && <FieldMessage tone="success">{success}</FieldMessage>}
        </div>

        <div className="mt-6 grid gap-6 xl:grid-cols-[minmax(300px,360px)_minmax(0,1fr)]">
          <Card className="rounded-4xl p-6 sm:p-7">
            <h2 className="text-2xl font-semibold tracking-tight text-zinc-900">
              {editingCategoryId === null ? "新規カテゴリ" : "カテゴリ編集"}
            </h2>
            <form onSubmit={handleSubmit} className="mt-6 grid gap-5">
              <FieldWrapper htmlFor="category-name" label="カテゴリ名">
                <Input
                  id="category-name"
                  type="text"
                  placeholder="カテゴリ名"
                  value={form.name}
                  onChange={(e) =>
                    setForm((current) => ({ ...current, name: e.target.value }))
                  }
                  required
                />
              </FieldWrapper>
              <FieldWrapper htmlFor="category-description" label="説明">
                <Textarea
                  id="category-description"
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
              </FieldWrapper>
              <div className="flex flex-wrap gap-3">
                <Button type="submit" disabled={submitting}>
                  {submitting
                    ? "保存中..."
                    : editingCategoryId === null
                      ? "作成する"
                      : "更新する"}
                </Button>
                {editingCategoryId !== null && (
                  <Button type="button" onClick={resetForm} variant="secondary">
                    編集をキャンセル
                  </Button>
                )}
              </div>
            </form>
          </Card>

          <section>
            <h2 className="mb-4 text-2xl font-semibold tracking-tight text-zinc-900">
              カテゴリ一覧
            </h2>
            {loading ? (
              <div>読み込み中...</div>
            ) : categories.length === 0 ? (
              <Card className="text-zinc-600">カテゴリがありません</Card>
            ) : (
              <div className="grid gap-4">
                {categories.map((category) => (
                  <Card key={category.id} className="rounded-3xl p-5">
                    <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
                      <div>
                        <h3 className="text-xl font-semibold text-zinc-900">
                          {category.name}
                        </h3>
                        <div className="mt-2 text-sm text-zinc-500">
                          ID: {category.id}
                        </div>
                        <p className="mt-3 text-sm leading-7 text-zinc-600">
                          {category.description || "説明なし"}
                        </p>
                      </div>
                      <div className="flex flex-wrap gap-3 sm:flex-col">
                        <Button
                          type="button"
                          onClick={() => startEdit(category)}
                          variant="secondary"
                        >
                          編集
                        </Button>
                        <Button
                          type="button"
                          onClick={() => void handleDelete(category.id)}
                          disabled={deletingCategoryId === category.id}
                          variant="danger"
                        >
                          {deletingCategoryId === category.id
                            ? "削除中..."
                            : "削除"}
                        </Button>
                      </div>
                    </div>
                  </Card>
                ))}
              </div>
            )}
          </section>
        </div>
      </main>
    </AdminRoute>
  );
}
