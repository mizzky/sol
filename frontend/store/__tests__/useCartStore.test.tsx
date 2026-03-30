import { useCartStore } from "../useCartStore";

const mockGetCart = jest.fn();
const mockAddToCart = jest.fn();
const mockUpdateCartItem = jest.fn();
const mockRemoveFromCart = jest.fn();
const mockClearCart = jest.fn();

jest.mock("../../lib/api", () => ({
  getCart: (...args: unknown[]) => mockGetCart(...args),
  addToCart: (...args: unknown[]) => mockAddToCart(...args),
  updateCartItem: (...args: unknown[]) => mockUpdateCartItem(...args),
  removeFromCart: (...args: unknown[]) => mockRemoveFromCart(...args),
  clearCart: (...args: unknown[]) => mockClearCart(...args),
}));

const sampleItems = [
  {
    id: 10,
    cart_id: 1,
    product_id: 100,
    quantity: 2,
    price: 450,
    product_name: "Blend",
  },
  {
    id: 11,
    cart_id: 1,
    product_id: 101,
    quantity: 1,
    price: 600,
    product_name: "Dark Roast",
  },
];

describe("useCartStore", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    useCartStore.getState().resetCart();
  });

  it("initial state is empty", () => {
    const state = useCartStore.getState();
    expect(state.items).toEqual([]);
    expect(state.totalQuantity).toBe(0);
    expect(state.totalPrice).toBe(0);
  });

  it("setCart computes totals", () => {
    useCartStore.getState().setCart(sampleItems);

    const state = useCartStore.getState();
    expect(state.items).toEqual(sampleItems);
    expect(state.totalQuantity).toBe(3);
    expect(state.totalPrice).toBe(1500);
  });

  it("syncCart loads items from API", async () => {
    mockGetCart.mockResolvedValue(sampleItems);

    await useCartStore.getState().syncCart();

    const state = useCartStore.getState();
    expect(mockGetCart).toHaveBeenCalledTimes(1);
    expect(state.items).toEqual(sampleItems);
    expect(state.totalQuantity).toBe(3);
    expect(state.totalPrice).toBe(1500);
  });

  it("addItem appends a new cart item", async () => {
    mockAddToCart.mockResolvedValue(sampleItems[0]);

    await useCartStore.getState().addItem(100, 2);

    const state = useCartStore.getState();
    expect(mockAddToCart).toHaveBeenCalledWith(100, 2);
    expect(state.items).toEqual([sampleItems[0]]);
    expect(state.totalQuantity).toBe(2);
    expect(state.totalPrice).toBe(900);
  });

  it("updateItem replaces an existing cart item", async () => {
    useCartStore.getState().setCart(sampleItems);
    mockUpdateCartItem.mockResolvedValue({ ...sampleItems[0], quantity: 4 });

    await useCartStore.getState().updateItem(10, 4);

    const state = useCartStore.getState();
    expect(mockUpdateCartItem).toHaveBeenCalledWith(10, 4);
    expect(state.items[0].quantity).toBe(4);
    expect(state.totalQuantity).toBe(5);
    expect(state.totalPrice).toBe(2400);
  });

  it("clearCart resets local state after API success", async () => {
    useCartStore.getState().setCart(sampleItems);
    mockClearCart.mockResolvedValue(undefined);

    await useCartStore.getState().clearCart();

    const state = useCartStore.getState();
    expect(mockClearCart).toHaveBeenCalledTimes(1);
    expect(state.items).toEqual([]);
    expect(state.totalQuantity).toBe(0);
    expect(state.totalPrice).toBe(0);
  });
});
