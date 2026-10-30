import { ArrowLeftOutlined, EditOutlined, PlusOutlined, StopOutlined, UndoOutlined } from '@ant-design/icons';
import { Alert, Button, Card, Descriptions, Result, Skeleton, Tag, Typography } from 'antd';
import { useState } from 'react';
import { Link, useParams } from 'react-router';
import { ApiError } from '../../../api/client.ts';
import { citiesApi } from '../../../api/taxonomy.ts';
import { useAuth } from '../../../auth/useAuth.ts';
import PageHeader from '../../../components/PageHeader.tsx';
import { ErrorCodes } from '../../../constants/errorCodes.ts';
import { Permissions } from '../../../constants/permissions.ts';
import CityTable from '../../../features/taxonomy/components/CityTable.tsx';
import TaxonomyFormModal from '../../../features/taxonomy/components/TaxonomyFormModal.tsx';
import { useConfirmCountryStatus } from '../../../features/taxonomy/hooks/useConfirmCountryStatus.ts';
import { useCountry } from '../../../features/taxonomy/hooks/useCountry.ts';
import { useCountryOptions } from '../../../features/taxonomy/hooks/useCountryOptions.ts';
import { useListQuery } from '../../../hooks/useListQuery.ts';
import { errorMessage } from '../../../lib/errors.ts';
import { formatDateTime } from '../../../lib/format.ts';
import type { City } from '../../../types/taxonomy.ts';

const BackLink = () => (
  <Link to="/taxonomy/countries" className="inline-flex w-fit items-center gap-1.5 text-ink-muted no-underline hover:text-ink">
    <ArrowLeftOutlined /> ประเทศทั้งหมด
  </Link>
);

export default function CountryDetailPage() {
  const { id = '' } = useParams();
  const { can } = useAuth();
  const { country, error, reload } = useCountry(id);
  const canListCities = can(Permissions.CityList);
  // Every city of the country (active and deleted); the largest has ~80.
  const cities = useListQuery(canListCities ? citiesApi.list : skipCities, { countryId: id, page: 1, pageSize: 500 });
  const { countries } = useCountryOptions();
  const confirmStatus = useConfirmCountryStatus(reload);

  const [editingCountry, setEditingCountry] = useState(false);
  const [editingCity, setEditingCity] = useState<City | null>(null);
  const [cityFormOpen, setCityFormOpen] = useState(false);

  const openCityForm = (city: City | null) => {
    setEditingCity(city);
    setCityFormOpen(true);
  };

  if (error instanceof ApiError && error.code === ErrorCodes.NotFound) {
    return (
      <Result
        status="404"
        title="ไม่พบประเทศนี้"
        extra={
          <Link to="/taxonomy/countries">
            <Button type="primary">กลับไปหน้าประเทศ</Button>
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
  if (!country) {
    return (
      <>
        <BackLink />
        <Card>
          <Skeleton active />
        </Card>
      </>
    );
  }

  const cityCount = cities.data?.total ?? 0;

  return (
    <>
      <BackLink />
      <PageHeader
        title={
          <span className="inline-flex flex-wrap items-center gap-3">
            {country.nameTh}
            {!country.status && <Tag className="m-0 text-sm">ลบแล้ว</Tag>}
          </span>
        }
        subtitle={country.nameEn}
        actions={
          <>
            {can(Permissions.CountryUpdate) && (
              <Button icon={<EditOutlined />} onClick={() => setEditingCountry(true)}>
                แก้ไข
              </Button>
            )}
            {can(Permissions.CountryStatus) &&
              (country.status ? (
                <Button danger icon={<StopOutlined />} onClick={() => confirmStatus(country)}>
                  ลบ
                </Button>
              ) : (
                <Button icon={<UndoOutlined />} onClick={() => confirmStatus(country)}>
                  กู้คืน
                </Button>
              ))}
          </>
        }
      />

      <Card title="ข้อมูลประเทศ">
        <Descriptions column={{ xs: 1, md: 2 }} size="small">
          <Descriptions.Item label="Slug">
            <Typography.Text code>{country.slug}</Typography.Text>
          </Descriptions.Item>
          <Descriptions.Item label="URL หน้าเว็บ">
            <Typography.Text code>/tours/{country.slug}/</Typography.Text>
          </Descriptions.Item>
          <Descriptions.Item label="Ref ID">{country.refId ?? '—'}</Descriptions.Item>
          <Descriptions.Item label="ลำดับการแสดง">{country.sortOrder}</Descriptions.Item>
          <Descriptions.Item label="สถานะ">
            {country.status ? <Tag color="green" className="m-0">ใช้งาน</Tag> : <Tag className="m-0">ลบแล้ว</Tag>}
          </Descriptions.Item>
          <Descriptions.Item label="แก้ไขล่าสุด">{formatDateTime(country.updateDate)}</Descriptions.Item>
          <Descriptions.Item label="คำอธิบาย" span={2}>
            {country.description ? <span className="whitespace-pre-line">{country.description}</span> : <span className="text-ink-muted">—</span>}
          </Descriptions.Item>
        </Descriptions>
      </Card>

      {canListCities && (
        <Card
          title={`เมืองในประเทศนี้ (${cityCount})`}
          extra={
            can(Permissions.CityCreate) &&
            country.status && (
              <Button type="primary" icon={<PlusOutlined />} onClick={() => openCityForm(null)}>
                เพิ่มเมือง
              </Button>
            )
          }
        >
          {cities.error !== null && <Alert type="error" showIcon title={errorMessage(cities.error)} className="mb-4" />}
          <CityTable
            cities={cities.data?.items ?? []}
            loading={cities.loading}
            pagination={false}
            showCountry={false}
            onEdit={openCityForm}
            onChanged={cities.reload}
          />
        </Card>
      )}

      <TaxonomyFormModal
        kind="country"
        open={editingCountry}
        item={country}
        onClose={() => setEditingCountry(false)}
        onSaved={() => {
          setEditingCountry(false);
          reload();
        }}
      />
      <TaxonomyFormModal
        kind="city"
        open={cityFormOpen}
        item={editingCity}
        countries={countries}
        defaultCountryId={country.id}
        onClose={() => setCityFormOpen(false)}
        onSaved={() => {
          setCityFormOpen(false);
          cities.reload(); // a city moved to another country drops out of this list
        }}
      />
    </>
  );
}

// Users without city:list never call the endpoint.
async function skipCities() {
  return { items: [] as City[], total: 0, page: 1, pageSize: 0 };
}
