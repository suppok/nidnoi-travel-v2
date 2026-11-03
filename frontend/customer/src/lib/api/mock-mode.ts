// TODO(public-api): set to false once the backend ships the public endpoints
// (/tour-programs/public…, /countries/public, /cities/public). Pages and islands need no other change.
// Kept in its own tiny module so islands can read it without bundling the fixtures.
export const USE_MOCK_DATA = true;
