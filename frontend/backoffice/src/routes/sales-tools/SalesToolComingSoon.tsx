import { Result } from 'antd';
import PageHeader from '../../components/PageHeader.tsx';
import type { SalesTool } from '../../features/sales-tool/tools.tsx';

/** Placeholder for a sales tool that is listed in the sidebar but not built yet. */
export default function SalesToolComingSoon({ tool }: { tool: SalesTool }) {
  return (
    <>
      <PageHeader title={tool.label} subtitle={tool.description} />
      <Result icon={<span className="text-5xl text-ink-muted">{tool.icon}</span>} title="เร็วๆ นี้" subTitle="เครื่องมือนี้กำลังพัฒนา" />
    </>
  );
}
