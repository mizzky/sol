import React from "react";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import ProductDetailPage from "../page";
import { useCartStore } from "../../../../store/useCartStore";

const mockUseParams = jest.fn();
const mockGetProductById = jest.fn();
const mockGetCategories = jest.fn();

jest.mock("next/navigation", () => ({
  useParams: () => mockUseParams(),
}));

jest.mock("../../../../lib/api", () => ({
  __esModule: true,
  getProductById: (...args: unknown[]) => mockGetProductById(...args),
  getCategories: (...args: unknown[]) => mockGetCategories(...args),
}));

describe("ProductDetailPage", () => {
  const addItem = jest.fn();

  beforeEach(() => {
    jest.clearAllMocks();
    mockUseParams.mockReturnValue({ id: "12" });
    mockGetProductById.mockResolvedValue({
      id: 12,
      name: "House Blend",
      price: 980,
      is_available: true,
      category_id: 2,
      sku: "HB-001",
      description: "香ばしい定番ブレンド",
      image_url: null,
      stock_quantity: 18,
      created_at: "2026-03-30T00:00:00Z",
      updated_at: "2026-03-30T00:00:00Z",
    });
    mockGetCategories.mockResolvedValue([
      { id: 2, name: "ブレンド", description: null },
    ]);
    useCartStore.setState({ addItem } as any);
  });

  it("商品詳細を表示してカート追加できる", async () => {
    render(<ProductDetailPage />);

    expect(
      await screen.findByRole("heading", { name: "House Blend" }),
    ).toBeInTheDocument();
    expect(screen.getByText("¥980")).toBeInTheDocument();
    expect(screen.getByText("ブレンド")).toBeInTheDocument();
    expect(screen.getByText("HB-001")).toBeInTheDocument();

    fireEvent.change(screen.getByDisplayValue("1"), { target: { value: "3" } });
    fireEvent.click(screen.getByRole("button", { name: "カートに追加" }));

    await waitFor(() => {
      expect(addItem).toHaveBeenCalledWith(12, 3);
    });
  });

  it("404のときはエラーメッセージを表示する", async () => {
    mockGetProductById.mockRejectedValue({ status: 404 });

    render(<ProductDetailPage />);

    expect(await screen.findByText("商品が見つかりません")).toBeInTheDocument();
  });
});
