import { App } from 'antd';
import { countriesApi } from '../../../api/taxonomy.ts';
import { errorMessage } from '../../../lib/errors.ts';
import type { Country } from '../../../types/taxonomy.ts';

// Returns a function that confirms and then soft-deletes or restores a country.
export function useConfirmCountryStatus(onDone: () => void) {
  const { message, modal } = App.useApp();

  return (country: Country) => {
    const remove = country.status;
    modal.confirm({
      title: remove ? `ลบ ${country.nameTh}?` : `กู้คืน ${country.nameTh}?`,
      content: remove
        ? 'ประเทศนี้จะไม่แสดงบนหน้าเว็บหลังอัปเดตครั้งถัดไป และเลือกให้ทัวร์หรือเมืองใหม่ไม่ได้ เมืองในประเทศนี้ยังอยู่ครบ กู้คืนได้ภายหลัง'
        : 'ประเทศนี้จะกลับมาใช้งานได้อีกครั้ง',
      okText: remove ? 'ลบ' : 'กู้คืน',
      okButtonProps: { danger: remove },
      cancelText: 'ยกเลิก',
      onOk: async () => {
        try {
          await countriesApi.updateStatus({ id: country.id, status: !remove });
          message.success(remove ? 'ลบแล้ว' : 'กู้คืนแล้ว');
          onDone();
        } catch (err) {
          message.error(errorMessage(err));
        }
      },
    });
  };
}
