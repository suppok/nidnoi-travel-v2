import type { ThemeConfig } from 'antd';

// Colors follow the design example; keep in sync with the @theme tokens in styles/global.css.
export const theme: ThemeConfig = {
  token: {
    colorPrimary: '#F8559C',
    colorLink: '#F8559C',
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
    Layout: { siderBg: '#D9EEFB', triggerBg: '#C3E3F8', triggerColor: '#1D2B4F', headerBg: '#FFFFFF', bodyBg: '#F6FAFE', headerHeight: 56, headerPadding: '0 32px' },
    Menu: {
      itemBg: 'transparent',
      itemSelectedBg: '#F8559C',
      itemSelectedColor: '#FFFFFF',
      itemColor: '#3D4A66',
      itemHoverBg: 'rgba(29, 43, 79, 0.07)',
      itemHeight: 44,
      itemBorderRadius: 12,
      groupTitleColor: '#5A6785',
      // No color fade: the selected item should switch instantly on navigation.
      motionDurationSlow: '0s',
      motionDurationMid: '0s',
    },
    // Brand-pink highlight so the active option is obvious.
    Segmented: {
      trackBg: '#EAF0F7',
      itemColor: '#5A6785',
      itemHoverColor: '#1D2B4F',
      itemHoverBg: 'rgba(29, 43, 79, 0.06)',
      itemSelectedBg: '#FFE6EF',
      itemSelectedColor: '#B02A61',
    },
    Table: { headerBg: '#F6FAFE' },
  },
};
