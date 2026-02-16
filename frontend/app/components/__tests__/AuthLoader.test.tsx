import React from "react";
import { render } from "@testing-library/react";
import AuthLoader from "../AuthLoader";

const mockLoad = jest.fn();

jest.mock("../../../store/useAuthStore", () => ({
  __esModule: true,
  default: {
    getState: () => ({ loadFromStorage: mockLoad }),
  },
}));

describe("AuthLoader", () => {
  beforeEach(() => {
    jest.resetAllMocks();
  });

  it("calls loadFromStorage on mount", () => {
    render(<AuthLoader />);
    expect(mockLoad).toHaveBeenCalled();
  });
});
