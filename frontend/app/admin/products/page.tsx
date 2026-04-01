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
import Badge from "../../components/ui/Badge";
import Button from "../../components/ui/Button";
import Card from "../../components/ui/Card";
import {
  CheckboxField,
  FieldMessage,
  FieldWrapper,
  Input,
  Select,
  Textarea,
} from "../../components/ui/Field";

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
      <main className="mx-auto max-w-7xl px-4 py-10 sm:px-6 lg:px-8">
        <h1 className="text-4xl font-semibold tracking-tight text-zinc-900">
          商品管理
        </h1>
        <p className="mt-3 mb-6 max-w-3xl text-sm leading-7 text-zinc-600 sm:text-base">
          商品の登録、編集、削除をこの画面で管理します。
        </p>
        <AdminNav />

        <div className="grid gap-4">
          {error && <FieldMessage tone="error">{error}</FieldMessage>}
          {success && <FieldMessage tone="success">{success}</FieldMessage>}
        </div>

        {loading ? (
          <div className="py-4">読み込み中...</div>
        ) : (
          <div className="grid gap-6 xl:grid-cols-[minmax(340px,420px)_minmax(0,1fr)]">
            <Card className="rounded-4xl p-6 sm:p-7">
              <h2 className="text-2xl font-semibold tracking-tight text-zinc-900">
                {isEditMode ? "商品編集" : "新規商品登録"}
              </h2>

              {categories.length === 0 ? (
                <FieldMessage tone="warning">
                  利用可能なカテゴリがありません。先に
                  <Link
                    href="/admin/categories"
                    className="ml-1 font-medium text-indigo-600 hover:text-indigo-500"
                  >
                    カテゴリ管理
                  </Link>
                  でカテゴリを作成してください。
                </FieldMessage>
              ) : (
                <form onSubmit={handleSubmit} className="mt-6 grid gap-5">
                  <FieldWrapper htmlFor="product-name" label="商品名">
                    <Input
                      id="product-name"
                      type="text"
                      placeholder="商品名"
                      value={form.name}
                      onChange={(e) => updateField("name", e.target.value)}
                      required
                    />
                  </FieldWrapper>
                  <FieldWrapper htmlFor="product-price" label="価格">
                    <Input
                      id="product-price"
                      type="number"
                      placeholder="980"
                      value={form.price}
                      onChange={(e) => updateField("price", e.target.value)}
                      min={1}
                      required
                    />
                  </FieldWrapper>
                  <FieldWrapper htmlFor="product-category" label="カテゴリ">
                    <Select
                      id="product-category"
                      aria-label="カテゴリ"
                      value={form.categoryId}
                      onChange={(e) =>
                        updateField("categoryId", e.target.value)
                      }
                      required
                    >
                      {categoryOptions.map((option) => (
                        <option key={option.value} value={option.value}>
                          {option.label}
                        </option>
                      ))}
                    </Select>
                  </FieldWrapper>
                  <FieldWrapper htmlFor="product-sku" label="SKU">
                    <Input
                      id="product-sku"
                      type="text"
                      placeholder="CF-2026-SPRING"
                      value={form.sku}
                      onChange={(e) => updateField("sku", e.target.value)}
                      required
                    />
                  </FieldWrapper>
                  <FieldWrapper htmlFor="product-stock" label="在庫数">
                    <Input
                      id="product-stock"
                      type="number"
                      placeholder="0"
                      value={form.stockQuantity}
                      onChange={(e) =>
                        updateField("stockQuantity", e.target.value)
                      }
                      min={0}
                    />
                  </FieldWrapper>
                  <FieldWrapper htmlFor="product-description" label="説明">
                    <Textarea
                      id="product-description"
                      placeholder="味の特徴や焙煎メモを入力"
                      value={form.description}
                      onChange={(e) =>
                        updateField("description", e.target.value)
                      }
                      rows={4}
                    />
                  </FieldWrapper>
                  <FieldWrapper
                    htmlFor="product-image-url"
                    label="画像URL（任意）"
                  >
                    <Input
                      id="product-image-url"
                      type="url"
                      placeholder="https://example.com/image.jpg"
                      value={form.imageUrl}
                      onChange={(e) => updateField("imageUrl", e.target.value)}
                    />
                  </FieldWrapper>
                  <CheckboxField
                    checked={form.isAvailable}
                    label="販売中にする"
                    onChange={(e) =>
                      updateField("isAvailable", e.target.checked)
                    }
                  />
                  <div className="flex flex-wrap gap-3">
                    <Button type="submit" disabled={submitting}>
                      {submitting
                        ? "保存中..."
                        : isEditMode
                          ? "更新する"
                          : "追加する"}
                    </Button>
                    {isEditMode && (
                      <Button
                        type="button"
                        onClick={resetForm}
                        variant="secondary"
                      >
                        編集をキャンセル
                      </Button>
                    )}
                  </div>
                </form>
              )}
            </Card>

            <section>
              <div className="mb-4 flex items-center justify-between gap-4">
                <h2 className="text-2xl font-semibold tracking-tight text-zinc-900">
                  登録済み商品
                </h2>
                <span className="text-sm text-zinc-500">
                  {products.length} items
                </span>
              </div>
              {products.length === 0 ? (
                <Card className="text-zinc-600">商品がありません</Card>
              ) : (
                <div className="grid gap-4">
                  {products.map((product) => (
                    <Card key={product.id} className="rounded-3xl p-5">
                      <div className="flex flex-col gap-5 lg:flex-row lg:items-start lg:justify-between">
                        <div className="min-w-0">
                          <div className="flex flex-wrap items-center gap-3">
                            <h3 className="text-xl font-semibold text-zinc-900">
                              {product.name}
                            </h3>
                            <Badge
                              tone={product.is_available ? "success" : "danger"}
                            >
                              {product.is_available ? "販売中" : "停止中"}
                            </Badge>
                          </div>
                          <div className="mt-3 grid gap-2 text-sm text-zinc-600 sm:grid-cols-2">
                            <div>価格: ¥{product.price}</div>
                            <div>カテゴリID: {product.category_id}</div>
                            <div>SKU: {product.sku}</div>
                            <div>在庫: {product.stock_quantity}</div>
                          </div>
                          {product.description && (
                            <p className="mt-4 text-sm leading-7 text-zinc-600">
                              {product.description}
                            </p>
                          )}
                        </div>
                        <div className="flex flex-wrap items-center gap-3 lg:flex-col lg:items-end">
                          <Link
                            href={`/products/${product.id}`}
                            className="text-sm font-medium text-indigo-600 hover:text-indigo-500"
                          >
                            公開ページを確認
                          </Link>
                          <Button
                            type="button"
                            onClick={() => startEdit(product)}
                            variant="secondary"
                          >
                            編集
                          </Button>
                          <Button
                            type="button"
                            onClick={() => void handleDelete(product.id)}
                            disabled={deletingProductId === product.id}
                            variant="danger"
                          >
                            {deletingProductId === product.id
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
        )}
      </main>
    </AdminRoute>
  );
}
