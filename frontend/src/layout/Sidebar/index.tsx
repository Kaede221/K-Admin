import { Menu, Layout } from 'antd';
import { useNavigate, useLocation } from 'react-router-dom';
import { useUserStore } from '@/store/userStore';
import { useAppStore } from '@/store/appStore';
import type { MenuItem as AntMenuItem } from 'antd';
import type { MenuItem } from '@/types/menu';
import * as Icons from '@ant-design/icons';

const { Sider } = Layout;

/**
 * Convert menu items to Ant Design Menu items
 */
function convertToAntMenuItems(menus: MenuItem[]): AntMenuItem[] {
  return menus
    .filter((menu) => !menu.meta?.hidden)
    .map((menu) => {
      // Get icon component dynamically
      const IconComponent = menu.meta?.icon
        ? (Icons as any)[menu.meta.icon]
        : null;

      const item: AntMenuItem = {
        key: menu.path,
        label: menu.meta?.title || menu.name,
        icon: IconComponent ? <IconComponent /> : null,
      };

      // Handle children recursively
      if (menu.children && menu.children.length > 0) {
        item.children = convertToAntMenuItems(menu.children);
      }

      return item;
    });
}

export function Sidebar() {
  const navigate = useNavigate();
  const location = useLocation();
  const menuTree = useUserStore((state) => state.menuTree);
  const collapsed = useAppStore((state) => state.collapsed);
  const addTab = useAppStore((state) => state.addTab);

  // Convert menu tree to Ant Design menu items
  const menuItems = convertToAntMenuItems(menuTree);

  // Handle menu click
  const handleMenuClick = ({ key }: { key: string }) => {
    // Find menu item by path
    const findMenuItem = (menus: MenuItem[], path: string): MenuItem | null => {
      for (const menu of menus) {
        if (menu.path === path) {
          return menu;
        }
        if (menu.children) {
          const found = findMenuItem(menu.children, path);
          if (found) return found;
        }
      }
      return null;
    };

    const menuItem = findMenuItem(menuTree, key);
    if (menuItem) {
      // Add tab
      addTab({
        key: menuItem.path,
        label: menuItem.meta?.title || menuItem.name,
        path: menuItem.path,
        closable: true,
      });

      // Navigate to route
      navigate(key);
    }
  };

  return (
    <Sider
      collapsible
      collapsed={collapsed}
      onCollapse={(value) => useAppStore.getState().setSidebarCollapsed(value)}
      width={200}
      style={{
        overflow: 'auto',
        height: '100vh',
        position: 'fixed',
        left: 0,
        top: 0,
        bottom: 0,
      }}
    >
      <div
        style={{
          height: 64,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          color: '#fff',
          fontSize: 20,
          fontWeight: 'bold',
        }}
      >
        {collapsed ? 'K' : 'K-Admin'}
      </div>
      <Menu
        theme="dark"
        mode="inline"
        selectedKeys={[location.pathname]}
        items={menuItems}
        onClick={handleMenuClick}
      />
    </Sider>
  );
}

export default Sidebar;
