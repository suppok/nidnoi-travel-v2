import type { SiteBuildOutcome, SiteBuildTrigger } from '../constants/siteBuilds.ts';
import type { Paging } from './api.ts';

// A customer-site build the backend triggered (manual button, content edits, daily). Git-push builds
// are not included. outcome = whether Cloudflare accepted the trigger, not the build result.
export type SiteBuild = {
  id: string;
  trigger: SiteBuildTrigger;
  outcome: SiteBuildOutcome;
  paths: string[];
  attempts: number;
  error: string;
  triggeredBy: { id: string; name: string } | null;
  createDate: string;
};

export type SiteBuildListRequest = Paging & { trigger?: SiteBuildTrigger };

export type SiteRebuildResponse = { triggered: boolean };
