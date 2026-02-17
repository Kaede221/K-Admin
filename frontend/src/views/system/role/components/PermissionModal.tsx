import { useEffect, useState } from 'react';
import { Modal, Tabs, Tree, Table, Input, Button, Space, message } from 'antd';
import { PlusOutlined, DeleteOutlined } from '@ant-design/icons';
import type { DataNode } from 'antd/es/tree';
import { getAllMenus } from '@/api/menu';
import { getRoleMenus, assignMenus, getRoleAPIs, assignAPIs, type RoleInfo } from '@/api/role';
import type { MenuItem } from '@/types/menu';

interface PermissionModalProps {
  visible: boolean;
  role?: RoleInfo;
  onSuccess: () => void;
  onCancel: () => void;
}

export function PermissionModal({ visible, role, onSuccess, onCancel }: PermissionModalProps) {
  const [loading, setLoading] = useState(false);
  const [menuTree, setMenuTree] = useState<DataNode[]>([]);
  const [checkedMenuKeys, setCheckedMenuKeys] = useState<React.Key[]>([]);
  const [apiList, setApiList] = useState<{ path: string; method: string }[]>([]);
  const [newApiPath, setNewApiPath] = useState('');
  const [newApiMethod, setNewApiMethod] = useState('GET');

  useEffect(() => {
    if (visible && role) {
      loadMenuTree();
      loadRoleMenus();
      loadRoleAPIs();
    }
  }, [visible, role]);

  const loadMenuTree = async () => {
    try {
      const menus = await getAllMenus();
      const treeData = convertMenusToTreeData(menus);
      setMenuTree(treeData);
    } catch (error) {
      message.error('加载菜单树失败');
    }
  };

  const loadRoleMenus = async () => {
    if (!role) return;
    try {
      const menuIds = await getRoleMenus(role.id);
      setCheckedMenuKeys(menuIds);
    } catch (error) {
      message.error('加载角色菜单失败');
    }
  };

  const loadRoleAPIs = async () => {
    if (!role) return;
    try {
      const apis = await getRoleAPIs(role.id);
      const formattedApis = apis.map(([path, method]) => ({ path, method }));
      setApiList(formattedApis);
    } catch (error) {
      message.error('加载角色API权限失败');
    }
  };

  const convertMenusToTreeData = (menus: MenuItem[]): DataNode[] => {
    return menus.map((menu) => ({
      key: menu.id,
      title: menu.meta.title,
      children: menu.children ? convertMenusToTreeData(menu.children) : undefined,
    }));
  };

  const handleMenuCheck = (checkedKeys: React.Key[] | { checked: React.Key[]; halfChecked: React.Key[] }) => {
    const keys = Array.isArray(checkedKeys) ? checkedKeys : checkedKeys.checked;
    setCheckedMenuKeys(keys);
  };

  const handleAddAPI = () => {
    if (!newApiPath.trim()) {
      message.warning('请输入API路径');
      return;
    }
    
    const exists = apiList.some(
      (api) => api.path === newApiPath && api.method === newApiMethod
    );
    
    if (exists) {
      message.warning('该API权限已存在');
      return;
    }

    setApiList([...apiList, { path: newApiPath, method: newApiMethod }]);
    setNewApiPath('');
    setNewApiMethod('GET');
  };

  const handleDeleteAPI = (index: number) => {
    const newList = [...apiList];
    newList.splice(index, 1);
    setApiList(newList);
  };

  const handleSubmit = async () => {
    if (!role) return;

    setLoading(true);
    try {
      // Assign menus
      await assignMenus({
        roleId: role.id,
        menuIds: checkedMenuKeys.map((key) => Number(key)),
      });

      // Assign APIs
      const policies = apiList.map((api) => [api.path, api.method]);
      await assignAPIs({
        roleId: role.id,
        policies,
      });

      message.success('权限分配成功');
      onSuccess();
    } catch (error) {
      message.error('权限分配失败');
    } finally {
      setLoading(false);
    }
  };

  const apiColumns = [
    {
      title: 'API路径',
      dataIndex: 'path',
      key: 'path',
    },
    {
      title: '请求方法',
      dataIndex: 'method',
      key: 'method',
      width: 120,
    },
    {
      title: '操作',
      key: 'action',
      width: 100,
      render: (_: any, __: any, index: number) => (
        <Button
          type="link"
          danger
          size="small"
          icon={<DeleteOutlined />}
          onClick={() => handleDeleteAPI(index)}
        >
          删除
        </Button>
      ),
    },
  ];

  const tabItems = [
    {
      key: 'menu',
      label: '菜单权限',
      children: (
        <Tree
          checkable
          treeData={menuTree}
          checkedKeys={checkedMenuKeys}
          onCheck={handleMenuCheck}
          style={{ maxHeight: 400, overflow: 'auto' }}
        />
      ),
    },
    {
      key: 'api',
      label: 'API权限',
      children: (
        <div>
          <Space style={{ marginBottom: 16 }}>
            <Input
              placeholder="API路径，如：/api/v1/user"
              value={newApiPath}
              onChange={(e) => setNewApiPath(e.target.value)}
              style={{ width: 300 }}
            />
            <Input
              placeholder="方法"
              value={newApiMethod}
              onChange={(e) => setNewApiMethod(e.target.value.toUpperCase())}
              style={{ width: 100 }}
            />
            <Button type="primary" icon={<PlusOutlined />} onClick={handleAddAPI}>
              添加
            </Button>
          </Space>
          <Table
            columns={apiColumns}
            dataSource={apiList}
            rowKey={(record, index) => `${record.path}-${record.method}-${index}`}
            pagination={false}
            size="small"
            style={{ maxHeight: 350, overflow: 'auto' }}
          />
        </div>
      ),
    },
  ];

  return (
    <Modal
      title={`分配权限 - ${role?.roleName}`}
      open={visible}
      onOk={handleSubmit}
      onCancel={onCancel}
      width={700}
      confirmLoading={loading}
      destroyOnClose
    >
      <Tabs items={tabItems} />
    </Modal>
  );
}

export default PermissionModal;
