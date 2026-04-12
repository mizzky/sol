import React from "react";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { act } from "react";
import { useRouter } from "next/navigation";
import AdminUsersPage from "../page";
import useAuthStore from "../../../../store/useAuthStore";

const mockSetUserRole = jest.fn();

jest.mock("next/navigation", () => ({
  useRouter: jest.fn(),
}));

jest.mock("../../../../lib/api", () => ({
  __esModule: true,
  setUserRole: (...args: unknown[]) => mockSetUserRole(...args),
}));

describe("AdminUsersPage", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    (useRouter as jest.Mock).mockReturnValue({ push: jest.fn() });
    useAuthStore.setState({
      isAuthenticated: true,
      user: { id: 1, name: "Admin", email: "admin@example.com", role: "admin" },
      loadFromStorage: jest.fn(async () => {}),
      logout: jest.fn(),
    } as any);
    mockSetUserRole.mockResolvedValue({});
  });

  it("ユーザーIDの入力を検証する", async () => {
    await act(async () => {
      render(<AdminUsersPage />);
    });

    fireEvent.change(screen.getByPlaceholderText("対象ユーザーID"), {
      target: { value: "0" },
    });
    fireEvent.click(screen.getByRole("button", { name: "権限を更新する" }));

    expect(
      await screen.findByText("ユーザーIDは正の整数で入力してください"),
    ).toBeInTheDocument();
    expect(mockSetUserRole).not.toHaveBeenCalled();
  });

  it("ユーザー権限を更新できる", async () => {
    await act(async () => {
      render(<AdminUsersPage />);
    });

    fireEvent.change(screen.getByPlaceholderText("対象ユーザーID"), {
      target: { value: "42" },
    });
    fireEvent.change(screen.getByDisplayValue("member"), {
      target: { value: "admin" },
    });
    fireEvent.click(screen.getByRole("button", { name: "権限を更新する" }));

    await waitFor(() => {
      expect(mockSetUserRole).toHaveBeenCalledWith(42, "admin");
    });

    expect(
      await screen.findByText("ユーザー権限を更新しました"),
    ).toBeInTheDocument();
  });
});
