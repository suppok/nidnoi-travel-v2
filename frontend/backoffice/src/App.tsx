import { App as AntApp, ConfigProvider } from 'antd';
import thTH from 'antd/locale/th_TH';
import { createBrowserRouter, Navigate, RouterProvider } from 'react-router';
import AuthLayout from './layouts/AuthLayout.tsx';
import LoginPage from './routes/login/LoginPage.tsx';
import NotFoundPage from './routes/not-found/NotFoundPage.tsx';

const router = createBrowserRouter([
  // No authenticated pages yet — send everyone to /login.
  { path: '/', element: <Navigate to="/login" replace /> },
  {
    element: <AuthLayout />,
    children: [{ path: '/login', element: <LoginPage /> }],
  },
  { path: '*', element: <NotFoundPage /> },
]);

export default function App() {
  return (
    <ConfigProvider locale={thTH} theme={{ token: { colorPrimary: '#0f766e', borderRadius: 8 } }}>
      <AntApp>
        <RouterProvider router={router} />
      </AntApp>
    </ConfigProvider>
  );
}
