import { Tabs as AntTabs, Dropdown } from 'antd';
import { useNavigate, useLocation } from 'react-router-dom';
import { useAppStore, type TabItem } from '@/store/appStore';
import type { MenuProps } from 'antd';
import { ReloadOutlined, CloseOutlined } from '@ant-design/icons';

export function Tabs() {
  const navigate = useNavigate();
  const location = useLocation();
  const tabs = useAppStore((state) => state.tabs);
  const activeTab = useAppStore((state) => state.activeTab);
  const removeTab = useAppStore((state) => state.removeTab);
  const setActiveTab = useAppStore((state) => state.setActiveTab);
  const closeOtherTabs = useAppStore((state) => state.closeOtherTabs);
  const refreshTab = useAppStore((state) => state.refreshTab);

  // Handle tab change
  const handleTabChange = (key: string) => {
    setActiveTab(key);
    const tab = tabs.find((t) => t.key === key);
    if (tab) {
      navigate(tab.path);
    }
  };

  // Handle tab remove
  const handleTabRemove = (targetKey: string) => {
    removeTab(targetKey);
  };

  // Context menu items
  const getContextMenuItems = (tab: TabItem): MenuProps['items'] => [
    {
      key: 'refresh',
      icon: <ReloadOutlined />,
      label: '刷新',
      onClick: () => refreshTab(tab.key),
    },
    {
      key: 'close',
      icon: <CloseOutlined />,
      label: '关闭',
      disabled: !tab.closable,
      onClick: () => handleTabRemove(tab.key),
    },
    {
      key: 'closeOthers',
      label: '关闭其他',
      onClick: () => closeOtherTabs(tab.key),
    },
  ];

  // Convert tabs to Ant Design tab items
  const tabItems = tabs.map((tab) => ({
    key: tab.key,
    label: (
      <Dropdown
        menu={{ items: getContextMenuItems(tab) }}
        trigger={['contextMenu']}
      >
        <span>{tab.label}</span>
      </Dropdown>
    ),
    closable: tab.closable,
  }));

  return (
    <AntTabs
      type="editable-card"
      activeKey={activeTab}
      items={tabItems}
      onChange={handleTabChange}
      onEdit={(targetKey, action) => {
        if (action === 'remove') {
          handleTabRemove(targetKey as string);
        }
      }}
      hideAdd
      style={{ marginBottom: 0 }}
    />
  );
}

export default Tabs;
