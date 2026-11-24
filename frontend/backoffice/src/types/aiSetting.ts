// GET / PUT /ai-settings — the Claude API key, model and on/off for AI features (admins).
export type AIModelOption = { id: string; label: string };

export type AISetting = {
  enabled: boolean;
  model: string;
  maxPdfSizeMb: number;
  hasApiKey: boolean; // the key itself is never returned
  apiKeyLast4: string; // '' when no key
  keyUnreadable: boolean; // stored key can't be decrypted (server key changed): enter it again
  encryptionConfigured: boolean; // false = AI_SETTINGS_ENCRYPTION_KEY missing on the server
  ready: boolean; // enabled + usable key: the AI prefill is offered to staff
  maxPdfSizeMbLimit: number;
  models: AIModelOption[]; // default first
  updateDate: string | null;
  updateBy: string;
};

// apiKey omitted / null / '' keeps the stored key; clearApiKey removes it.
export type AISettingUpdateRequest = {
  enabled: boolean;
  model: string;
  maxPdfSizeMb: number;
  apiKey?: string | null;
  clearApiKey?: boolean;
};

// POST /ai-settings/test — apiKey '' = the stored key; model '' = the stored model.
export type AISettingTestRequest = { apiKey: string; model: string };
export type AISettingTestResponse = { ok: boolean; model: string };
