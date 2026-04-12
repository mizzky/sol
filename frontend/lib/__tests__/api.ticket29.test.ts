import {
  cancelOrder,
  createCategory,
  createOrder,
  deleteCategory,
  deleteProduct,
  getCategories,
  getOrders,
  getProductById,
  setUserRole,
  updateCategory,
  updateProduct,
} from "../api";

describe("ticket29 api functions", () => {
  beforeEach(() => {
    localStorage.clear();
    jest.resetAllMocks();
  });

  const mockResponse = (body: unknown, ok = true, status = 200) =>
    Promise.resolve({
      ok,
      status,
      json: async () => body,
    } as unknown as Response);

  it("getCategories returns categories", async () => {
    global.fetch = jest.fn(() =>
      mockResponse({ categories: [{ id: 1, name: "Beans", description: null }] }),
    ) as unknown as typeof global.fetch;

    await expect(getCategories()).resolves.toEqual([
      { id: 1, name: "Beans", description: null },
    ]);
  });

  it("createCategory throws on bad request", async () => {
    global.fetch = jest.fn(() =>
      mockResponse({ error: "カテゴリ名は必須です" }, false, 400),
    ) as unknown as typeof global.fetch;

    await expect(createCategory({ name: "" })).rejects.toMatchObject({ status: 400 });
  });

  it("updateCategory returns updated category", async () => {
    global.fetch = jest.fn(() =>
      mockResponse({ id: 1, name: "Hot", description: "updated" }),
    ) as unknown as typeof global.fetch;

    await expect(updateCategory(1, { name: "Hot", description: "updated" })).resolves.toEqual({
      id: 1,
      name: "Hot",
      description: "updated",
    });
  });

  it("deleteCategory resolves on no content", async () => {
    global.fetch = jest.fn(() =>
      mockResponse({}, true, 204),
    ) as unknown as typeof global.fetch;

    await expect(deleteCategory(2)).resolves.toBeUndefined();
  });

  it("getProductById returns product", async () => {
    global.fetch = jest.fn(() =>
      mockResponse({
        id: 10,
        name: "Mocha",
        price: 700,
        is_available: true,
        category_id: 1,
        sku: "MOCHA-001",
        stock_quantity: 12,
        created_at: "2026-03-29T00:00:00Z",
        updated_at: "2026-03-29T00:00:00Z",
      }),
    ) as unknown as typeof global.fetch;

    await expect(getProductById(10)).resolves.toMatchObject({ id: 10, name: "Mocha" });
  });

  it("updateProduct throws on conflict", async () => {
    global.fetch = jest.fn(() =>
      mockResponse({ error: "SKUが既に存在します" }, false, 409),
    ) as unknown as typeof global.fetch;

    await expect(
      updateProduct(10, {
        name: "Mocha",
        price: 700,
        is_available: true,
        category_id: 1,
        sku: "DUPLICATED",
        description: null,
        image_url: null,
        stock_quantity: 10,
      }),
    ).rejects.toMatchObject({ status: 409 });
  });

  it("deleteProduct resolves on no content", async () => {
    global.fetch = jest.fn(() =>
      mockResponse({}, true, 204),
    ) as unknown as typeof global.fetch;

    await expect(deleteProduct(10)).resolves.toBeUndefined();
  });

  it("setUserRole returns payload", async () => {
    global.fetch = jest.fn(() =>
      mockResponse({ id: 3, role: "admin" }),
    ) as unknown as typeof global.fetch;

    await expect(setUserRole(3, "admin")).resolves.toEqual({ id: 3, role: "admin" });
  });

  it("getOrders normalizes backend payload", async () => {
    global.fetch = jest.fn(() =>
      mockResponse({
        orders: [
          {
            order: {
              ID: 9,
              UserID: 1,
              Total: 1200,
              Status: "pending",
              CreatedAt: "2026-03-29T00:00:00Z",
              UpdatedAt: "2026-03-29T00:00:00Z",
            },
            items: [
              {
                ID: 100,
                OrderID: 9,
                ProductID: 10,
                Quantity: 2,
                UnitPrice: 600,
                ProductNameSnapshot: "House Blend",
              },
            ],
          },
        ],
      }),
    ) as unknown as typeof global.fetch;

    await expect(getOrders()).resolves.toEqual([
      {
        order: {
          id: 9,
          user_id: 1,
          total: 1200,
          status: "pending",
          created_at: "2026-03-29T00:00:00Z",
          updated_at: "2026-03-29T00:00:00Z",
          cancelled_at: null,
        },
        items: [
          {
            id: 100,
            order_id: 9,
            product_id: 10,
            quantity: 2,
            unit_price: 600,
            product_name_snapshot: "House Blend",
          },
        ],
      },
    ]);
  });

  it("createOrder returns normalized order", async () => {
    global.fetch = jest.fn(() =>
      mockResponse({
        order: {
          ID: 11,
          UserID: 1,
          Total: 800,
          Status: "pending",
          CreatedAt: "2026-03-29T00:00:00Z",
          UpdatedAt: "2026-03-29T00:00:00Z",
        },
      }, true, 201),
    ) as unknown as typeof global.fetch;

    await expect(createOrder()).resolves.toEqual({
      id: 11,
      user_id: 1,
      total: 800,
      status: "pending",
      created_at: "2026-03-29T00:00:00Z",
      updated_at: "2026-03-29T00:00:00Z",
      cancelled_at: null,
    });
  });

  it("cancelOrder throws on bad request", async () => {
    global.fetch = jest.fn(() =>
      mockResponse({ error: "この注文はキャンセルできません" }, false, 400),
    ) as unknown as typeof global.fetch;

    await expect(cancelOrder(11)).rejects.toMatchObject({ status: 400 });
  });
});
