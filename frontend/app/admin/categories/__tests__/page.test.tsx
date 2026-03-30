import React from "react";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { useRouter } from "next/navigation";
import AdminCategoriesPage from "../page";
import useAuthStore from "../../../../store/useAuthStore";

const mockGetCategories = jest.fn();
const mockCreateCategory = jest.fn();
const mockUpdateCategory = jest.fn();
const mockDeleteCategory = jest.fn();

jest.mock("next/navigation", () => ({
  useRouter: jest.fn(),
}));

jest.mock("../../../../lib/api", () => ({
  __esModule: true,
  getCategories: (...args: unknown[]) => mockGetCategories(...args),
  createCategory: (...args: unknown[]) => mockCreateCategory(...args),
  updateCategory: (...args: unknown[]) => mockUpdateCategory(...args),
  deleteCategory: (...args: unknown[]) => mockDeleteCategory(...args),
}));

describe("AdminCategoriesPage", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    (useRouter as jest.Mock).mockReturnValue({ push: jest.fn() });
    useAuthStore.setState({
      token: "admin-token",
      user: { id: 1, name: "Admin", email: "admin@example.com", role: "admin" },
      loadFromStorage: jest.fn(async () => {}),
      logout: jest.fn(),
    } as any);
    mockGetCategories.mockResolvedValue([
      { id: 1, name: "定番", description: "人気商品" },
    ]);
    mockCreateCategory.mockResolvedValue({});
    mockUpdateCategory.mockResolvedValue({});
    mockDeleteCategory.mockResolvedValue(undefined);
    window.confirm = jest.fn(() => true);
  });

  it("カテゴリを作成できる", async () => {
    render(<AdminCategoriesPage />);

    expect(await screen.findByText("定番")).toBeInTheDocument();

    fireEvent.change(screen.getByPlaceholderText("カテゴリ名"), {
      target: { value: "深煎り" },
    });
    fireEvent.change(screen.getByPlaceholderText("説明"), {
      target: { value: "ビター系" },
    });
    fireEvent.click(screen.getByRole("button", { name: "作成する" }));

    await waitFor(() => {
      expect(mockCreateCategory).toHaveBeenCalledWith({
        name: "深煎り",
        description: "ビター系",
      });
    });
  });

  it("カテゴリを削除できる", async () => {
    render(<AdminCategoriesPage />);

    expect(await screen.findByText("定番")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "削除" }));

    await waitFor(() => {
      expect(mockDeleteCategory).toHaveBeenCalledWith(1);
    });
  });
});
