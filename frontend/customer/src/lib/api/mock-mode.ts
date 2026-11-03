// true = build from the typed fixtures in ./mock-data.ts instead of the backend's public API
// (handy for UI work without a running backend). Production builds must use the API.
// Kept in its own tiny module so the period table's script can read it without bundling the fixtures.
export const USE_MOCK_DATA = false;
