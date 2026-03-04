import { render } from "@testing-library/react";
import CartPage from "../page";
import React from "react";

describe("CartPage", () => {
  it("renders cart page", () => {
    const { getByRole } = render(<CartPage />);
    expect(getByRole("heading", { name: /カート/ })).toBeTruthy();
  });
});
