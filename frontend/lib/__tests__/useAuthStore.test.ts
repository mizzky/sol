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
});