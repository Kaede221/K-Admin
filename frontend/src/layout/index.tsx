import { Layout as AntLayout } from 'antd';
import { Outlet } from 'react-router-dom';
import { Suspense } from 'react';
import { useAppStore } from '@/store/appStore';
import Sidebar from './Sidebar';
import Header from './Header';
import Tabs from './Tabs';

const { Content } = AntLayout;

export function Layout() {
  const collapsed = useAppStore((state) => state.collapsed);

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
