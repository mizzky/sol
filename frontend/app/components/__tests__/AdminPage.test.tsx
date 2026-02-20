import React from "react";
import { render, screen } from "@testing-library/react";
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
    useAuthStore.setState({ token: null, user: null });
    localStorage.clear();
  });

  it("未ログイン時は /login にリダイレクトされる", () => {
    render(
      <AdminRoute>
        <AdminContent />
      </AdminRoute>,
    );
    expect(mockRouter.push).toHaveBeenCalledWith("/login");
    expect(screen.queryByText("Admin Products Page")).not.toBeInTheDocument();
  });

  it("ログイン（member）の場合は / にリダイレクトされる", () => {
    useAuthStore.setState({
      token: "t",
      user: { id: 1, name: "Member", email: "m@e", role: "member" },
    });

    render(
      <AdminRoute>
        <AdminContent />
      </AdminRoute>,
    );
    expect(mockRouter.push).toHaveBeenCalledWith("/");
    expect(screen.queryByText("Admin Products Page")).not.toBeInTheDocument();
  });

  it("ログイン（admin）の場合はコンテンツが表示される", () => {
    useAuthStore.setState({
      token: "t",
      user: { id: 2, name: "Admin", email: "a@e", role: "admin" },
    });

    render(
      <AdminRoute>
        <AdminContent />
      </AdminRoute>,
    );
    expect(mockRouter.push).not.toHaveBeenCalled();
    expect(screen.getByText("Admin Products Page")).toBeInTheDocument();
  });
});
