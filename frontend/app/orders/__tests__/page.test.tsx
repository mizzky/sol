import React from "react";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import OrdersPage from "../page";

const mockGetOrders = jest.fn();
const mockCancelOrder = jest.fn();

jest.mock("../../../lib/api", () => ({
  __esModule: true,
  getOrders: (...args: unknown[]) => mockGetOrders(...args),
  cancelOrder: (...args: unknown[]) => mockCancelOrder(...args),
}));

describe("OrdersPage", () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it("loads and renders orders", async () => {
    mockGetOrders.mockResolvedValue([
      {
        order: {
          id: 1,
          status: "pending",
          total: 1800,
          created_at: "2026-03-29T10:00:00Z",
        },
        items: [
          {
            product_id: 10,
            product_name_snapshot: "House Blend",
            quantity: 2,
            unit_price: 900,
          },
        ],
      },
    ]);

    render(<OrdersPage />);

    await waitFor(() => {
      expect(mockGetOrders).toHaveBeenCalledTimes(1);
    });

    expect(
      await screen.findByRole("heading", { name: "注文履歴" }),
    ).toBeInTheDocument();
    expect(screen.getByText("注文ID: 1")).toBeInTheDocument();
    expect(screen.getByText("ステータス: pending")).toBeInTheDocument();
    expect(screen.getByText("合計: ¥1800")).toBeInTheDocument();
    expect(screen.getByText(/House Blend/)).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: "注文をキャンセル" }),
    ).toBeInTheDocument();
  });

  it("calls cancel and reloads orders", async () => {
    mockGetOrders
      .mockResolvedValueOnce([
        {
          order: {
            id: 1,
            status: "pending",
            total: 1800,
            created_at: "2026-03-29T10:00:00Z",
          },
          items: [],
        },
      ])
      .mockResolvedValueOnce([
        {
          order: {
            id: 1,
            status: "cancelled",
            total: 1800,
            created_at: "2026-03-29T10:00:00Z",
          },
          items: [],
        },
      ]);
    mockCancelOrder.mockResolvedValue({
      id: 1,
      status: "cancelled",
      total: 1800,
    });

    render(<OrdersPage />);

    await waitFor(() => {
      expect(
        screen.getByRole("button", { name: "注文をキャンセル" }),
      ).toBeInTheDocument();
    });

    fireEvent.click(screen.getByRole("button", { name: "注文をキャンセル" }));

    await waitFor(() => {
      expect(mockCancelOrder).toHaveBeenCalledWith(1);
      expect(mockGetOrders).toHaveBeenCalledTimes(2);
    });

    expect(screen.getByText("ステータス: cancelled")).toBeInTheDocument();
  });

  it("shows empty state", async () => {
    mockGetOrders.mockResolvedValue([]);

    render(<OrdersPage />);

    await waitFor(() => {
      expect(mockGetOrders).toHaveBeenCalledTimes(1);
    });

    expect(await screen.findByText("注文履歴はありません")).toBeInTheDocument();
  });
});
