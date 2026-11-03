import { Permissions } from '../../constants/permissions.ts';
import { SiteBuildTriggers } from '../../constants/siteBuilds.ts';
import type { ListResponse } from '../../types/api.ts';
import type { SiteBuild } from '../../types/siteBuild.ts';
import { users } from './db.ts';
import { asRecord, authorize, type MockHandler, ok, str, validation } from './http.ts';

// Mock-only build history. The real backend records every deploy-hook trigger in site_build.
const DB_KEY = 'nidnoi.mock.siteBuilds.v1';

function seed(): SiteBuild[] {
  const at = (minutesAgo: number) => new Date(Date.now() - minutesAgo * 60_000).toISOString();
  return [
    { id: crypto.randomUUID(), trigger: 'auto', outcome: 'triggered', paths: ['/', '/tour/NIDN000001/', '/tours/japan/'], attempts: 1, error: '', triggeredBy: null, createDate: at(42) },
    { id: crypto.randomUUID(), trigger: 'auto', outcome: 'failed', paths: ['/', '/tour/NIDN000002/'], attempts: 4, error: 'deploy hook: unexpected status 500', triggeredBy: null, createDate: at(180) },
    { id: crypto.randomUUID(), trigger: 'daily', outcome: 'triggered', paths: [], attempts: 1, error: '', triggeredBy: null, createDate: at(60 * 20) },
  ];
}

function load(): SiteBuild[] {
  try {
    const raw = localStorage.getItem(DB_KEY);
    if (raw) return JSON.parse(raw) as SiteBuild[];
  } catch {
    // fall through to seed
  }
  return seed();
}

const builds: SiteBuild[] = load();

function save() {
  try {
    localStorage.setItem(DB_KEY, JSON.stringify(builds));
  } catch {
    // storage unavailable: keep in memory only
  }
}

const TRIGGERS: string[] = Object.values(SiteBuildTriggers);

export const siteRoutes: Record<string, MockHandler> = {
  // The mock always "accepts" the trigger (the real one returns SITE_REBUILD_UNAVAILABLE without a hook).
  'POST /site/rebuild': (req) => {
    const user = authorize(req, 'POST /site/rebuild', Permissions.SiteRebuild);
    builds.unshift({
      id: crypto.randomUUID(),
      trigger: 'manual',
      outcome: 'triggered',
      paths: [],
      attempts: 1,
      error: '',
      triggeredBy: { id: user.id, name: users.find((u) => u.id === user.id)?.name ?? user.name },
      createDate: new Date().toISOString(),
    });
    save();
    return ok({ triggered: true });
  },

  'POST /site-builds/list': (req) => {
    authorize(req, 'POST /site-builds/list', Permissions.SiteBuildList);
    const body = asRecord(req.body);
    const trigger = str(body.trigger);
    if (trigger && !TRIGGERS.includes(trigger)) throw validation({ trigger: 'must be a valid value' });
    const page = Math.max(1, Number(body.page) || 1);
    const pageSize = Number(body.pageSize) || 20;
    if (pageSize > 100) throw validation({ pageSize: 'must be no greater than 100' });

    const rows = builds.filter((b) => !trigger || b.trigger === trigger).sort((a, b) => b.createDate.localeCompare(a.createDate));
    return ok({ items: rows.slice((page - 1) * pageSize, page * pageSize), total: rows.length, page, pageSize } satisfies ListResponse<SiteBuild>);
  },
};
