import useAuthStore from "../../store/useAuthStore";

describe("useAuthStore", () => {
  beforeEach(() => {
    // reset zustand state
    useAuthStore.setState({ isAuthenticated: false, user: null } as unknown as any);
    localStorage.clear();
    jest.resetAllMocks();
  });

  it("setUser updates user and isAuthenticated", () => {
    const setUser = useAuthStore.getState().setUser;
    setUser({ id: 1, name: "User", email: "a@b", role: "member" });
    expect(useAuthStore.getState().isAuthenticated).toBe(true);
    expect(useAuthStore.getState().user?.name).toBe("User");
  });

  it("logout clears state", () => {
    global.fetch = jest.fn(() =>
      Promise.resolve({ ok: true, status: 200, json: async () => ({}) }) as unknown as Response,
    ) as unknown as typeof global.fetch;

    useAuthStore.getState().setUser({ id: 1, name: "User", email: "a@b", role: "member" });
    useAuthStore.getState().logout();

    expect(useAuthStore.getState().isAuthenticated).toBe(false);
    expect(useAuthStore.getState().user).toBeNull();
  });

  it("loadFromStorage calls /api/me and restores user", async () => {
    const fakeUser = { id: 2, name: "User", email: "me@example.com", role: "member" };

    global.fetch = jest.fn(() =>
      Promise.resolve({
        ok: true,
        json: () => Promise.resolve({ user: fakeUser }),
      }) as unknown as Response,
    ) as unknown as typeof global.fetch;

    await useAuthStore.getState().loadFromStorage();

    expect(global.fetch).toHaveBeenCalled();
    expect(useAuthStore.getState().user).toEqual(fakeUser);
    expect(useAuthStore.getState().isAuthenticated).toBe(true);
  });

  it("loadFromStorage refresh後に /api/me を再試行して復元できる", async () => {
    const fakeUser = { id: 3, name: "Refreshed User", email: "refresh@example.com", role: "member" };

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
        json: async () => ({ message: "refreshed" }),
      } as unknown as Response)
      .mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: async () => ({ user: fakeUser }),
      } as unknown as Response) as unknown as typeof global.fetch;

    await useAuthStore.getState().loadFromStorage();

    expect(global.fetch).toHaveBeenNthCalledWith(
      2,
      expect.stringContaining("/api/refresh"),
      expect.objectContaining({ method: "POST", credentials: "include" }),
    );
    expect(useAuthStore.getState().user).toEqual(fakeUser);
    expect(useAuthStore.getState().isAuthenticated).toBe(true);
  });

  describe("login", () => {
    it("正常系: ログイン成功でユーザー情報が保存される", async () => {
      const mockResponse = {
        user: { id: 1, name: "Test User", email: "test@example.com", role: "member" },
      };

      global.fetch = jest.fn(() =>
        Promise.resolve({
          ok: true,
          json: async () => mockResponse,
        }) as unknown as Response,
      ) as unknown as typeof global.fetch;

      const { login } = useAuthStore.getState();
      await login("test@example.com", "password");

      const state = useAuthStore.getState();
      expect(state.isAuthenticated).toBe(true);
      expect(state.user).toEqual(mockResponse.user);
    });

    it("異常系: 401エラーでエラーがthrowされる", async () => {
      global.fetch = jest.fn(() =>
        Promise.resolve({
          ok: false,
          status: 401,
          json: async () => ({ error: "Unauthorized" }),
        }) as unknown as Response,
      ) as unknown as typeof global.fetch;

      const { login } = useAuthStore.getState();
      await expect(
        login("test@example.com", "wrong-password"),
      ).rejects.toThrow();
    });

    it("異常系: network errorでエラーがthrowされる", async () => {
      global.fetch = jest.fn(() =>
        Promise.reject(new Error("Network error")),
      ) as unknown as typeof global.fetch;

      const { login } = useAuthStore.getState();
      await expect(
        login("test@example.com", "password"),
      ).rejects.toThrow();
    });
  });

  describe("register", () => {
    it("正常系: 登録成功でAPIが呼ばれる", async () => {
      const mockResponse = {
        id: 1,
        name: "New User",
        email: "new@example.com",
        role: "member",
      };

      global.fetch = jest.fn(() =>
        Promise.resolve({
          ok: true,
          status: 201,
          json: async () => mockResponse,
        }) as unknown as Response,
      ) as unknown as typeof global.fetch;

      const { register } = useAuthStore.getState();
      await register("New User", "new@example.com", "password123");

      expect(global.fetch).toHaveBeenCalledWith(
        expect.stringContaining("/api/register"),
        expect.objectContaining({
          method: "POST",
          body: JSON.stringify({
            name: "New User",
            email: "new@example.com",
            password: "password123",
          }),
        }),
      );
    });

    it("異常系: 400エラー(重複メール等)でエラーがthrowされる", async () => {
      global.fetch = jest.fn(() =>
        Promise.resolve({
          ok: false,
          status: 400,
          json: async () => ({ error: "このメールアドレスは既に登録されています" }),
        }) as unknown as Response,
      ) as unknown as typeof global.fetch;

      const { register } = useAuthStore.getState();
      await expect(
        register("New User", "duplicate@example.com", "password"),
      ).rejects.toThrow();
    });
  });
});