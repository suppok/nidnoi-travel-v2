import { StyleProvider } from '@ant-design/cssinjs';
import { App as AntApp, ConfigProvider } from 'antd';
import thTH from 'antd/locale/th_TH';
import { createBrowserRouter, Navigate, RouterProvider } from 'react-router';
import AuthProvider from './auth/AuthProvider.tsx';
import RequireAuth, { CHANGE_PASSWORD_PATH } from './auth/RequireAuth.tsx';
import RequirePermission from './auth/RequirePermission.tsx';
import { Permissions } from './constants/permissions.ts';
import AppLayout from './layouts/AppLayout.tsx';
import AuthLayout from './layouts/AuthLayout.tsx';
import AirlineListPage from './routes/airlines/AirlineListPage.tsx';
import BookingListPage from './routes/bookings/BookingListPage.tsx';
import ChangePasswordPage from './routes/change-password/ChangePasswordPage.tsx';
import DashboardPage from './routes/dashboard/DashboardPage.tsx';
import LoginPage from './routes/login/LoginPage.tsx';
import NotFoundPage from './routes/not-found/NotFoundPage.tsx';
import SiteBuildListPage from './routes/site-builds/SiteBuildListPage.tsx';
import CityListPage from './routes/taxonomy/cities/CityListPage.tsx';
import CountryDetailPage from './routes/taxonomy/countries/CountryDetailPage.tsx';
import CountryListPage from './routes/taxonomy/countries/CountryListPage.tsx';
import TourProgramEditPage from './routes/tour-programs/[id]/edit/TourProgramEditPage.tsx';
import TourProgramCreatePage from './routes/tour-programs/create/TourProgramCreatePage.tsx';
import TourProgramListPage from './routes/tour-programs/list/TourProgramListPage.tsx';
import UserListPage from './routes/users/UserListPage.tsx';
import WholesaleListPage from './routes/wholesales/WholesaleListPage.tsx';
import { theme } from './theme.ts';

const router = createBrowserRouter([
  { path: '/', element: <Navigate to="/dashboard" replace /> },
  {
    element: <AuthLayout />,
    children: [{ path: '/login', element: <LoginPage /> }],
  },
  {
    element: <RequireAuth />,
    children: [
      {
        element: <AuthLayout />,
        children: [{ path: CHANGE_PASSWORD_PATH, element: <ChangePasswordPage /> }],
      },
      {
        element: <AppLayout />,
        children: [
          { path: '/dashboard', element: <DashboardPage /> },
          {
            element: <RequirePermission code={Permissions.BookingList} />,
            children: [{ path: '/bookings', element: <BookingListPage /> }],
          },
          {
            element: <RequirePermission code={Permissions.TourProgramList} />,
            children: [{ path: '/tour-programs', element: <TourProgramListPage /> }],
          },
          {
            element: <RequirePermission code={Permissions.TourProgramCreate} />,
            children: [{ path: '/tour-programs/create', element: <TourProgramCreatePage /> }],
          },
          {
            element: <RequirePermission code={Permissions.TourProgramGet} />,
            children: [{ path: '/tour-programs/:id/edit', element: <TourProgramEditPage /> }],
          },
          {
            element: <RequirePermission code={Permissions.CountryList} />,
            children: [{ path: '/taxonomy/countries', element: <CountryListPage /> }],
          },
          {
            element: <RequirePermission code={Permissions.CountryGet} />,
            children: [{ path: '/taxonomy/countries/:id', element: <CountryDetailPage /> }],
          },
          {
            element: <RequirePermission code={Permissions.CityList} />,
            children: [{ path: '/taxonomy/cities', element: <CityListPage /> }],
          },
          {
            element: <RequirePermission code={Permissions.WholesaleList} />,
            children: [{ path: '/wholesales', element: <WholesaleListPage /> }],
          },
          {
            element: <RequirePermission code={Permissions.AirlineList} />,
            children: [{ path: '/airlines', element: <AirlineListPage /> }],
          },
          {
            element: <RequirePermission code={Permissions.UserList} />,
            children: [{ path: '/users', element: <UserListPage /> }],
          },
          {
            element: <RequirePermission code={Permissions.SiteBuildList} />,
            children: [{ path: '/site-builds', element: <SiteBuildListPage /> }],
          },
        ],
      },
    ],
  },
  { path: '*', element: <NotFoundPage /> },
]);

export default function App() {
  return (
    // `layer`: antd styles go in @layer antd, below Tailwind utilities (order in styles/global.css).
    <StyleProvider layer>
      <ConfigProvider locale={thTH} theme={theme}>
        <AntApp className="h-full">
          <AuthProvider>
            <RouterProvider router={router} />
          </AuthProvider>
        </AntApp>
      </ConfigProvider>
    </StyleProvider>
  );
}
