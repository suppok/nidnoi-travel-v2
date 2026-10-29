import type { ThemeConfig } from 'antd';

// Colors follow the design example; keep in sync with the @theme tokens in styles/global.css.
export const theme: ThemeConfig = {
  token: {
    colorPrimary: '#C9356F',
    colorLink: '#C9356F',
    colorInfo: '#1C6496',
    colorText: '#1D2B4F',
    colorTextSecondary: '#5A6785',
    colorBorder: '#D6E2EE',
    colorBorderSecondary: '#E3EAF2',
    colorBgLayout: '#F6FAFE',
    borderRadius: 12,
    borderRadiusLG: 20,
    controlHeight: 40,
    fontFamily: "'Anuphan', 'Noto Sans Thai', system-ui, sans-serif",
    fontSize: 15,
  },
  components: {
    Layout: { siderBg: '#FFFFFF', headerBg: '#FFFFFF', bodyBg: '#F6FAFE' },
    Menu: {
      itemBg: 'transparent',
      itemSelectedBg: '#FFE6EF',
      itemSelectedColor: '#B02A61',
      itemColor: '#3D4A66',
      itemHeight: 44,
      itemBorderRadius: 12,
      groupTitleColor: '#5A6785',
    },
    Table: { headerBg: '#F6FAFE' },
  },
};
