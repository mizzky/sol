import React from "react";
import { render, screen } from "@testing-library/react";
import Home from "../page";
import * as api from "../../lib/api";
import useAuthStore from "../../store/useAuthStore";

jest.mock("../../lib/api");

describe("Top Page", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    // loadFromStorage が実際にネットワーク呼び出しして state を更新しないようモック化
    useAuthStore.setState({
      token: null,
      user: null,
      loadFromStorage: jest.fn(),
    } as any);
    localStorage.clear();
    (api.getProducts as jest.Mock).mockResolvedValue([]);
  });

  it("トップページは一般向けの商品一覧のみで、管理用登録フォームが存在しない（TDD: Red）", async () => {
    render(<Home />);

    // 非同期で読み込みが終わるまで待つ
    await screen.findByText(/本日のおすすめ/);

    // 管理者用の見出しや「追加」ボタン等がないことを期待（現状は存在するので失敗するはず）
    expect(screen.queryByText(/新規商品登録/)).not.toBeInTheDocument();
    expect(screen.queryByText(/追加/)).not.toBeInTheDocument();
  });
});
