import { App as AntApp, ConfigProvider } from 'antd';
import thTH from 'antd/locale/th_TH';
import { createBrowserRouter, Navigate, RouterProvider } from 'react-router';
import AuthProvider from './auth/AuthProvider.tsx';
import RequireAuth, { CHANGE_PASSWORD_PATH } from './auth/RequireAuth.tsx';
import RequirePermission from './auth/RequirePermission.tsx';
import { Permissions } from './constants/permissions.ts';
import AppLayout from './layouts/AppLayout.tsx';
import AuthLayout from './layouts/AuthLayout.tsx';
import ChangePasswordPage from './routes/change-password/ChangePasswordPage.tsx';
import DashboardPage from './routes/dashboard/DashboardPage.tsx';
import LoginPage from './routes/login/LoginPage.tsx';
import NotFoundPage from './routes/not-found/NotFoundPage.tsx';
import UserListPage from './routes/users/UserListPage.tsx';
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
            element: <RequirePermission code={Permissions.UserList} />,
            children: [{ path: '/users', element: <UserListPage /> }],
          },
        ],
      },
    ],
  },
  { path: '*', element: <NotFoundPage /> },
]);

export default function App() {
  return (
    <ConfigProvider locale={thTH} theme={theme}>
      <AntApp className="h-full">
        <AuthProvider>
          <RouterProvider router={router} />
        </AuthProvider>
      </AntApp>
    </ConfigProvider>
  );
}
