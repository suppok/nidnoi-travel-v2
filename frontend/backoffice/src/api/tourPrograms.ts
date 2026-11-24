import type { AIExtractConfig, AIExtractJob, AIExtractStart } from '../types/aiExtract.ts';
import type { ListResponse } from '../types/api.ts';
import type {
  PeriodCreateRequest,
  PeriodListRequest,
  PeriodUpdateRequest,
  StatusRequest,
  TourProgram,
  TourProgramCreateRequest,
  TourProgramListItem,
  TourProgramListRequest,
  TourProgramPeriod,
  TourProgramUpdateRequest,
} from '../types/tourProgram.ts';
import { request } from './client.ts';

export const tourProgramsApi = {
  list: (body: TourProgramListRequest) => request<ListResponse<TourProgramListItem>>('POST', '/tour-programs/list', body),
  get: (id: string) => request<TourProgram>('GET', `/tour-programs/${encodeURIComponent(id)}`),
  create: (body: TourProgramCreateRequest) => request<TourProgram>('POST', '/tour-programs', body),
  update: (body: TourProgramUpdateRequest) => request<TourProgram>('PUT', '/tour-programs', body),
  updateStatus: (body: StatusRequest) => request<null>('PUT', '/tour-programs/status', body),
};

// AI prefill of the create form from a PDF. Nothing is saved: staff review and save the form.
export const aiExtractApi = {
  config: () => request<AIExtractConfig>('GET', '/tour-programs/ai-extract/config'),
  start: (file: File) => {
    const body = new FormData();
    body.append('file', file);
    return request<AIExtractStart>('POST', '/tour-programs/ai-extract', body);
  },
  result: (jobId: string) => request<AIExtractJob>('GET', `/tour-programs/ai-extract/${encodeURIComponent(jobId)}`),
};

export const periodsApi = {
  list: (body: PeriodListRequest) => request<{ items: TourProgramPeriod[] }>('POST', '/tour-program-periods/list', body),
  create: (body: PeriodCreateRequest) => request<TourProgramPeriod>('POST', '/tour-program-periods', body),
  update: (body: PeriodUpdateRequest) => request<TourProgramPeriod>('PUT', '/tour-program-periods', body),
  updateStatus: (body: StatusRequest) => request<null>('PUT', '/tour-program-periods/status', body),
};
