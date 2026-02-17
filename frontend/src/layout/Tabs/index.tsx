import { Tag, Dropdown, Space } from 'antd';
import { useNavigate } from 'react-router-dom';
import { useAppStore, type TabItem } from '@/store/appStore';
import type { MenuProps } from 'antd';
import { ReloadOutlined, CloseOutlined } from '@ant-design/icons';

export function Tabs() {
  const navigate = useNavigate();
  const tabs = useAppStore((state) => state.tabs);
  const activeTab = useAppStore((state) => state.activeTab);
  const removeTab = useAppStore((state) => state.removeTab);
  const setActiveTab = useAppStore((state) => state.setActiveTab);
  const closeOtherTabs = useAppStore((state) => state.closeOtherTabs);
  const refreshTab = useAppStore((state) => state.refreshTab);

  // Handle tab click
  const handleTabClick = (tab: TabItem) => {
    setActiveTab(tab.key);
    navigate(tab.path);
  };

  // Handle tab close
  const handleTabClose = (e: React.MouseEvent, key: string) => {
    e.stopPropagation();
    removeTab(key);
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
      onClick: () => removeTab(tab.key),
    },
    {
      key: 'closeOthers',
      label: '关闭其他',
      onClick: () => closeOtherTabs(tab.key),
    },
  ];

  return (
    <div
      style={{
        padding: '12px 16px',
        background: '#f5f5f5',
        borderBottom: '1px solid #e8e8e8',
      }}
    >
      <Space size={8}>
        {tabs.map((tab) => (
          <Dropdown
            key={tab.key}
            menu={{ items: getContextMenuItems(tab) }}
            trigger={['contextMenu']}
          >
            <Tag
              color={activeTab === tab.key ? 'blue' : 'default'}
              closable={tab.closable}
              onClose={(e) => handleTabClose(e, tab.key)}
              onClick={() => handleTabClick(tab)}
              style={{
                cursor: 'pointer',
                userSelect: 'none',
                margin: 0,
                padding: '4px 12px',
                fontSize: '13px',
                borderRadius: '16px',
              }}
            >
              {tab.label}
            </Tag>
          </Dropdown>
        ))}
      </Space>
    </div>
  );
}

export default Tabs;
