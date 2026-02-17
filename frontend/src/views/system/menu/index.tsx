import { useState, useEffect } from 'react';
import { Button, Space, Tag, Table, Card, message, Modal } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import { AuthButton } from '@/components/Auth';
import { getAllMenus, deleteMenu } from '@/api/menu';
import type { MenuItem } from '@/types/menu';
import MenuModal from './components/MenuModal';

export function MenuManagement() {
  const [loading, setLoading] = useState(false);
  const [dataSource, setDataSource] = useState<MenuItem[]>([]);
  const [modalVisible, setModalVisible] = useState(false);
  const [editingMenu, setEditingMenu] = useState<MenuItem | undefined>();

  useEffect(() => {
    loadMenus();
  }, []);

  const loadMenus = async () => {
    setLoading(true);
    try {
      const menus = await getAllMenus();
      setDataSource(menus);
    } catch (error) {
      message.error('加载菜单失败');
    } finally {
      setLoading(false);
    }
  };

  const handleCreate = () => {
    setEditingMenu(undefined);
    setModalVisible(true);
  };

  const handleEdit = (record: MenuItem) => {
    setEditingMenu(record);
    setModalVisible(true);
  };

  const handleDelete = (record: MenuItem) => {
    Modal.confirm({
      title: '确认删除',
      content: `确定要删除菜单 "${record.meta.title}" 吗？`,
      onOk: async () => {
        try {
          await deleteMenu(record.id);
          message.success('删除成功');
          loadMenus();
        } catch (error) {
          message.error('删除失败');
        }
      },
    });
  };

  const handleModalSuccess = () => {
    setModalVisible(false);
    loadMenus();
  };

  const columns: ColumnsType<MenuItem> = [
    {
      title: 'ID',
      dataIndex: 'id',
      width: 80,
    },
    {
      title: '菜单名称',
      dataIndex: ['meta', 'title'],
      width: 200,
    },
    {
      title: '图标',
      dataIndex: ['meta', 'icon'],
      width: 100,
      render: (icon: string) => icon || '-',
    },
    {
      title: '路由路径',
      dataIndex: 'path',
      width: 200,
    },
    {
      title: '组件路径',
      dataIndex: 'component',
      width: 250,
      ellipsis: true,
    },
    {
      title: '排序',
      dataIndex: 'sort',
      width: 80,
    },
    {
      title: '隐藏',
      dataIndex: ['meta', 'hidden'],
      width: 80,
      render: (hidden: boolean) => (
        <Tag color={hidden ? 'warning' : 'success'}>
          {hidden ? '是' : '否'}
        </Tag>
      ),
    },
    {
      title: '缓存',
      dataIndex: ['meta', 'keep_alive'],
      width: 80,
      render: (keepAlive: boolean) => (
        <Tag color={keepAlive ? 'success' : 'default'}>
          {keepAlive ? '是' : '否'}
        </Tag>
      ),
    },
    {
      title: '操作',
      key: 'action',
      width: 200,
      fixed: 'right',
      render: (_, record) => (
        <Space>
          <AuthButton
            perm="menu:edit"
            type="link"
            size="small"
            icon={<EditOutlined />}
            onClick={() => handleEdit(record)}
          >
            编辑
          </AuthButton>
          <AuthButton
            perm="menu:delete"
            type="link"
            size="small"
            danger
            icon={<DeleteOutlined />}
            onClick={() => handleDelete(record)}
          >
            删除
          </AuthButton>
        </Space>
      ),
    },
  ];

  return (
    <>
      <Card>
        <div style={{ marginBottom: 16 }}>
          <AuthButton
            perm="menu:create"
            type="primary"
            icon={<PlusOutlined />}
            onClick={handleCreate}
          >
            新建菜单
          </AuthButton>
        </div>

        <Table
          rowKey="id"
          columns={columns}
          dataSource={dataSource}
          loading={loading}
          pagination={false}
          expandable={{
            defaultExpandAllRows: true,
          }}
          scroll={{ x: 1400 }}
        />
      </Card>

      <MenuModal
        visible={modalVisible}
        menu={editingMenu}
        onSuccess={handleModalSuccess}
        onCancel={() => setModalVisible(false)}
      />
    </>
  );
}

export default MenuManagement;
