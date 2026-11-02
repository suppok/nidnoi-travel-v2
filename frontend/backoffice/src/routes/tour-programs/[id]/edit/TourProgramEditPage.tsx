import { ArrowLeftOutlined, StopOutlined, UndoOutlined } from '@ant-design/icons';
import { Alert, App, Button, Result, Skeleton, Tag, Typography } from 'antd';
import { useEffect } from 'react';
import { Link, useLocation, useParams } from 'react-router';
import { ApiError } from '../../../../api/client.ts';
import { tourProgramsApi } from '../../../../api/tourPrograms.ts';
import { useAuth } from '../../../../auth/useAuth.ts';
import PageHeader from '../../../../components/PageHeader.tsx';
import { ErrorCodes } from '../../../../constants/errorCodes.ts';
import { Permissions } from '../../../../constants/permissions.ts';
import { tourProgramStateInfo } from '../../../../constants/tourProgramStates.ts';
import PeriodTable from '../../../../features/tour-program/components/PeriodTable.tsx';
import TourProgramForm from '../../../../features/tour-program/components/TourProgramForm.tsx';
import { useTourProgram } from '../../../../features/tour-program/hooks/useTourProgram.ts';
import { errorMessage } from '../../../../lib/errors.ts';
import { formatDateTime } from '../../../../lib/format.ts';

const BackLink = () => (
  <Link to="/tour-programs" className="inline-flex w-fit items-center gap-1.5 text-ink-muted no-underline hover:text-ink">
    <ArrowLeftOutlined /> โปรแกรมทัวร์ทั้งหมด
  </Link>
);

export default function TourProgramEditPage() {
  const { id = '' } = useParams();
  const { hash } = useLocation();
  const { message, modal } = App.useApp();
  const { can } = useAuth();
  const { tour, error, reload, replace } = useTourProgram(id);

  // After create, jump to the periods table.
  const loaded = tour !== null;
  useEffect(() => {
    if (loaded && hash === '#periods') document.getElementById('periods')?.scrollIntoView({ behavior: 'smooth' });
  }, [loaded, hash]);

  if (error instanceof ApiError && error.code === ErrorCodes.NotFound) {
    return (
      <Result
        status="404"
        title="ไม่พบโปรแกรมทัวร์นี้"
        extra={
          <Link to="/tour-programs">
            <Button type="primary">กลับไปหน้าโปรแกรมทัวร์</Button>
          </Link>
        }
      />
    );
  }
  if (error !== null) {
    return (
      <>
        <BackLink />
        <Alert type="error" showIcon title={errorMessage(error)} />
      </>
    );
  }
  if (!tour) {
    return (
      <>
        <BackLink />
        <Skeleton active paragraph={{ rows: 8 }} />
      </>
    );
  }

  const canEdit = can(Permissions.TourProgramUpdate) && tour.status;
  const state = tourProgramStateInfo(tour.state);

  const confirmStatus = () => {
    const remove = tour.status;
    modal.confirm({
      title: remove ? `ลบ ${tour.name}?` : `กู้คืน ${tour.name}?`,
      content: remove ? 'ทัวร์นี้จะหายจากหน้าเว็บหลังอัปเดตครั้งถัดไป ข้อมูลทั้งหมดยังอยู่ครบ กู้คืนได้ภายหลัง' : 'ทัวร์นี้จะกลับมาแก้ไขและแสดงบนหน้าเว็บตามสถานะโปรแกรม',
      okText: remove ? 'ลบ' : 'กู้คืน',
      okButtonProps: { danger: remove },
      cancelText: 'ยกเลิก',
      onOk: async () => {
        try {
          await tourProgramsApi.updateStatus({ id: tour.id, status: !remove });
          message.success(remove ? 'ลบแล้ว' : 'กู้คืนแล้ว');
          reload();
        } catch (err) {
          message.error(errorMessage(err));
        }
      },
    });
  };

  return (
    <>
      <BackLink />
      <PageHeader
        title={tour.name}
        subtitle={
          <span className="flex flex-wrap items-center gap-2">
            <Typography.Text code copyable={{ text: tour.code, tooltips: ['คัดลอกรหัสทัวร์', 'คัดลอกแล้ว'] }}>
              {tour.code}
            </Typography.Text>
            <span className="font-mono text-[13px]">/tour/{tour.code}/</span>
            <Tag color={state.color} className="m-0" title={state.hint}>
              {state.label}
            </Tag>
            {!tour.status && <Tag className="m-0">ลบแล้ว</Tag>}
            <span className="text-[13px]">แก้ไขล่าสุด {formatDateTime(tour.updateDate)}</span>
          </span>
        }
        actions={
          can(Permissions.TourProgramStatus) && (
            <Button danger={tour.status} icon={tour.status ? <StopOutlined /> : <UndoOutlined />} onClick={confirmStatus}>
              {tour.status ? 'ลบโปรแกรม' : 'กู้คืนโปรแกรม'}
            </Button>
          )
        }
      />

      {!tour.status && <Alert type="warning" showIcon title="โปรแกรมนี้ถูกลบแล้ว กู้คืนก่อนจึงจะแก้ไขได้" />}

      <TourProgramForm
        tour={tour}
        readOnly={!canEdit}
        submitText="บันทึกโปรแกรม"
        onSubmit={async (body) => {
          const saved = await tourProgramsApi.update({ ...body, id: tour.id });
          replace(saved);
          message.success('บันทึกแล้ว — หน้าเว็บจะอัปเดตภายในไม่กี่นาที');
        }}
      />

      {can(Permissions.PeriodList) && (
        <div id="periods" className="scroll-mt-6">
          <PeriodTable tourProgramId={tour.id} days={tour.days} disabled={!tour.status} />
        </div>
      )}
    </>
  );
}
