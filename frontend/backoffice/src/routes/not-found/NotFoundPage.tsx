import { Button, Result } from 'antd';
import { Link } from 'react-router';

export default function NotFoundPage() {
  return (
    <Result
      status="404"
      title="404"
      subTitle="ไม่พบหน้าที่คุณต้องการ"
      extra={
        <Link to="/">
          <Button type="primary">กลับหน้าแรก</Button>
        </Link>
      }
    />
  );
}
