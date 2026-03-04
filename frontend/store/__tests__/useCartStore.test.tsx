import { renderHook, act } from "@testing-library/react-hooks";
import { useCartStore } from "../useCartStore";

describe("useCartStore", () => {
  it("initial state is empty", () => {
    const state = useCartStore.getState();
    expect(state.items).toEqual([]);
    expect(state.totalQuantity).toBe(0);
    expect(state.totalPrice).toBe(0);
  });
});
