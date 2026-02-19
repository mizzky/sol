import React from "react";
import "@testing-library/jest-dom";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import RegisterPage from "../../app/register/page";

const mockPush = jest.fn();
const mockRegister = jest.fn();

jest.mock("next/navigation", () => ({
  useRouter: () => ({ push: mockPush }),
}));

jest.mock("../../store/useAuthStore", () => {
  const mock = (selector: (s: any) => any) =>
    selector({
      token: null,
      user: null,
      register: mockRegister,
      login: jest.fn(),
      setToken: jest.fn(),
      setUser: jest.fn(),
      logout: jest.fn(),
      loadFromStorage: jest.fn(),
    });
  
  mock.getState = () => ({
    token: null,
    user: null,
    register: mockRegister,
    login: jest.fn(),
    setToken: jest.fn(),
    setUser: jest.fn(),
    logout: jest.fn(),
    loadFromStorage: jest.fn(),
  });
  
  return { __esModule: true, default: mock };
});

describe("RegisterPage", () => {
  let consoleErrorSpy: jest.SpyInstance;

  beforeEach(() => {
    consoleErrorSpy = jest.spyOn(console, "error").mockImplementation(() => {});
    jest.resetAllMocks();
  });

  afterEach(() => {
    consoleErrorSpy.mockRestore();
  });

  // テストケース1: 正常系
  it("正常系: 有効な情報で登録し、ログインページへリダイレクト", async () => {
    mockRegister.mockResolvedValue(undefined);

    render(<RegisterPage />);

    fireEvent.change(screen.getByPlaceholderText(/名前/i), {
      target: { value: "田中 太郎" },
    });
    fireEvent.change(screen.getByPlaceholderText(/メールアドレス/i), {
      target: { value: "tanaka@example.com" },
    });
    fireEvent.change(screen.getByPlaceholderText(/パスワード/i), {
      target: { value: "password123" },
    });

    fireEvent.click(screen.getByRole("button", { name: /登録/i }));

    await waitFor(() => {
      expect(mockRegister).toHaveBeenCalledWith(
        "田中 太郎",
        "tanaka@example.com",
        "password123"
      );
      expect(mockPush).toHaveBeenCalledWith("/login");
    });
  });

  // テストケース2: 異常系（メール重複）
  it("異常系: メール重複でエラーメッセージを表示", async () => {
    mockRegister.mockRejectedValue(
      new Error("このメールアドレスは既に登録されています")
    );

    render(<RegisterPage />);

    fireEvent.change(screen.getByPlaceholderText(/名前/i), {
      target: { value: "新規 ユーザー" },
    });
    fireEvent.change(screen.getByPlaceholderText(/メールアドレス/i), {
      target: { value: "existing@example.com" },
    });
    fireEvent.change(screen.getByPlaceholderText(/パスワード/i), {
      target: { value: "password123" },
    });

    fireEvent.click(screen.getByRole("button", { name: /登録/i }));

    await waitFor(() => {
      expect(screen.getByText(/既に登録されています/i)).toBeInTheDocument();
    });
  });

  // テストケース3: UI - ローディング状態
  it("UI: 送信中はボタンがdisabledになり、テキストが変わる", async () => {
    mockRegister.mockImplementation(
      () => new Promise((resolve) => setTimeout(resolve, 1000))
    );

    render(<RegisterPage />);

    const button = screen.getByRole("button", { name: /登録/i });

    fireEvent.change(screen.getByPlaceholderText(/名前/i), {
      target: { value: "テスト ユーザー" },
    });
    fireEvent.change(screen.getByPlaceholderText(/メールアドレス/i), {
      target: { value: "test@example.com" },
    });
    fireEvent.change(screen.getByPlaceholderText(/パスワード/i), {
      target: { value: "password123" },
    });

    fireEvent.click(button);

    expect(button).toBeDisabled();
    expect(button.textContent).toMatch(/登録中/i);
  });

  // テストケース6: HTML属性
  it("HTML: メールフィールドはtype=emailである", () => {
    render(<RegisterPage />);
    const emailInput = screen.getByPlaceholderText(/メールアドレス/i) as HTMLInputElement;
    expect(emailInput.type).toBe("email");
  });
});
