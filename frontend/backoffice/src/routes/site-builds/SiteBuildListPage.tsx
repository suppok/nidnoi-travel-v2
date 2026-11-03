import { ReloadOutlined } from '@ant-design/icons';
import { Alert, Button, Card, Segmented, Table, type TableProps, Tag, Tooltip, Typography } from 'antd';
import { useEffect } from 'react';
import { useSearchParams } from 'react-router';
import { siteApi } from '../../api/site.ts';
import PageHeader from '../../components/PageHeader.tsx';
import { outcomeInfo, type SiteBuildTrigger, SiteBuildTriggers, triggerInfo } from '../../constants/siteBuilds.ts';
import SiteRebuildButton, { SITE_REBUILT_EVENT } from '../../features/site/components/SiteRebuildButton.tsx';
import { useListQuery } from '../../hooks/useListQuery.ts';
import { errorMessage } from '../../lib/errors.ts';
import { formatDateTime } from '../../lib/format.ts';
import type { SiteBuild } from '../../types/siteBuild.ts';

type TriggerFilter = SiteBuildTrigger | 'all';
const DEFAULT_PAGE_SIZE = 20;
const TRIGGERS: string[] = Object.values(SiteBuildTriggers);
const MAX_PAGES_SHOWN = 4;

function useFilters() {
  const [params, setParams] = useSearchParams();
  const t = params.get('trigger') ?? '';
  const filters = {
    trigger: (TRIGGERS.includes(t) ? t : 'all') as TriggerFilter,
    page: Math.max(1, Number(params.get('page')) || 1),
    pageSize: Number(params.get('pageSize')) || DEFAULT_PAGE_SIZE,
  };

  // Any filter change goes back to page 1 unless the page itself is being set.
  const update = (patch: Partial<Record<'trigger' | 'page' | 'pageSize', string | undefined>>) => {
    setParams(
      (prev) => {
        const next = new URLSearchParams(prev);
        if (!('page' in patch)) next.delete('page');
        for (const [key, value] of Object.entries(patch)) {
          if (value) next.set(key, value);
          else next.delete(key);
        }
        return next;
      },
      { replace: true },
    );
  };

  return { filters, update };
}

// "/" → หน้าแรก, "/tour/NIDN000001/" → NIDN000001, "/tours/japan/tokyo/" → japan/tokyo
function pageLabel(path: string): string {
  if (path === '/') return 'หน้าแรก';
  const tour = /^\/tour\/([^/]+)\/$/.exec(path);
  if (tour) return tour[1];
  return path.replace(/^\/tours\//, '').replace(/\/$/, '');
}

function Pages({ build }: { build: SiteBuild }) {
  if (build.trigger === 'daily') return <span className="text-ink-muted">ทั้งเว็บ (ตัดรอบเดินทางที่ผ่านไปแล้ว)</span>;
  if (build.paths.length === 0) return <span className="text-ink-muted">ทั้งเว็บ</span>;
  const shown = build.paths.slice(0, MAX_PAGES_SHOWN);
  const rest = build.paths.slice(MAX_PAGES_SHOWN);
  return (
    <div className="flex flex-wrap gap-1">
      {shown.map((p) => (
        <Tag key={p} className="m-0 font-mono text-xs">
          {pageLabel(p)}
        </Tag>
      ))}
      {rest.length > 0 && (
        <Tooltip title={rest.map(pageLabel).join(', ')}>
          <Tag className="m-0 cursor-help">+{rest.length}</Tag>
        </Tooltip>
      )}
    </div>
  );
}

export default function SiteBuildListPage() {
  const { filters, update } = useFilters();
  const { data, loading, error, reload } = useListQuery(siteApi.listBuilds, {
    trigger: filters.trigger === 'all' ? undefined : filters.trigger,
    page: filters.page,
    pageSize: filters.pageSize,
  });

  // A manual rebuild (header button or this page) shows up without a manual refresh.
  useEffect(() => {
    window.addEventListener(SITE_REBUILT_EVENT, reload);
    return () => window.removeEventListener(SITE_REBUILT_EVENT, reload);
  }, [reload]);

  const columns: TableProps<SiteBuild>['columns'] = [
    { title: 'เวลา', dataIndex: 'createDate', width: 180, render: (v: string) => formatDateTime(v) },
    {
      title: 'ประเภท',
      dataIndex: 'trigger',
      width: 190,
      render: (t: SiteBuild['trigger']) => (
        <Tag color={triggerInfo[t].color} className="m-0">
          {triggerInfo[t].label}
        </Tag>
      ),
    },
    {
      title: 'ผลการสั่ง',
      key: 'outcome',
      width: 200,
      render: (_, b) => (
        <div className="flex flex-col items-start gap-1">
          <Tag color={outcomeInfo[b.outcome].color} className="m-0">
            {outcomeInfo[b.outcome].label}
          </Tag>
          {b.attempts > 1 && <span className="text-xs text-ink-muted">ลอง {b.attempts} ครั้ง</span>}
          {b.error && (
            <Typography.Text type="danger" className="text-xs" ellipsis={{ tooltip: b.error }} style={{ maxWidth: 180 }}>
              {b.error}
            </Typography.Text>
          )}
        </div>
      ),
    },
    { title: 'หน้าที่เกี่ยวข้อง', key: 'paths', render: (_, b) => <Pages build={b} /> },
    { title: 'โดย', key: 'by', width: 160, render: (_, b) => b.triggeredBy?.name ?? <span className="text-ink-muted">ระบบ</span> },
  ];

  return (
    <>
      <PageHeader
        title="ประวัติอัปเดตหน้าเว็บ"
        subtitle="การสร้างหน้าเว็บลูกค้าใหม่ที่ระบบสั่ง: กดอัปเดตเอง, อัตโนมัติหลังแก้ข้อมูล และรายวัน (ไม่รวมการ deploy โค้ด)"
        actions={
          <>
            <Button icon={<ReloadOutlined />} onClick={reload}>
              รีเฟรช
            </Button>
            <SiteRebuildButton type="primary" />
          </>
        }
      />

      <Card>
        <div className="mb-4 flex flex-wrap items-center gap-3">
          <Segmented<TriggerFilter>
            value={filters.trigger}
            options={[
              { value: 'all', label: 'ทั้งหมด' },
              { value: 'manual', label: triggerInfo.manual.label },
              { value: 'auto', label: triggerInfo.auto.label },
              { value: 'daily', label: triggerInfo.daily.label },
            ]}
            onChange={(trigger) => update({ trigger: trigger === 'all' ? undefined : trigger })}
          />
        </div>

        <Alert
          type="info"
          showIcon
          className="mb-4"
          title="“สั่งสร้างแล้ว” หมายถึง Cloudflare รับคำสั่งแล้ว การสร้างใช้เวลา 2–5 นาที ถ้าหน้าเว็บไม่เปลี่ยน ให้ดูผลการ build ใน Cloudflare"
        />

        {error !== null && <Alert type="error" showIcon title={errorMessage(error)} className="mb-4" />}

        <Table<SiteBuild>
          rowKey="id"
          columns={columns}
          dataSource={data?.items ?? []}
          loading={loading}
          pagination={{
            current: filters.page,
            pageSize: filters.pageSize,
            total: data?.total ?? 0,
            showSizeChanger: true,
            showTotal: (total) => `ทั้งหมด ${total} ครั้ง`,
            onChange: (page, pageSize) =>
              update({ page: page > 1 ? String(page) : undefined, pageSize: pageSize !== DEFAULT_PAGE_SIZE ? String(pageSize) : undefined }),
          }}
          scroll={{ x: 860 }}
        />
      </Card>
    </>
  );
}
