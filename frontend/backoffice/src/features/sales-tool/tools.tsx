import { FileTextOutlined, IdcardOutlined, SafetyOutlined, SearchOutlined } from '@ant-design/icons';
import type { ReactNode } from 'react';
import { type Permission, Permissions } from '../../constants/permissions.ts';
import BookingSummaryTool from './booking-summary/BookingSummaryTool.tsx';

export type SalesTool = {
  key: string;
  /** Route of the tool page, `/sales-tools/<key>`. */
  path: string;
  label: string;
  description: string;
  icon: ReactNode;
  /** The tool page. Omit while the tool is not built yet — the route shows a "coming soon" page. */
  element?: ReactNode;
  /** Permission the tool requires. Omit when any signed-in staff may use it. */
  permission?: Permission;
};

// Single registry for the sales tools section: drives the sidebar group and the routes in App.tsx.
export const SALES_TOOLS: SalesTool[] = [
  {
    key: 'booking-summary',
    path: '/sales-tools/booking-summary',
    label: 'สรุปยอดโอน',
    description: 'สร้างข้อความสรุปการจองและยอดชำระ ส่งให้ลูกค้าโอนเงิน',
    icon: <FileTextOutlined />,
    element: <BookingSummaryTool />,
    // Also reads tours, periods and booking sources; sales-user:list marks it as a sales tool.
    permission: Permissions.SalesUserList,
  },
  {
    key: 'passport-ocr',
    path: '/sales-tools/passport-ocr',
    label: 'อ่านพาสปอร์ต',
    description: 'ถ่ายหรืออัปโหลดหน้าพาสปอร์ต แล้วดึงชื่อ เลขที่ และวันหมดอายุอัตโนมัติ',
    icon: <IdcardOutlined />,
  },
  {
    key: 'program-search',
    path: '/sales-tools/program-search',
    label: 'ค้นหาโปรแกรมด่วน',
    description: 'ค้นหาโปรแกรมทัวร์ ช่วงเดินทาง ราคา และที่นั่งว่างได้ในที่เดียว',
    icon: <SearchOutlined />,
  },
  {
    key: 'insurance-search',
    path: '/sales-tools/insurance-search',
    label: 'ค้นหาประกันเดินทาง',
    description: 'ค้นหาแผนประกันเดินทางและเบี้ยประกันตามประเทศและจำนวนวัน',
    icon: <SafetyOutlined />,
  },
];
