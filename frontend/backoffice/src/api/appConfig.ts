import type { EmailContactCard, EmailContactCardUpdateRequest } from '../types/appConfig.ts';
import { request } from './client.ts';

export const appConfigApi = {
  getEmailContactCard: () => request<EmailContactCard>('GET', '/app-configs/email-contact-card'),
  updateEmailContactCard: (body: EmailContactCardUpdateRequest) => request<EmailContactCard>('PUT', '/app-configs/email-contact-card', body),
};
