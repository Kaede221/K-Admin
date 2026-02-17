import { RouterProvider } from 'react-router-dom';
import { ConfigProvider, App as AntApp } from 'antd';
import zhCN from 'antd/locale/zh_CN';
import { createAppRouter } from './router';
import { useUserStore } from './store/userStore';
import ErrorBoundary from './components/ErrorBoundary';
import { useMemo, useEffect } from 'react';
import { setNavigate } from './utils/navigation';

const App = () => {
  const menuTree = useUserStore((state) => state.menuTree);
  
  // Recreate router when menuTree changes
  const router = useMemo(() => createAppRouter(), [menuTree]);

  // Set up navigation function for stores
  useEffect(() => {
    // Create a navigate function that works with the router
    const navigate = (path: string) => {
      router.navigate(path);
    };
    setNavigate(navigate);
  }, [router]);

  return (
    <ErrorBoundary>
      <ConfigProvider
        locale={zhCN}
        theme={{
          token: {
            colorPrimary: '#1890ff',
          },
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
