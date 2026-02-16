// frontend/lib/__tests__/api.login.test.ts
import { login } from '../../lib/api';

describe('login()', () => {
  const OLD_FETCH = global.fetch;

  afterEach(() => {
    global.fetch = OLD_FETCH;
    jest.resetAllMocks();
  });

  const createFetchResponse = (body: unknown, ok = true, status = 200) =>
    Promise.resolve({
      ok,
      status,
      json: async () => body,
    } as unknown as Response);

  it('resolves with token and user on 200', async () => {
    const mockResponse = {
      token: 'mock-token',
      user: { id: 1, name: 'User', email: 'user@example.com' },
    };
    global.fetch = jest.fn(() => createFetchResponse(mockResponse, true, 200)) as unknown as typeof global.fetch;

    const res = await login('user@example.com', 'password123');
    expect(res).toEqual(mockResponse);
    expect(global.fetch).toHaveBeenCalled();
  });

  it('rejects/throws on 400 with error body', async () => {
    const mockError = { error: 'Invalid credentials' };
    global.fetch = jest.fn(() => createFetchResponse(mockError, false, 400));

    await expect(login('bad@example.com', 'wrong')).rejects.toBeDefined();
    expect(global.fetch).toHaveBeenCalled();
  });
});