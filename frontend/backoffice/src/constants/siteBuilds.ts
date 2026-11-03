// MUST match backend/pkg/constants/sitebuild (and the site_build CHECK constraints).
export const SiteBuildTriggers = {
  Manual: 'manual',
  Auto: 'auto',
  Daily: 'daily',
} as const;

export type SiteBuildTrigger = (typeof SiteBuildTriggers)[keyof typeof SiteBuildTriggers];

export const SiteBuildOutcomes = {
  Triggered: 'triggered',
  Failed: 'failed',
  Skipped: 'skipped',
} as const;

export type SiteBuildOutcome = (typeof SiteBuildOutcomes)[keyof typeof SiteBuildOutcomes];

export const triggerInfo: Record<SiteBuildTrigger, { label: string; color: string }> = {
  manual: { label: 'กดอัปเดตเอง', color: 'blue' },
  auto: { label: 'อัตโนมัติจากการแก้ไข', color: 'cyan' },
  daily: { label: 'รายวัน', color: 'default' },
};

export const outcomeInfo: Record<SiteBuildOutcome, { label: string; color: string }> = {
  triggered: { label: 'สั่งสร้างแล้ว', color: 'green' },
  failed: { label: 'สั่งไม่สำเร็จ', color: 'red' },
  skipped: { label: 'ข้าม (ยังไม่ตั้งค่า)', color: 'orange' },
};
