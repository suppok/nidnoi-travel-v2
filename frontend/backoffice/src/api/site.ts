import type { ListResponse } from '../types/api.ts';
import type { SiteBuild, SiteBuildListRequest, SiteRebuildResponse } from '../types/siteBuild.ts';
import { request } from './client.ts';

export const siteApi = {
  rebuild: () => request<SiteRebuildResponse>('POST', '/site/rebuild'),
  listBuilds: (body: SiteBuildListRequest) => request<ListResponse<SiteBuild>>('POST', '/site-builds/list', body),
};
