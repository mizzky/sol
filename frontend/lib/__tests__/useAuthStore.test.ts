import useAuthStore from "../../store/useAuthStore";

describe("useAuthStore", () => {
  beforeEach(() => {
    // reset zustand state
    useAuthStore.setState({ token: null, user: null } as unknown as any);
    localStorage.clear();
    jest.resetAllMocks();
  });

  it("setToken saves to localStorage and state", () => {
    const setToken = useAuthStore.getState().setToken;
    setToken("abc-token");
    expect(localStorage.getItem("auth_token")).toBe("abc-token");
    expect(useAuthStore.getState().token).toBe("abc-token");
  });

  it("logout clears storage and state", () => {
    useAuthStore.getState().setToken("t");
    useAuthStore.getState().setUser({ id: 1, name: 'User', email: "a@b" });
    useAuthStore.getState().logout();

    expect(localStorage.getItem("auth_token")).toBeNull();
    expect(localStorage.getItem("auth_user")).toBeNull();
    expect(useAuthStore.getState().token).toBeNull();
    expect(useAuthStore.getState().user).toBeNull();
  });

  it("loadFromStorage calls /api/me and restores user when token present", async () => {
    const fakeUser = { id: 2, name: 'User', email: "me@example.com" };
    localStorage.setItem("auth_token", "tk-1");

    global.fetch = jest.fn(() =>
      Promise.resolve({
        ok: true,
        json: () => Promise.resolve({ user: fakeUser }),
      }) as unknown as Response,
    ) as unknown as typeof global.fetch;

    await useAuthStore.getState().loadFromStorage();

    expect(global.fetch).toHaveBeenCalled();
    expect(useAuthStore.getState().user).toEqual(fakeUser);
    expect(localStorage.getItem("auth_user")).toEqual(JSON.stringify(fakeUser));
  });

  describe("login", () => {
    it("正常系: ログイン成功でトークンとユーザー情報が保存される", async () => {
      const mockResponse = {
        token: "test-token-123",
        user: { id: 1, name: "Test User", email: "test@example.com" },
      };

      global.fetch = jest.fn(() =>
        Promise.resolve({
          ok: true,
          json: async () => mockResponse,
        }) as unknown as Response,
      ) as unknown as typeof global.fetch;

      const { login, token, user } = useAuthStore.getState();
      await login("test@example.com", "password");

      const state = useAuthStore.getState();
      expect(state.token).toBe("test-token-123");
      expect(state.user).toEqual(mockResponse.user);
      expect(localStorage.getItem("auth_token")).toBe("test-token-123");
      expect(localStorage.getItem("auth_user")).toBe(
        JSON.stringify(mockResponse.user),
      );
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