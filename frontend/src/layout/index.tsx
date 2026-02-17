import { Layout as AntLayout } from 'antd';
import { Outlet } from 'react-router-dom';
import { Suspense, useEffect } from 'react';
import { useAppStore } from '@/store/appStore';
import { useUserStore } from '@/store/userStore';
import Sidebar from './Sidebar';
import Header from './Header';
import Tabs from './Tabs';

const { Content } = AntLayout;

export function Layout() {
  const collapsed = useAppStore((state) => state.collapsed);
  const fetchUserMenu = useUserStore((state) => state.fetchUserMenu);

  // 每次进入主布局时获取菜单（包括刷新页面）
  useEffect(() => {
    fetchUserMenu().catch((error) => {
      console.error('Failed to fetch user menu:', error);
    });
  }, []);

  return (
    <AntLayout style={{ minHeight: '100vh' }}>
      <Sidebar />
      
      <AntLayout
        style={{
          marginLeft: collapsed ? 80 : 200,
          transition: 'margin-left 0.2s',
        }}
      >
        <Header />
        
        <Tabs />
        
        <Content
          style={{
            margin: '16px',
            padding: '24px',
            background: '#fff',
            minHeight: 280,
          }}
        >
          <Suspense fallback={<div>Loading...</div>}>
            <Outlet />
          </Suspense>
        </Content>
      </AntLayout>
    </AntLayout>
  );
}

export default Layout;
