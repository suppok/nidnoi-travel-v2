import { ErrorCodes } from '../../constants/errorCodes.ts';
import { Permissions } from '../../constants/permissions.ts';
import type { AISetting } from '../../types/aiSetting.ts';
import { asRecord, authorize, MockFail, type MockHandler, ok, str, validation } from './http.ts';

// Mock of the AI settings. Like the backend, the key itself is never returned. In the mock a key
// containing "invalid" fails the connection test, anything else that looks like sk-ant-… passes.
const DB_KEY = 'nidnoi.mock.aiSettings.v1';
const KEY_PATTERN = /^sk-ant-[A-Za-z0-9_-]{10,300}$/;

export const AI_MODELS = [
  { id: 'claude-sonnet-5-5', label: 'Claude Sonnet 5.5 (แนะนำ)' },
  { id: 'claude-opus-5-5', label: 'Claude Opus 5.5 (แม่นยำกว่า ราคาสูงกว่า)' },
  { id: 'claude-haiku-5-5', label: 'Claude Haiku 5.5 (เร็วและถูกที่สุด)' },
];
const MAX_PDF_LIMIT = 40;

type Stored = { enabled: boolean; model: string; maxPdfSizeMb: number; apiKey: string; updateDate: string | null; updateBy: string };

function load(): Stored {
  try {
    const raw = localStorage.getItem(DB_KEY);
    if (raw) return JSON.parse(raw) as Stored;
  } catch {
    // fall through
  }
  return { enabled: false, model: AI_MODELS[0].id, maxPdfSizeMb: 40, apiKey: '', updateDate: null, updateBy: '' };
}

let stored = load();

/** The settings as GET /ai-settings returns them; also read by the mock extraction. */
export function mockAISetting(): AISetting {
  return {
    enabled: stored.enabled,
    model: stored.model,
    maxPdfSizeMb: stored.maxPdfSizeMb,
    hasApiKey: stored.apiKey !== '',
    apiKeyLast4: stored.apiKey.slice(-4),
    keyUnreadable: false,
    encryptionConfigured: true,
    ready: stored.enabled && stored.apiKey !== '',
    maxPdfSizeMbLimit: MAX_PDF_LIMIT,
    models: AI_MODELS,
    updateDate: stored.updateDate,
    updateBy: stored.updateBy,
  };
}

function testKey(key: string) {
  if (key.includes('invalid')) throw new MockFail(400, ErrorCodes.AIInvalidApiKey, 'The Anthropic API key was rejected');
}

export const aiSettingRoutes: Record<string, MockHandler> = {
  'GET /ai-settings': (req) => {
    authorize(req, 'GET /ai-settings', Permissions.AISettingGet);
    return ok(mockAISetting());
  },
  'PUT /ai-settings': (req) => {
    const user = authorize(req, 'PUT /ai-settings', Permissions.AISettingUpdate);
    const body = asRecord(req.body);
    const model = str(body.model);
    const maxPdfSizeMb = Number(body.maxPdfSizeMb);
    const apiKey = str(body.apiKey).trim();
    const clearApiKey = body.clearApiKey === true;
    const enabled = body.enabled === true;
    const fields: Record<string, string> = {};
    if (!AI_MODELS.some((m) => m.id === model)) fields.model = 'must be a valid value';
    if (!Number.isInteger(maxPdfSizeMb) || maxPdfSizeMb < 1 || maxPdfSizeMb > MAX_PDF_LIMIT) fields.maxPdfSizeMb = `must be between 1 and ${MAX_PDF_LIMIT}`;
    if (apiKey && !KEY_PATTERN.test(apiKey)) fields.apiKey = 'must be an Anthropic API key (sk-ant-…)';
    if (apiKey && clearApiKey) fields.apiKey = 'cannot set and clear the key at once';
    const keyAfter = apiKey !== '' || (stored.apiKey !== '' && !clearApiKey);
    if (enabled && !keyAfter) fields.enabled = 'an API key is required to turn AI on';
    if (Object.keys(fields).length) throw validation(fields);

    stored = {
      enabled,
      model,
      maxPdfSizeMb,
      apiKey: apiKey || (clearApiKey ? '' : stored.apiKey),
      updateDate: new Date().toISOString(),
      updateBy: user.name,
    };
    try {
      localStorage.setItem(DB_KEY, JSON.stringify(stored));
    } catch {
      // memory only
    }
    return ok(mockAISetting());
  },
  'POST /ai-settings/test': (req) => {
    authorize(req, 'POST /ai-settings/test', Permissions.AISettingUpdate);
    const body = asRecord(req.body);
    const apiKey = str(body.apiKey).trim();
    const model = str(body.model) || stored.model;
    if (apiKey && !KEY_PATTERN.test(apiKey)) throw validation({ apiKey: 'must be an Anthropic API key (sk-ant-…)' });
    const key = apiKey || stored.apiKey;
    if (!key) throw validation({ apiKey: 'enter an API key or save one first' });
    testKey(key);
    return ok({ ok: true, model });
  },
};
