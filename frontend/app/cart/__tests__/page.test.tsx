import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import CartPage from "../page";
import React from "react";

const mockPush = jest.fn();
const mockCreateOrder = jest.fn();

const mockSyncCart = jest.fn();
const mockUpdateItem = jest.fn();
const mockRemoveItem = jest.fn();
const mockClearCart = jest.fn();

const mockState = {
  items: [
    {
      id: 1,
      cart_id: 1,
      product_id: 7,
      quantity: 2,
      price: 500,
      product_price: 500,
      product_stock: 12,
      product_name: "House Blend",
    },
  ],
  totalPrice: 1000,
  totalQuantity: 2,
  loading: false,
  error: null,
  updateItem: mockUpdateItem,
  removeItem: mockRemoveItem,
  clearCart: mockClearCart,
  syncCart: mockSyncCart,
};

jest.mock("../../../store/useCartStore", () => ({
  __esModule: true,
  default: () => mockState,
}));

jest.mock("next/navigation", () => ({
  useRouter: () => ({
    push: mockPush,
  }),
}));

jest.mock("../../../lib/api", () => ({
  __esModule: true,
  createOrder: (...args: unknown[]) => mockCreateOrder(...args),
}));

describe("CartPage", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    mockCreateOrder.mockResolvedValue({
      id: 1,
      status: "pending",
      total: 1000,
    });
  });

  it("renders cart items from backend-compatible fields", () => {
    render(<CartPage />);

    expect(screen.getByRole("heading", { name: /カート/ })).toBeInTheDocument();
    expect(screen.getByText("House Blend")).toBeInTheDocument();
    expect(screen.getByText("単価: ¥500")).toBeInTheDocument();
    expect(screen.getByText("在庫: 12")).toBeInTheDocument();
    expect(screen.getByText("小計: ¥1000")).toBeInTheDocument();
    expect(screen.getByText("合計数量: 2")).toBeInTheDocument();
    expect(screen.getByText("合計金額: ¥1000")).toBeInTheDocument();
  });

  it("syncs cart on mount", async () => {
    render(<CartPage />);

    await waitFor(() => {
      expect(mockSyncCart).toHaveBeenCalledTimes(1);
    });
  });

  it("calls update and remove actions", () => {
    render(<CartPage />);

    fireEvent.click(screen.getByRole("button", { name: "+" }));
    fireEvent.click(screen.getByRole("button", { name: "削除" }));
    fireEvent.click(screen.getByRole("button", { name: "カートを空にする" }));

    expect(mockUpdateItem).toHaveBeenCalledWith(1, 3);
    expect(mockRemoveItem).toHaveBeenCalledWith(1);
    expect(mockClearCart).toHaveBeenCalledTimes(1);
  });

  it("calls createOrder and navigates to orders on checkout", async () => {
    render(<CartPage />);

    fireEvent.click(
      screen.getByRole("button", { name: "チェックアウトへ進む" }),
    );

    await waitFor(() => {
      expect(mockCreateOrder).toHaveBeenCalledTimes(1);
      expect(mockPush).toHaveBeenCalledWith("/orders");
    });
  });
});
