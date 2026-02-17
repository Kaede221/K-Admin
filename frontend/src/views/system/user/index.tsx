import { useRef, useState } from 'react';
import { Button, Space, Tag, Input, Select, message, Modal } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined, KeyOutlined, PoweroffOutlined } from '@ant-design/icons';
import { ProTable, type ProTableActionRef, type ProTableColumn } from '@/components/ProTable';
import { AuthButton } from '@/components/Auth';
import { getUserList, deleteUser, toggleUserStatus, resetPassword } from '@/api/user';
import { formatDate } from '@/utils/format';
import type { UserInfo } from '@/types/user';
import UserModal from './components/UserModal';

export function UserManagement() {
  const actionRef = useRef<ProTableActionRef>(null);
  const [modalVisible, setModalVisible] = useState(false);
  const [editingUser, setEditingUser] = useState<UserInfo | undefined>();

  const handleCreate = () => {
    setEditingUser(undefined);
    setModalVisible(true);
  };

  const handleEdit = (record: UserInfo) => {
    setEditingUser(record);
    setModalVisible(true);
  };

  const handleDelete = (record: UserInfo) => {
    Modal.confirm({
      title: '确认删除',
      content: `确定要删除用户 "${record.username}" 吗？`,
      onOk: async () => {
        try {
          await deleteUser(record.id);
          message.success('删除成功');
          actionRef.current?.reload();
        } catch (error) {
          message.error('删除失败');
        }
      },
    });
  };

  const handleToggleStatus = async (record: UserInfo) => {
    try {
      await toggleUserStatus(record.id, !record.active);
      message.success(record.active ? '已禁用' : '已启用');
      actionRef.current?.reload();
    } catch (error) {
      message.error('操作失败');
    }
  };

  const handleResetPassword = (record: UserInfo) => {
    Modal.confirm({
      title: '重置密码',
      content: `确定要重置用户 "${record.username}" 的密码吗？新密码将设置为 "123456"`,
      onOk: async () => {
        try {
          await resetPassword({ userId: record.id, newPassword: '123456' });
          message.success('密码重置成功');
        } catch (error) {
          message.error('密码重置失败');
        }
      },
    });
  };

  const handleModalSuccess = () => {
    setModalVisible(false);
    actionRef.current?.reload();
  };

  const columns: ProTableColumn<UserInfo>[] = [
    {
      title: 'ID',
      dataIndex: 'id',
      width: 80,
      hideInSearch: true,
    },
    {
      title: '用户名',
      dataIndex: 'username',
      renderFormItem: () => <Input placeholder="请输入用户名" />,
    },
    {
      title: '昵称',
      dataIndex: 'nickname',
      hideInSearch: true,
    },
    {
      title: '手机号',
      dataIndex: 'phone',
      renderFormItem: () => <Input placeholder="请输入手机号" />,
    },
    {
      title: '邮箱',
      dataIndex: 'email',
      renderFormItem: () => <Input placeholder="请输入邮箱" />,
    },
    {
      title: '角色',
      dataIndex: ['role', 'role_name'],
      renderFormItem: () => (
        <Select placeholder="请选择角色" allowClear>
          {/* TODO: Load roles dynamically */}
          <Select.Option value={1}>管理员</Select.Option>
          <Select.Option value={2}>普通用户</Select.Option>
        </Select>
      ),
    },
    {
      title: '状态',
      dataIndex: 'active',
      width: 100,
      render: (active: boolean) => (
        <Tag color={active ? 'success' : 'error'}>
          {active ? '启用' : '禁用'}
        </Tag>
      ),
      renderFormItem: () => (
        <Select placeholder="请选择状态" allowClear>
          <Select.Option value={true}>启用</Select.Option>
          <Select.Option value={false}>禁用</Select.Option>
        </Select>
      ),
    },
    {
      title: '创建时间',
      dataIndex: 'createdAt',
      width: 180,
      hideInSearch: true,
      render: (createdAt: string) => formatDate(createdAt, 'YYYY-MM-DD'),
    },
    {
      title: '操作',
      key: 'action',
      width: 280,
      hideInSearch: true,
      render: (_, record) => (
        <Space>
          <AuthButton
            perm="user:edit"
            type="link"
            size="small"
            icon={<EditOutlined />}
            onClick={() => handleEdit(record)}
          >
            编辑
          </AuthButton>
          <AuthButton
            perm="user:delete"
            type="link"
            size="small"
            danger
            icon={<DeleteOutlined />}
            onClick={() => handleDelete(record)}
          >
            删除
          </AuthButton>
          <AuthButton
            perm="user:reset-password"
            type="link"
            size="small"
            icon={<KeyOutlined />}
            onClick={() => handleResetPassword(record)}
          >
            重置密码
          </AuthButton>
          <AuthButton
            perm="user:toggle-status"
            type="link"
            size="small"
            icon={<PoweroffOutlined />}
            onClick={() => handleToggleStatus(record)}
          >
            {record.active ? '禁用' : '启用'}
          </AuthButton>
        </Space>
      ),
    },
  ];

  const request = async (params: any) => {
    const { page, page_size, ...filters } = params;
    const result = await getUserList({
      page,
      pageSize: page_size,
      ...filters,
    });
    return {
      data: result.list,
      total: result.total,
    };
  };

  return (
    <>
      <ProTable<UserInfo>
        ref={actionRef}
        columns={columns}
        request={request}
        toolbar={
          <AuthButton
            perm="user:create"
            type="primary"
            icon={<PlusOutlined />}
            onClick={handleCreate}
          >
            新建用户
          </AuthButton>
        }
      />

      <UserModal
        visible={modalVisible}
        user={editingUser}
        onSuccess={handleModalSuccess}
        onCancel={() => setModalVisible(false)}
      />
    </>
  );
}

export default UserManagement;
