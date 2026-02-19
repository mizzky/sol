import React from "react";
import "@testing-library/jest-dom";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import LoginPage from "../../app/login/page";
import * as api from "../../lib/api";

const mockPush = jest.fn();
const mockSetToken = jest.fn();
const mockSetUser = jest.fn();
const mockLogin = jest.fn();
const mockRegister = jest.fn();
const mockLogout = jest.fn();
const mockLoadFromStorage = jest.fn();

jest.mock("next/navigation", () => ({
  useRouter: () => ({ push: mockPush }),
}));

jest.mock("../../store/useAuthStore", () => {
  // Return a function that accepts a selector and invokes it with a full AuthState-like object
  const mock = (selector: (s: any) => any) =>
    selector({
      token: null,
      user: null,
      setToken: mockSetToken,
      setUser: mockSetUser,
      login: mockLogin,
      register: mockRegister,
      logout: mockLogout,
      loadFromStorage: mockLoadFromStorage,
    });

  // getState() を呼び出すための実装
  mock.getState = () => ({
    token: null,
    user: null,
    setToken: mockSetToken,
    setUser: mockSetUser,
    login: mockLogin,
    register: mockRegister,
    logout: mockLogout,
    loadFromStorage: mockLoadFromStorage,
  });

  return { __esModule: true, default: mock };
});

describe("LoginPage", () => {
  let consoleErrorSpy: jest.SpyInstance;

  beforeEach(() => {
    consoleErrorSpy = jest.spyOn(console, "error").mockImplementation(() => {});
    jest.resetAllMocks();
    localStorage.clear();
  });

  afterEach(() => {
    consoleErrorSpy.mockRestore();
  });

  it("submits credentials, updates store and redirects on success", async () => {
    mockLogin.mockResolvedValue(undefined);

    render(<LoginPage />);

    fireEvent.change(screen.getByPlaceholderText(/email/i), {
      target: { value: "user@example.com" },
    });
    fireEvent.change(screen.getByPlaceholderText(/password/i), {
      target: { value: "password123" },
    });

    fireEvent.click(screen.getByRole("button", { name: /ログイン/i }));

    await waitFor(() => {
      expect(mockLogin).toHaveBeenCalledWith("user@example.com", "password123");
      expect(mockPush).toHaveBeenCalledWith("/");
    });
  });

  it("shows error message on login failure", async () => {
    mockLogin.mockRejectedValue(new Error("invalid"));

    render(<LoginPage />);

    fireEvent.change(screen.getByPlaceholderText(/email/i), {
      target: { value: "bad@example.com" },
    });
    fireEvent.change(screen.getByPlaceholderText(/password/i), {
      target: { value: "wrong" },
    });

    fireEvent.click(screen.getByRole("button", { name: /ログイン/i }));

    await waitFor(() => {
      expect(screen.getByText(/ログインに失敗しました/i)).toBeInTheDocument();
    });
  });
});
