import React from "react";
import { render, screen, act, waitFor } from "@testing-library/react";
import { useRouter } from "next/navigation";
import AdminRoute from "../AdminRoute";
import useAuthStore from "../../../store/useAuthStore";

jest.mock("next/navigation", () => ({
  useRouter: jest.fn(),
}));

describe("Admin Page (protected)", () => {
  const mockRouter = { push: jest.fn() };
  const AdminContent = () => <div>Admin Products Page</div>;

  beforeEach(() => {
    jest.clearAllMocks();
    (useRouter as jest.Mock).mockReturnValue(mockRouter);
    useAuthStore.setState({
      token: null,
      user: null,
      loadFromStorage: jest.fn(async () => {}),
      logout: jest.fn(),
    } as any);
    localStorage.clear();
  });

  it("未ログイン時は /login にリダイレクトされる", async () => {
    await act(async () => {
      render(
        <AdminRoute>
          <AdminContent />
        </AdminRoute>,
      );
    });

    await waitFor(() => {
      expect(mockRouter.push).toHaveBeenCalledWith("/login");
    });
    expect(screen.queryByText("Admin Products Page")).not.toBeInTheDocument();
  });

  it("ログイン（member）の場合は / にリダイレクトされる", async () => {
    useAuthStore.setState({
      token: "t",
      user: { id: 1, name: "Member", email: "m@e", role: "member" },
      loadFromStorage: jest.fn(async () => {}),
      logout: jest.fn(),
    } as any);

    await act(async () => {
      render(
        <AdminRoute>
          <AdminContent />
        </AdminRoute>,
      );
    });

    await waitFor(() => {
      expect(mockRouter.push).toHaveBeenCalledWith("/");
    });
    expect(screen.queryByText("Admin Products Page")).not.toBeInTheDocument();
  });

  it("ログイン（admin）の場合はコンテンツが表示される", async () => {
    useAuthStore.setState({
      token: "t",
      user: { id: 2, name: "Admin", email: "a@e", role: "admin" },
      loadFromStorage: jest.fn(async () => {}),
      logout: jest.fn(),
    } as any);

    await act(async () => {
      render(
        <AdminRoute>
          <AdminContent />
        </AdminRoute>,
      );
    });

    await waitFor(() => {
      expect(screen.getByText("Admin Products Page")).toBeInTheDocument();
    });
    expect(mockRouter.push).not.toHaveBeenCalled();
  });
});
