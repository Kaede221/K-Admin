import { Layout, Space, Avatar, Dropdown, Typography } from 'antd';
import { UserOutlined, LogoutOutlined, SettingOutlined } from '@ant-design/icons';
import type { MenuProps } from 'antd';
import { useUserStore } from '@/store/userStore';
import { ThemeSwitch } from '@/components/Theme/ThemeSwitch';

const { Header: AntHeader } = Layout;
const { Text } = Typography;

export function Header() {
  const userInfo = useUserStore((state) => state.userInfo);
  const logout = useUserStore((state) => state.logout);

  // User dropdown menu items
  const userMenuItems: MenuProps['items'] = [
    {
      key: 'profile',
      icon: <UserOutlined />,
      label: '个人中心',
    },
    {
      key: 'settings',
      icon: <SettingOutlined />,
      label: '设置',
    },
    {
      type: 'divider',
    },
    {
      key: 'logout',
      icon: <LogoutOutlined />,
      label: '退出登录',
      danger: true,
    },
  ];

  // Handle user menu click
  const handleMenuClick: MenuProps['onClick'] = ({ key }) => {
    if (key === 'logout') {
      logout();
    }
  };

  return (
    <AntHeader
      style={{
        padding: '0 24px',
        background: '#fff',
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center',
        boxShadow: '0 1px 4px rgba(0,21,41,.08)',
      }}
    >
      <div />
      
      <Space size="large">
        <ThemeSwitch />
        
        <Dropdown
          menu={{ items: userMenuItems, onClick: handleMenuClick }}
          placement="bottomRight"
        >
          <Space style={{ cursor: 'pointer' }}>
            <Avatar
              src={userInfo?.headerImg || undefined}
              icon={<UserOutlined />}
              size="small"
            />
            <Text>{userInfo?.nickname || userInfo?.username || '用户'}</Text>
          </Space>
        </Dropdown>
      </Space>
    </AntHeader>
  );
}

export default Header;
