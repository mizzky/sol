import React from "react";
import { render, screen, act, waitFor } from "@testing-library/react";
import { useRouter } from "next/navigation";
import AdminRoute from "../AdminRoute";
import useAuthStore from "../../../store/useAuthStore";

jest.mock("next/navigation", () => ({
  useRouter: jest.fn(),
}));

describe("AdminRoute 強化テスト", () => {
  const mockRouter = { push: jest.fn() };

  beforeEach(() => {
    jest.clearAllMocks();
    (useRouter as jest.Mock).mockReturnValue(mockRouter);
    localStorage.clear();
    useAuthStore.setState({
      token: null,
      user: null,
      loadFromStorage: jest.fn(),
      logout: jest.fn(),
    } as any);
  });

  it("ロード中（loadFromStorage が未解決）の間は即時リダイレクトしないこと（期待：Red -> 実装で Green）", async () => {
    let resolvePromise: () => void;
    const deferred = new Promise<void>((res) => {
      resolvePromise = res;
    });
    useAuthStore.setState({
      token: null,
      user: null,
      loadFromStorage: jest.fn(() => deferred),
      logout: jest.fn(),
    } as any);

    render(
      <AdminRoute>
        <div>Admin Content</div>
      </AdminRoute>,
    );

    expect(mockRouter.push).not.toHaveBeenCalled();

    await act(async () => {
      resolvePromise();
      await deferred;
    });

    expect(mockRouter.push).toHaveBeenCalledWith("/login");
  });

  it("/api/me 相当の処理で 401 が発生した場合は logout が呼ばれ /login にリダイレクトされること（期待：Red -> 実装で Green）", async () => {
    const logoutMock = jest.fn();
    useAuthStore.setState({
      token: null,
      user: null,
      loadFromStorage: jest.fn(async () => {
        logoutMock();
      }),
      logout: logoutMock,
    } as any);

    await act(async () => {
      render(
        <AdminRoute>
          <div>Admin Content</div>
        </AdminRoute>,
      );
    });

    expect(logoutMock).toHaveBeenCalled();
    expect(mockRouter.push).toHaveBeenCalledWith("/login");
  });

  it("ログイン済みだが role が 'member' の場合は / へリダイレクトされる", async () => {
    useAuthStore.setState({
      token: "t",
      user: { id: 1, name: "Member", email: "m@e", role: "member" },
      loadFromStorage: jest.fn(async () => {}), // ← async にして完了を待てるように
      logout: jest.fn(),
    } as any);

    await act(async () => {
      render(
        <AdminRoute>
          <div>Admin Content</div>
        </AdminRoute>,
      );
    });

    await waitFor(() => {
      expect(mockRouter.push).toHaveBeenCalledWith("/");
    });
    expect(screen.queryByText("Admin Content")).not.toBeInTheDocument();
  });

  it("ログイン済みで role が 'admin' の場合はコンテンツが表示される", async () => {
    useAuthStore.setState({
      token: "t",
      user: { id: 2, name: "Admin", email: "a@e", role: "admin" },
      loadFromStorage: jest.fn(async () => {}), // ← async にして完了を待てるように
      logout: jest.fn(),
    } as any);

    await act(async () => {
      render(
        <AdminRoute>
          <div>Admin Content</div>
        </AdminRoute>,
      );
    });

    await waitFor(() => {
      expect(screen.getByText("Admin Content")).toBeInTheDocument();
    });
    expect(mockRouter.push).not.toHaveBeenCalled();
  });
});
