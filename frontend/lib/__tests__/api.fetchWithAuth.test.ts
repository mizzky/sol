import { fetchWithAuth } from "../api";
import useAuthStore from "../../store/useAuthStore";

describe("fetchWithAuth", () => {
  beforeEach(() => {
    localStorage.clear();
    useAuthStore.setState({ isAuthenticated: false, user: null });
    jest.resetAllMocks();
  });

  it("credentials: include でリクエストされる", async () => {
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
        credentials: "include",
      }),
    );
  });

  it("401時: refresh成功なら元リクエストを再試行する", async () => {
    global.fetch = jest
      .fn()
      .mockResolvedValueOnce({
        ok: false,
        status: 401,
        json: async () => ({ error: "Unauthorized" }),
      } as unknown as Response)
      .mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: async () => ({}),
      } as unknown as Response)
      .mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: async () => ({ success: true }),
      } as unknown as Response) as unknown as typeof global.fetch;

    const res = await fetchWithAuth("/api/products", { method: "GET" });

    expect(res.ok).toBe(true);
    expect(global.fetch).toHaveBeenNthCalledWith(
      2,
      expect.stringContaining("/api/refresh"),
      expect.objectContaining({ method: "POST", credentials: "include" }),
    );
    expect((global.fetch as jest.Mock).mock.calls).toHaveLength(3);
  });

  it("401時: refresh失敗ならログアウトされる", async () => {
    useAuthStore.setState({
      isAuthenticated: true,
      user: { id: 1, name: "User", email: "u@example.com", role: "member" },
    });

    global.fetch = jest.fn(() =>
      Promise.resolve({
        ok: false,
        status: 401,
        json: async () => ({ error: "Unauthorized" }),
      }) as unknown as Response,
    ) as unknown as typeof global.fetch;

    await expect(fetchWithAuth("/api/products", { method: "GET" })).rejects.toThrow(
      "認証が必要です",
    );

    expect(useAuthStore.getState().isAuthenticated).toBe(false);
    expect(useAuthStore.getState().user).toBeNull();
  });

  it("並行401時: refresh は1回だけ実行され、成功後に各リクエストを再試行する", async () => {
    let refreshCalls = 0;
    const requestCounts = new Map<string, number>();

    global.fetch = jest.fn((input: RequestInfo | URL) => {
      const url = String(input);

      if (url.includes("/api/refresh")) {
        refreshCalls += 1;
        return Promise.resolve({
          ok: true,
          status: 200,
          json: async () => ({ message: "refreshed" }),
        } as unknown as Response);
      }

      const count = requestCounts.get(url) ?? 0;
      requestCounts.set(url, count + 1);

      if (count === 0) {
        return Promise.resolve({
          ok: false,
          status: 401,
          json: async () => ({ error: "Unauthorized" }),
        } as unknown as Response);
      }

      return Promise.resolve({
        ok: true,
        status: 200,
        json: async () => ({ success: true }),
      } as unknown as Response);
    }) as unknown as typeof global.fetch;

    const [first, second] = await Promise.all([
      fetchWithAuth("/api/products", { method: "GET" }),
      fetchWithAuth("/api/cart", { method: "GET" }),
    ]);

    expect(first.ok).toBe(true);
    expect(second.ok).toBe(true);
    expect(refreshCalls).toBe(1);
    expect(requestCounts.get("/api/products")).toBe(2);
    expect(requestCounts.get("/api/cart")).toBe(2);
  });

  it("並行401かつrefresh失敗時: revoke は1回だけ実行される", async () => {
    useAuthStore.setState({
      isAuthenticated: true,
      user: { id: 1, name: "User", email: "u@example.com", role: "member" },
    });

    let refreshCalls = 0;
    let revokeCalls = 0;

    global.fetch = jest.fn((input: RequestInfo | URL) => {
      const url = String(input);

      if (url.includes("/api/refresh/revoke")) {
        revokeCalls += 1;
        return Promise.resolve({
          ok: true,
          status: 200,
          json: async () => ({}),
        } as unknown as Response);
      }

      if (url.includes("/api/refresh")) {
        refreshCalls += 1;
        return Promise.resolve({
          ok: false,
          status: 401,
          json: async () => ({ error: "Unauthorized" }),
        } as unknown as Response);
      }

      return Promise.resolve({
        ok: false,
        status: 401,
        json: async () => ({ error: "Unauthorized" }),
      } as unknown as Response);
    }) as unknown as typeof global.fetch;

    await Promise.allSettled([
      fetchWithAuth("/api/products", { method: "GET" }),
      fetchWithAuth("/api/cart", { method: "GET" }),
    ]);
    await Promise.resolve();

    expect(refreshCalls).toBe(1);
    expect(revokeCalls).toBe(1);
    expect(useAuthStore.getState().isAuthenticated).toBe(false);
    expect(useAuthStore.getState().user).toBeNull();
  });

  it("正常系: レスポンスを返す", async () => {
    global.fetch = jest.fn(() =>
      Promise.resolve({
        ok: true,
        status: 200,
        json: async () => ({ success: true }),
      }) as unknown as Response,
    ) as unknown as typeof global.fetch;

    await fetchWithAuth("/api/products", { method: "GET" });

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
