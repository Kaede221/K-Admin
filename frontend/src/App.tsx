import { RouterProvider } from 'react-router-dom';
import { ConfigProvider, App as AntApp } from 'antd';
import zhCN from 'antd/locale/zh_CN';
import { createAppRouter } from './router';
import { useAppStore } from './store/appStore';
import { useUserStore } from './store/userStore';
import ErrorBoundary from './components/ErrorBoundary';
import { useMemo } from 'react';

const App = () => {
  const theme = useAppStore((state) => state.theme);
  const menuTree = useUserStore((state) => state.menuTree);
  
  // Recreate router when menuTree changes
  const router = useMemo(() => createAppRouter(), [menuTree]);

  return (
    <ErrorBoundary>
      <ConfigProvider
        locale={zhCN}
        theme={{
          token: {
            colorPrimary: '#1890ff',
          },
          algorithm: theme === 'dark' ? undefined : undefined,
        }}
      >
        <AntApp>
          <RouterProvider router={router} />
        </AntApp>
      </ConfigProvider>
    </ErrorBoundary>
  );
};

export default App;
