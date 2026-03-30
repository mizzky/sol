"use client";

import Link from "next/link";
import React, { useCallback, useEffect, useMemo, useState } from "react";
import AdminRoute from "../../../app/components/AdminRoute";
import AdminNav from "../../../app/components/AdminNav";
import {
  type Category,
  type CreateProductRequest,
  type Product,
  createProduct,
  deleteProduct,
  getCategories,
  getProducts,
  updateProduct,
} from "../../../lib/api";

type ProductFormState = {
  name: string;
  price: string;
  categoryId: string;
  sku: string;
  stockQuantity: string;
  isAvailable: boolean;
  description: string;
  imageUrl: string;
};

const initialFormState: ProductFormState = {
  name: "",
  price: "",
  categoryId: "",
  sku: "",
  stockQuantity: "0",
  isAvailable: true,
  description: "",
  imageUrl: "",
};

function toFormState(product?: Product): ProductFormState {
  if (!product) {
    return initialFormState;
  }

  return {
    name: product.name,
    price: String(product.price),
    categoryId: String(product.category_id),
    sku: product.sku,
    stockQuantity: String(product.stock_quantity),
    isAvailable: product.is_available,
    description: product.description ?? "",
    imageUrl: product.image_url ?? "",
  };
}

function getErrorMessage(error: unknown, fallback: string): string {
  const status = (error as { status?: number } | null)?.status;
  if (status === 400) return "入力内容を確認してください";
  if (status === 401) return "認証が必要です";
  if (status === 403) return "管理者権限が必要です";
  if (status === 404) return "対象データが見つかりません";
  if (status === 409) return "SKUが重複しています";
  return fallback;
}

export default function AdminProductsPage() {
  const [products, setProducts] = useState<Product[]>([]);
  const [categories, setCategories] = useState<Category[]>([]);
  const [form, setForm] = useState<ProductFormState>(initialFormState);
  const [editingProductId, setEditingProductId] = useState<number | null>(null);
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [deletingProductId, setDeletingProductId] = useState<number | null>(
    null,
  );
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  const loadData = useCallback(async () => {
    setLoading(true);
    setError(null);

    try {
      const [productList, categoryList] = await Promise.all([
        getProducts(),
        getCategories(),
      ]);
      setProducts(productList);
      setCategories(categoryList);
      setForm((current) => {
        if (current.categoryId || categoryList.length === 0) {
          return current;
        }
        return { ...current, categoryId: String(categoryList[0].id) };
      });
    } catch (err: unknown) {
      setError(getErrorMessage(err, "商品管理データの取得に失敗しました"));
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void loadData();
  }, [loadData]);

  const isEditMode = editingProductId !== null;
  const categoryOptions = useMemo(
    () =>
      categories.map((category) => ({
        value: String(category.id),
        label: category.name,
      })),
    [categories],
  );

  const updateField = <K extends keyof ProductFormState>(
    key: K,
    value: ProductFormState[K],
  ) => {
    setForm((current) => ({ ...current, [key]: value }));
  };

  const resetForm = () => {
    setEditingProductId(null);
    setForm({
      ...initialFormState,
      categoryId: categories.length > 0 ? String(categories[0].id) : "",
    });
  };

  const handleSubmit = async (event: React.FormEvent) => {
    event.preventDefault();
    setError(null);
    setSuccess(null);
    setSubmitting(true);

    const payload: CreateProductRequest = {
      name: form.name.trim(),
      price: Number(form.price || 0),
      is_available: form.isAvailable,
      category_id: Number(form.categoryId || 0),
      sku: form.sku.trim(),
      description: form.description.trim() ? form.description.trim() : null,
      image_url: form.imageUrl.trim() ? form.imageUrl.trim() : null,
      stock_quantity: Number(form.stockQuantity || 0),
    };

    try {
      if (isEditMode && editingProductId !== null) {
        await updateProduct(editingProductId, payload);
        setSuccess("商品を更新しました");
      } else {
        await createProduct(payload);
        setSuccess("商品を追加しました");
      }
      resetForm();
      await loadData();
    } catch (err: unknown) {
      setError(
        getErrorMessage(
          err,
          isEditMode ? "商品更新に失敗しました" : "商品追加に失敗しました",
        ),
      );
    } finally {
      setSubmitting(false);
    }
  };

  const startEdit = (product: Product) => {
    setEditingProductId(product.id);
    setForm(toFormState(product));
    setError(null);
    setSuccess(null);
  };

  const handleDelete = async (productId: number) => {
    if (
      typeof window !== "undefined" &&
      !window.confirm("この商品を削除しますか？")
    ) {
      return;
    }

    setDeletingProductId(productId);
    setError(null);
    setSuccess(null);

    try {
      await deleteProduct(productId);
      setSuccess("商品を削除しました");
      if (editingProductId === productId) {
        resetForm();
      }
      await loadData();
    } catch (err: unknown) {
      setError(getErrorMessage(err, "商品削除に失敗しました"));
    } finally {
      setDeletingProductId(null);
    }
  };

  return (
    <AdminRoute>
      <main style={{ padding: "2rem", maxWidth: "1100px", margin: "0 auto" }}>
        <h1>商品管理</h1>
        <p style={{ color: "#57534e", marginBottom: "1rem" }}>
          商品の登録、編集、削除をこの画面で管理します。
        </p>
        <AdminNav />

        {error && (
          <div style={{ color: "crimson", marginBottom: "1rem" }}>{error}</div>
        )}
        {success && (
          <div style={{ color: "green", marginBottom: "1rem" }}>{success}</div>
        )}

        {loading ? (
          <div style={{ padding: "1rem 0" }}>読み込み中...</div>
        ) : (
          <div
            style={{
              display: "grid",
              gap: "1.5rem",
              gridTemplateColumns: "minmax(320px, 380px) minmax(0, 1fr)",
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
                {isEditMode ? "商品編集" : "新規商品登録"}
              </h2>

              {categories.length === 0 ? (
                <div style={{ color: "#b45309" }}>
                  利用可能なカテゴリがありません。先に
                  <Link
                    href="/admin/categories"
                    style={{ color: "#0f766e", marginLeft: 4 }}
                  >
                    カテゴリ管理
                  </Link>
                  でカテゴリを作成してください。
                </div>
              ) : (
                <form
                  onSubmit={handleSubmit}
                  style={{ display: "grid", gap: "0.75rem" }}
                >
                  <input
                    type="text"
                    placeholder="商品名"
                    value={form.name}
                    onChange={(e) => updateField("name", e.target.value)}
                    required
                  />
                  <input
                    type="number"
                    placeholder="価格"
                    value={form.price}
                    onChange={(e) => updateField("price", e.target.value)}
                    min={1}
                    required
                  />
                  <select
                    aria-label="カテゴリ"
                    value={form.categoryId}
                    onChange={(e) => updateField("categoryId", e.target.value)}
                    required
                  >
                    {categoryOptions.map((option) => (
                      <option key={option.value} value={option.value}>
                        {option.label}
                      </option>
                    ))}
                  </select>
                  <input
                    type="text"
                    placeholder="SKU"
                    value={form.sku}
                    onChange={(e) => updateField("sku", e.target.value)}
                    required
                  />
                  <input
                    type="number"
                    placeholder="在庫数"
                    value={form.stockQuantity}
                    onChange={(e) =>
                      updateField("stockQuantity", e.target.value)
                    }
                    min={0}
                  />
                  <textarea
                    placeholder="説明"
                    value={form.description}
                    onChange={(e) => updateField("description", e.target.value)}
                    rows={4}
                  />
                  <input
                    type="url"
                    placeholder="画像URL（任意）"
                    value={form.imageUrl}
                    onChange={(e) => updateField("imageUrl", e.target.value)}
                  />
                  <label
                    style={{
                      display: "flex",
                      gap: "0.5rem",
                      alignItems: "center",
                    }}
                  >
                    <input
                      type="checkbox"
                      checked={form.isAvailable}
                      onChange={(e) =>
                        updateField("isAvailable", e.target.checked)
                      }
                    />
                    販売中にする
                  </label>
                  <div style={{ display: "flex", gap: "0.75rem" }}>
                    <button type="submit" disabled={submitting}>
                      {submitting
                        ? "保存中..."
                        : isEditMode
                          ? "更新する"
                          : "追加する"}
                    </button>
                    {isEditMode && (
                      <button type="button" onClick={resetForm}>
                        編集をキャンセル
                      </button>
                    )}
                  </div>
                </form>
              )}
            </section>

            <section>
              <h2 style={{ marginTop: 0 }}>登録済み商品</h2>
              {products.length === 0 ? (
                <div>商品がありません</div>
              ) : (
                <div style={{ display: "grid", gap: "0.75rem" }}>
                  {products.map((product) => (
                    <article
                      key={product.id}
                      style={{
                        border: "1px solid #e7e5e4",
                        borderRadius: 12,
                        padding: "1rem",
                        background: "#ffffff",
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
                            {product.name}
                          </h3>
                          <div>価格: ¥{product.price}</div>
                          <div>カテゴリID: {product.category_id}</div>
                          <div>SKU: {product.sku}</div>
                          <div>在庫: {product.stock_quantity}</div>
                          <div>
                            状態: {product.is_available ? "販売中" : "停止中"}
                          </div>
                        </div>
                        <div
                          style={{
                            display: "flex",
                            flexDirection: "column",
                            gap: "0.5rem",
                            alignItems: "flex-end",
                          }}
                        >
                          <Link
                            href={`/products/${product.id}`}
                            style={{ color: "#0f766e" }}
                          >
                            公開ページを確認
                          </Link>
                          <button
                            type="button"
                            onClick={() => startEdit(product)}
                          >
                            編集
                          </button>
                          <button
                            type="button"
                            onClick={() => void handleDelete(product.id)}
                            disabled={deletingProductId === product.id}
                          >
                            {deletingProductId === product.id
                              ? "削除中..."
                              : "削除"}
                          </button>
                        </div>
                      </div>
                      {product.description && (
                        <p style={{ marginBottom: 0, color: "#57534e" }}>
                          {product.description}
                        </p>
                      )}
                    </article>
                  ))}
                </div>
              )}
            </section>
          </div>
        )}
      </main>
    </AdminRoute>
  );
}
