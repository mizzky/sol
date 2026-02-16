import '@testing-library/jest-dom';

// Provide a minimal localStorage mock for Node/Jest environment
if (typeof globalThis.localStorage === 'undefined') {
	let store: Record<string, string> = {};
	const localStorageMock: Storage = {
		getItem: (key: string) => (key in store ? store[key] : null),
		setItem: (key: string, value: string) => {
			store[key] = String(value);
		},
		removeItem: (key: string) => {
			delete store[key];
		},
		clear: () => {
			store = {};
		},
		key: (index: number) => Object.keys(store)[index] ?? null,
		get length() {
			return Object.keys(store).length;
		},
	};
	Object.defineProperty(globalThis, 'localStorage', {
		value: localStorageMock,
		writable: true,
	});
}