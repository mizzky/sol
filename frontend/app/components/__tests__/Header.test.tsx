import React from "react";
import { render, screen, fireEvent } from "@testing-library/react";
import { useRouter } from "next/navigation";
import Header from "../Header";
import useAuthStore from "../../../store/useAuthStore";

// next/navigation モックの設定
jest.mock("next/navigation", () => ({
  useRouter: jest.fn(),
}));

describe("Header Component", () => {
  const mockRouter = {
    push: jest.fn(),
  };

  beforeEach(() => {
    jest.clearAllMocks();
    (useRouter as jest.Mock).mockReturnValue(mockRouter);
    useAuthStore.setState({ token: null, user: null });
    localStorage.clear();
  });

  describe("未ログイン時", () => {
    it("ホーム、ログイン、登録リンクが表示される", () => {
      render(<Header />);

      expect(screen.getByText(/Home/i)).toBeInTheDocument();
      expect(screen.getByText(/Login/i)).toBeInTheDocument();
      expect(screen.getByText(/Sign Up/i)).toBeInTheDocument();
      expect(screen.queryByText(/Logout/i)).not.toBeInTheDocument();
    });
  });

  describe("ログイン時（member）", () => {
    beforeEach(() => {
      useAuthStore.setState({
        token: "test-token",
        user: {
          id: 1,
          name: "Test User",
          email: "test@example.com",
          role: "member",
        },
      });
    });

    it("ホーム、商品一覧、ユーザー名、ログアウトが表示される", () => {
      render(<Header />);

      expect(screen.getByText(/Home/i)).toBeInTheDocument();
      expect(screen.getByText(/Products/i)).toBeInTheDocument();
      expect(screen.getByText("Test User")).toBeInTheDocument();
      expect(screen.getByText(/Logout/i)).toBeInTheDocument();
      expect(screen.queryByText(/Admin/i)).not.toBeInTheDocument();
    });
  });

  describe("ログイン時（admin）", () => {
    beforeEach(() => {
      useAuthStore.setState({
        token: "test-token",
        user: {
          id: 2,
          name: "Admin User",
          email: "admin@example.com",
          role: "admin",
        },
      });
    });

    it("ホーム、商品一覧、管理ページ、ユーザー名、ログアウトが表示される", () => {
      render(<Header />);

      expect(screen.getByText(/Home/i)).toBeInTheDocument();
      expect(screen.getByText(/Products/i)).toBeInTheDocument();
      expect(screen.getByRole("link", { name: "Admin" })).toBeInTheDocument();
      expect(screen.getByText("Admin User")).toBeInTheDocument();
      expect(screen.getByText(/Logout/i)).toBeInTheDocument();
    });
  });

  describe("ログアウト機能", () => {
    beforeEach(() => {
      useAuthStore.setState({
        token: "test-token",
        user: {
          id: 1,
          name: "Test User",
          email: "test@example.com",
          role: "member",
        },
      });
    });

    it("ログアウトボタンクリックで認証情報がクリアされる", () => {
      render(<Header />);

      const logoutButton = screen.getByText(/Logout/i);
      fireEvent.click(logoutButton);

      expect(useAuthStore.getState().token).toBeNull();
      expect(useAuthStore.getState().user).toBeNull();
      expect(localStorage.getItem("auth_token")).toBeNull();
    });

    it("ログアウト後にホームページにリダイレクトされる", () => {
      render(<Header />);

      const logoutButton = screen.getByText(/Logout/i);
      fireEvent.click(logoutButton);

      expect(mockRouter.push).toHaveBeenCalledWith("/");
    });
  });
});
