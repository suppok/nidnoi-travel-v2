import { ArrowLeftOutlined } from '@ant-design/icons';
import { App, Button } from 'antd';
import { Link, useNavigate } from 'react-router';
import { tourProgramsApi } from '../../../api/tourPrograms.ts';
import PageHeader from '../../../components/PageHeader.tsx';
import TourProgramForm from '../../../features/tour-program/components/TourProgramForm.tsx';

export default function TourProgramCreatePage() {
  const { message } = App.useApp();
  const navigate = useNavigate();

  return (
    <>
      <Link to="/tour-programs" className="inline-flex w-fit items-center gap-1.5 text-ink-muted no-underline hover:text-ink">
        <ArrowLeftOutlined /> โปรแกรมทัวร์ทั้งหมด
      </Link>
      <PageHeader title="เพิ่มโปรแกรมทัวร์" subtitle="รหัสทัวร์จะถูกสร้างอัตโนมัติหลังบันทึก จากนั้นเพิ่มรอบเดินทางและราคาได้" />
      <TourProgramForm
        tour={null}
        submitText="สร้างโปรแกรม"
        extraActions={
          <Link to="/tour-programs">
            <Button>ยกเลิก</Button>
          </Link>
        }
        onSubmit={async (body) => {
          const created = await tourProgramsApi.create(body);
          message.success(`สร้างโปรแกรม ${created.code} แล้ว — เพิ่มรอบเดินทางได้ด้านล่าง`);
          navigate(`/tour-programs/${created.id}/edit#periods`, { replace: true });
        }}
      />
    </>
  );
}
