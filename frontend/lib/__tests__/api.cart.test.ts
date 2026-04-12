import { addToCart, getCart, updateCartItem } from "../api";

describe("cart api", () => {
  beforeEach(() => {
    localStorage.clear();
    jest.resetAllMocks();
  });

  it("getCart returns cart items array", async () => {
    global.fetch = jest.fn(() =>
      Promise.resolve({
        ok: true,
        status: 200,
        json: async () => ({
          items: [
            { id: 1, cart_id: 1, product_id: 10, quantity: 2, price: 450, product_name: "Blend" },
          ],
        }),
      }) as unknown as Response,
    ) as unknown as typeof global.fetch;

    await expect(getCart()).resolves.toEqual([
      { id: 1, cart_id: 1, product_id: 10, quantity: 2, price: 450, product_name: "Blend" },
    ]);
  });

  it("addToCart unwraps item response", async () => {
    global.fetch = jest.fn(() =>
      Promise.resolve({
        ok: true,
        status: 201,
        json: async () => ({
          item: { id: 1, cart_id: 1, product_id: 10, quantity: 2, price: 450 },
        }),
      }) as unknown as Response,
    ) as unknown as typeof global.fetch;

    await expect(addToCart(10, 2)).resolves.toEqual({
      id: 1,
      cart_id: 1,
      product_id: 10,
      quantity: 2,
      price: 450,
    });
  });

  it("updateCartItem unwraps item response", async () => {
    global.fetch = jest.fn(() =>
      Promise.resolve({
        ok: true,
        status: 200,
        json: async () => ({
          item: { id: 1, cart_id: 1, product_id: 10, quantity: 3, price: 450 },
        }),
      }) as unknown as Response,
    ) as unknown as typeof global.fetch;

    await expect(updateCartItem(1, 3)).resolves.toEqual({
      id: 1,
      cart_id: 1,
      product_id: 10,
      quantity: 3,
      price: 450,
    });
  });
});