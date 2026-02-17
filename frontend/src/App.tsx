import { RouterProvider } from 'react-router-dom';
import { ConfigProvider, App as AntApp, Spin } from 'antd';
import zhCN from 'antd/locale/zh_CN';
import { createAppRouter } from './router';
import { useUserStore } from './store/userStore';
import ErrorBoundary from './components/ErrorBoundary';
import { useMemo, useEffect, useState } from 'react';
import { setNavigate } from './utils/navigation';

const App = () => {
  const menuTree = useUserStore((state) => state.menuTree);
  const accessToken = useUserStore((state) => state.accessToken);
  const fetchUserMenu = useUserStore((state) => state.fetchUserMenu);
  const [isMenuLoading, setIsMenuLoading] = useState(false);
  
  // Fetch menu on mount if user is logged in but menu is empty
  useEffect(() => {
    if (accessToken && menuTree.length === 0) {
      setIsMenuLoading(true);
      fetchUserMenu()
        .catch((error) => {
          console.error('Failed to fetch user menu:', error);
        })
        .finally(() => {
          setIsMenuLoading(false);
        });
    }
  }, [accessToken, fetchUserMenu]);

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

  // Show loading spinner while fetching menu
  if (isMenuLoading) {
    return (
      <div style={{ 
        display: 'flex', 
        justifyContent: 'center', 
        alignItems: 'center', 
        height: '100vh' 
      }}>
        <Spin size="large" tip="加载菜单中..." />
      </div>
    );
  }

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
