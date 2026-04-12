import React from "react";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { useRouter } from "next/navigation";
import AdminProductsPage from "../page";
import useAuthStore from "../../../../store/useAuthStore";

const mockGetProducts = jest.fn();
const mockGetCategories = jest.fn();
const mockCreateProduct = jest.fn();
const mockUpdateProduct = jest.fn();
const mockDeleteProduct = jest.fn();

jest.mock("next/navigation", () => ({
  useRouter: jest.fn(),
}));

jest.mock("../../../../lib/api", () => ({
  __esModule: true,
  getProducts: (...args: unknown[]) => mockGetProducts(...args),
  getCategories: (...args: unknown[]) => mockGetCategories(...args),
  createProduct: (...args: unknown[]) => mockCreateProduct(...args),
  updateProduct: (...args: unknown[]) => mockUpdateProduct(...args),
  deleteProduct: (...args: unknown[]) => mockDeleteProduct(...args),
}));

describe("AdminProductsPage", () => {
  const mockRouter = { push: jest.fn() };

  beforeEach(() => {
    jest.clearAllMocks();
    (useRouter as jest.Mock).mockReturnValue(mockRouter);
    useAuthStore.setState({
      isAuthenticated: true,
      user: { id: 1, name: "Admin", email: "admin@example.com", role: "admin" },
      loadFromStorage: jest.fn(async () => {}),
      logout: jest.fn(),
    } as any);
    mockGetProducts.mockResolvedValue([
      {
        id: 10,
        name: "House Blend",
        price: 980,
        is_available: true,
        category_id: 2,
        sku: "HB-001",
        description: "香ばしい定番ブレンド",
        image_url: null,
        stock_quantity: 10,
        created_at: "2026-03-30T00:00:00Z",
        updated_at: "2026-03-30T00:00:00Z",
      },
    ]);
    mockGetCategories.mockResolvedValue([
      { id: 2, name: "ブレンド", description: null },
      { id: 3, name: "シングルオリジン", description: null },
    ]);
    mockUpdateProduct.mockResolvedValue({});
    mockDeleteProduct.mockResolvedValue(undefined);
    window.confirm = jest.fn(() => true);
  });

  it("カテゴリ選択を使って商品を編集できる", async () => {
    render(<AdminProductsPage />);

    expect(await screen.findByText("House Blend")).toBeInTheDocument();
    expect(screen.queryByPlaceholderText("カテゴリID")).not.toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "編集" }));
    fireEvent.change(screen.getByPlaceholderText("商品名"), {
      target: { value: "Updated Blend" },
    });
    fireEvent.change(screen.getByLabelText("カテゴリ"), {
      target: { value: "3" },
    });
    fireEvent.click(screen.getByRole("button", { name: "更新する" }));

    await waitFor(() => {
      expect(mockUpdateProduct).toHaveBeenCalledWith(
        10,
        expect.objectContaining({ name: "Updated Blend", category_id: 3 }),
      );
    });
  });

  it("商品を削除できる", async () => {
    render(<AdminProductsPage />);

    expect(await screen.findByText("House Blend")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "削除" }));

    await waitFor(() => {
      expect(mockDeleteProduct).toHaveBeenCalledWith(10);
    });
  });
});
