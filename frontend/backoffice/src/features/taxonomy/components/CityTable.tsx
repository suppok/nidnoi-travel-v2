import { EditOutlined, MoreOutlined, StopOutlined, UndoOutlined } from '@ant-design/icons';
import { App, Button, Dropdown, type MenuProps, Table, type TableProps, Tag, Typography } from 'antd';
import { citiesApi } from '../../../api/taxonomy.ts';
import { useAuth } from '../../../auth/useAuth.ts';
import { Permissions } from '../../../constants/permissions.ts';
import { errorMessage } from '../../../lib/errors.ts';
import type { City } from '../../../types/taxonomy.ts';

type Props = {
  cities: City[];
  loading: boolean;
  pagination: TableProps<City>['pagination'];
  showCountry?: boolean; // false inside a country's own page
  onEdit: (city: City) => void;
  onChanged: () => void; // after delete/restore
};

// City list with edit and delete/restore actions; used by the cities page and a country's detail page.
export default function CityTable({ cities, loading, pagination, showCountry = true, onEdit, onChanged }: Props) {
  const { message, modal } = App.useApp();
  const { can } = useAuth();

  const confirmStatus = (city: City) => {
    const remove = city.status;
    modal.confirm({
      title: remove ? `ลบ ${city.nameTh}?` : `กู้คืน ${city.nameTh}?`,
      content: remove ? 'เมืองนี้จะไม่แสดงบนหน้าเว็บหลังอัปเดตครั้งถัดไป และเลือกให้ทัวร์ใหม่ไม่ได้ กู้คืนได้ภายหลัง' : 'เมืองนี้จะกลับมาใช้งานได้อีกครั้ง',
      okText: remove ? 'ลบ' : 'กู้คืน',
      okButtonProps: { danger: remove },
      cancelText: 'ยกเลิก',
      onOk: async () => {
        try {
          await citiesApi.updateStatus({ id: city.id, status: !remove });
          message.success(remove ? 'ลบแล้ว' : 'กู้คืนแล้ว');
          onChanged();
        } catch (err) {
          message.error(errorMessage(err));
        }
      },
    });
  };

  const rowActions = (city: City): MenuProps['items'] =>
    [
      can(Permissions.CityUpdate) && { key: 'edit', icon: <EditOutlined />, label: 'แก้ไข', onClick: () => onEdit(city) },
      can(Permissions.CityStatus) &&
        (city.status
          ? { key: 'status', icon: <StopOutlined />, label: 'ลบ', danger: true, onClick: () => confirmStatus(city) }
          : {
              key: 'status',
              icon: <UndoOutlined />,
              // Backend: restoring needs an active country.
              label: city.country.status ? 'กู้คืน' : 'กู้คืนไม่ได้ (ประเทศถูกลบ)',
              disabled: !city.country.status,
              onClick: () => confirmStatus(city),
            }),
    ].filter((item) => item !== false);

  const columns: TableProps<City>['columns'] = [
    {
      title: 'เมือง',
      key: 'name',
      render: (_, city) => (
        <div className="flex flex-col">
          <span className="font-medium">{city.nameTh}</span>
          <span className="text-[13px] text-ink-muted">{city.nameEn}</span>
        </div>
      ),
    },
    ...(showCountry
      ? [
          {
            title: 'ประเทศ',
            key: 'country',
            render: (_: unknown, city: City) => (
              <span className="inline-flex flex-wrap items-center gap-1">
                {city.country.nameTh}
                {!city.country.status && <Tag className="m-0">ลบแล้ว</Tag>}
              </span>
            ),
          },
        ]
      : []),
    { title: 'Slug', dataIndex: 'slug', render: (slug: string) => <Typography.Text code>{slug}</Typography.Text> },
    { title: 'Ref ID', dataIndex: 'refId', width: 100, render: (v: number | null) => v ?? '—' },
    { title: 'ลำดับ', dataIndex: 'sortOrder', width: 80 },
    {
      title: 'สถานะ',
      dataIndex: 'status',
      width: 100,
      render: (status: boolean) => (status ? <Tag color="green" className="m-0">ใช้งาน</Tag> : <Tag className="m-0">ลบแล้ว</Tag>),
    },
    {
      key: 'actions',
      width: 56,
      align: 'right',
      render: (_, city) => {
        const items = rowActions(city);
        return items && items.length > 0 ? (
          <Dropdown menu={{ items }} trigger={['click']} placement="bottomRight">
            <Button type="text" icon={<MoreOutlined />} aria-label={`จัดการ ${city.nameTh}`} />
          </Dropdown>
        ) : null;
      },
    },
  ];

  return (
    <Table<City>
      rowKey="id"
      columns={columns}
      dataSource={cities}
      loading={loading}
      pagination={pagination}
      scroll={{ x: showCountry ? 820 : 680 }}
    />
  );
}
