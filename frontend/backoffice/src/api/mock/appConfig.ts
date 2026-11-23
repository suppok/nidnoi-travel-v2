import { Permissions } from '../../constants/permissions.ts';
import type { EmailContactCard } from '../../types/appConfig.ts';
import { asRecord, authorize, type MockHandler, ok, str, validation } from './http.ts';

// Mock of the email contact card setting. The built-in card lives in the backend binary, so the mock
// has no preview of it (defaultImage '').
const DB_KEY = 'nidnoi.mock.emailContactCard.v1';

function load(): EmailContactCard {
  try {
    const raw = localStorage.getItem(DB_KEY);
    if (raw) return JSON.parse(raw) as EmailContactCard;
  } catch {
    // fall through
  }
  return { imageUrl: '', defaultImage: '', updateDate: null, updateBy: '' };
}

let card = load();

export const appConfigRoutes: Record<string, MockHandler> = {
  'GET /app-configs/email-contact-card': (req) => {
    authorize(req, 'GET /app-configs/email-contact-card', Permissions.AppConfigEmailContactCard);
    return ok(card);
  },
  'PUT /app-configs/email-contact-card': (req) => {
    const user = authorize(req, 'PUT /app-configs/email-contact-card', Permissions.AppConfigEmailContactCard);
    const imageUrl = str(asRecord(req.body).imageUrl);
    if (imageUrl && !imageUrl.startsWith('blob:')) throw validation({ imageUrl: 'must be an image uploaded with purpose email-contact-card' });
    card = { ...card, imageUrl, updateDate: new Date().toISOString(), updateBy: user.name };
    try {
      localStorage.setItem(DB_KEY, JSON.stringify(card));
    } catch {
      // memory only
    }
    return ok(card);
  },
};
