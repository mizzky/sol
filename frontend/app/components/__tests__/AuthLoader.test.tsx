import React from "react";
import { render } from "@testing-library/react";
import AuthLoader from "../AuthLoader";

const mockLoad = jest.fn();
const mockSyncCart = jest.fn();
const mockResetCart = jest.fn();
const authState = { isAuthenticated: false };

jest.mock("../../../store/useAuthStore", () => ({
  __esModule: true,
  default: {
    getState: () => ({ loadFromStorage: mockLoad, isAuthenticated: authState.isAuthenticated }),
  },
}));

jest.mock("../../../store/useCartStore", () => ({
  __esModule: true,
  default: {
    getState: () => ({ syncCart: mockSyncCart, resetCart: mockResetCart }),
  },
}));

describe("AuthLoader", () => {
  beforeEach(() => {
    jest.resetAllMocks();
    authState.isAuthenticated = false;
  });

  it("calls loadFromStorage on mount", () => {
    render(<AuthLoader />);
    expect(mockLoad).toHaveBeenCalled();
  });

  it("syncs cart when auth is restored", async () => {
    mockLoad.mockImplementation(async () => {
      authState.isAuthenticated = true;
    });

    render(<AuthLoader />);

    await Promise.resolve();
    expect(mockSyncCart).toHaveBeenCalled();
    expect(mockResetCart).not.toHaveBeenCalled();
  });

  it("resets cart when auth is not available", async () => {
    mockLoad.mockResolvedValue(undefined);

    render(<AuthLoader />);

    await Promise.resolve();
    expect(mockResetCart).toHaveBeenCalled();
  });

  it("resets cart when syncCart fails", async () => {
    mockLoad.mockImplementation(async () => {
      authState.isAuthenticated = true;
    });
    mockSyncCart.mockRejectedValue(new Error("sync failed"));

    render(<AuthLoader />);

    await Promise.resolve();
    await Promise.resolve();
    expect(mockSyncCart).toHaveBeenCalled();
    expect(mockResetCart).toHaveBeenCalled();
  });
});
