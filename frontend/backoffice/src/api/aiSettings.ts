import type { AISetting, AISettingTestRequest, AISettingTestResponse, AISettingUpdateRequest } from '../types/aiSetting.ts';
import { request } from './client.ts';

export const aiSettingsApi = {
  get: () => request<AISetting>('GET', '/ai-settings'),
  update: (body: AISettingUpdateRequest) => request<AISetting>('PUT', '/ai-settings', body),
  test: (body: AISettingTestRequest) => request<AISettingTestResponse>('POST', '/ai-settings/test', body),
};
