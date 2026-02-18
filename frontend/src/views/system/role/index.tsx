import { useRef, useState } from 'react';
import { Button, Space, Tag, message, Modal } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined, SafetyOutlined } from '@ant-design/icons';
import { ProTable, type ProTableActionRef, type ProTableColumn } from '@/components/ProTable';
import { AuthButton } from '@/components/Auth';
import { getRoleList, deleteRole, type RoleInfo } from '@/api/role';
import { formatDate } from '@/utils/format';
import RoleModal from './components/RoleModal';
import PermissionModal from './components/PermissionModal';

export function RoleManagement() {
  const actionRef = useRef<ProTableActionRef>(null);
  const [roleModalVisible, setRoleModalVisible] = useState(false);
  const [permissionModalVisible, setPermissionModalVisible] = useState(false);
  const [editingRole, setEditingRole] = useState<RoleInfo | undefined>();
  const [selectedRole, setSelectedRole] = useState<RoleInfo | undefined>();

  const handleCreate = () => {
    setEditingRole(undefined);
    setRoleModalVisible(true);
  };

  const handleEdit = (record: RoleInfo) => {
    setEditingRole(record);
    setRoleModalVisible(true);
  };

  const handleDelete = (record: RoleInfo) => {
    Modal.confirm({
      title: '确认删除',
      content: `确定要删除角色 "${record.roleName}" 吗？`,
      onOk: async () => {
        try {
          await deleteRole(record.id);
          message.success('删除成功');
          actionRef.current?.reload();
        } catch (error) {
          message.error('删除失败');
        }
      },
    });
  };

  const handleAssignPermissions = (record: RoleInfo) => {
    setSelectedRole(record);
    setPermissionModalVisible(true);
  };

  const handleRoleModalSuccess = () => {
    setRoleModalVisible(false);
    actionRef.current?.reload();
  };

  const handlePermissionModalSuccess = () => {
    setPermissionModalVisible(false);
    message.success('权限分配成功');
  };

  const columns: ProTableColumn<RoleInfo>[] = [
    {
      title: 'ID',
      dataIndex: 'id',
      width: 80,
      hideInSearch: true,
    },
    {
      title: '角色名称',
      dataIndex: 'roleName',
      hideInSearch: true,
    },
    {
      title: '角色标识',
      dataIndex: 'roleKey',
      hideInSearch: true,
    },
    {
      title: '数据权限',
      dataIndex: 'dataScope',
      width: 120,
      hideInSearch: true,
      render: (dataScope: string) => {
        const scopeMap: Record<string, { text: string; color: string }> = {
          all: { text: '全部数据', color: 'blue' },
          dept: { text: '部门数据', color: 'green' },
          self: { text: '个人数据', color: 'orange' },
        };
        const scope = scopeMap[dataScope] || { text: dataScope, color: 'default' };
        return <Tag color={scope.color}>{scope.text}</Tag>;
      },
    },
    {
      title: '排序',
      dataIndex: 'sort',
      width: 80,
      hideInSearch: true,
    },
    {
      title: '状态',
      dataIndex: 'status',
      width: 100,
      hideInSearch: true,
      render: (status: boolean) => (
        <Tag color={status ? 'success' : 'error'}>
          {status ? '启用' : '禁用'}
        </Tag>
      ),
    },
    {
      title: '备注',
      dataIndex: 'remark',
      hideInSearch: true,
      ellipsis: true,
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
      width: 260,
      hideInSearch: true,
      render: (_, record) => (
        <Space>
          <AuthButton
            perm="role:update"
            type="link"
            size="small"
            icon={<EditOutlined />}
            onClick={() => handleEdit(record)}
          >
            编辑
          </AuthButton>
          <AuthButton
            perm="role:delete"
            type="link"
            size="small"
            danger
            icon={<DeleteOutlined />}
            onClick={() => handleDelete(record)}
          >
            删除
          </AuthButton>
          <AuthButton
            perm="role:update"
            type="link"
            size="small"
            icon={<SafetyOutlined />}
            onClick={() => handleAssignPermissions(record)}
          >
            分配权限
          </AuthButton>
        </Space>
      ),
    },
  ];

  const request = async (params: any) => {
    const { page, page_size } = params;
    const result = await getRoleList({
      page,
      pageSize: page_size,
    });
    return {
      data: result.list,
      total: result.total,
    };
  };

  return (
    <>
      <ProTable<RoleInfo>
        ref={actionRef}
        columns={columns}
        request={request}
        search={false}
        toolbar={
          <AuthButton
            perm="role:create"
            type="primary"
            icon={<PlusOutlined />}
            onClick={handleCreate}
          >
            新建角色
          </AuthButton>
        }
      />

      <RoleModal
        visible={roleModalVisible}
        role={editingRole}
        onSuccess={handleRoleModalSuccess}
        onCancel={() => setRoleModalVisible(false)}
      />

      <PermissionModal
        visible={permissionModalVisible}
        role={selectedRole}
        onSuccess={handlePermissionModalSuccess}
        onCancel={() => setPermissionModalVisible(false)}
      />
    </>
  );
}

export default RoleManagement;
