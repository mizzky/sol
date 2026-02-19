import { fetchWithAuth } from "../api";
import useAuthStore from "../../store/useAuthStore";

describe("fetchWithAuth", () => {
  beforeEach(() => {
    localStorage.clear();
    useAuthStore.setState({ token: null, user: null });
    jest.resetAllMocks();
  });

  it("トークンあり: Authorizationヘッダが自動付与される", async () => {
    localStorage.setItem("auth_token", "test-token-123");

    global.fetch = jest.fn(() =>
      Promise.resolve({
        ok: true,
        json: async () => ({ success: true }),
      }) as unknown as Response,
    ) as unknown as typeof global.fetch;

    await fetchWithAuth("/api/products", { method: "GET" });

    expect(global.fetch).toHaveBeenCalledWith(
      expect.stringContaining("/api/products"),
      expect.objectContaining({
        headers: expect.objectContaining({
          Authorization: "Bearer test-token-123",
        }),
      }),
    );
  });

  it("トークンなし: Authorizationヘッダが付与されない", async () => {
    global.fetch = jest.fn(() =>
      Promise.resolve({
        ok: true,
        json: async () => ({ success: true }),
      }) as unknown as Response,
    ) as unknown as typeof global.fetch;

    await fetchWithAuth("/api/products", { method: "GET" });

    const callArgs = (global.fetch as jest.Mock).mock.calls[0];
    const headers = callArgs[1]?.headers || {};
    
    expect(headers.Authorization).toBeUndefined();
  });

  it("401エラー: 自動ログアウトが実行される", async () => {
    localStorage.setItem("auth_token", "expired-token");
    useAuthStore.getState().setToken("expired-token");

    global.fetch = jest.fn(() =>
      Promise.resolve({
        ok: false,
        status: 401,
        json: async () => ({ error: "Unauthorized" }),
      }) as unknown as Response,
    ) as unknown as typeof global.fetch;

    await expect(
      fetchWithAuth("/api/products", { method: "GET" }),
    ).rejects.toThrow("認証が必要です");

    // ログアウトが実行されたことを確認
    expect(useAuthStore.getState().token).toBeNull();
    expect(localStorage.getItem("auth_token")).toBeNull();
  });

  it("正常系: レスポンスを返す", async () => {
    localStorage.setItem("auth_token", "valid-token");

    const mockData = { id: 1, name: "Product" };
    global.fetch = jest.fn(() =>
      Promise.resolve({
        ok: true,
        status: 200,
        json: async () => mockData,
      }) as unknown as Response,
    ) as unknown as typeof global.fetch;

    const response = await fetchWithAuth("/api/products/1", { method: "GET" });
    const data = await response.json();

    expect(response.ok).toBe(true);
    expect(data).toEqual(mockData);
  });
});
