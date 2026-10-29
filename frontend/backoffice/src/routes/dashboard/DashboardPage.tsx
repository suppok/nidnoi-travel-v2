import { Card, Empty } from 'antd';
import { useState } from 'react';
import { useAuth } from '../../auth/useAuth.ts';
import PageHeader from '../../components/PageHeader.tsx';
import { formatLongDate } from '../../lib/format.ts';

export default function DashboardPage() {
  const { user } = useAuth();
  const [today] = useState(() => new Date());

  return (
    <>
      <PageHeader title={`สวัสดี ${user?.name ?? ''}`} subtitle={formatLongDate(today)} />
      <Card>
        <Empty description="ภาพรวมทัวร์และการจองจะแสดงที่นี่" />
      </Card>
    </>
  );
}
