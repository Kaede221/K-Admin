import { RouterProvider } from 'react-router-dom';
import { ConfigProvider, App as AntApp } from 'antd';
import zhCN from 'antd/locale/zh_CN';
import { router } from './router';
import { useAppStore } from './store/appStore';
import ErrorBoundary from './components/ErrorBoundary';

const App = () => {
  const theme = useAppStore((state) => state.theme);

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
