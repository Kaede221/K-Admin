import { useEffect } from 'react';
import { Modal, Form, Input, Select, InputNumber, message } from 'antd';
import { createRole, updateRole, type RoleInfo } from '@/api/role';

interface RoleModalProps {
  visible: boolean;
  role?: RoleInfo;
  onSuccess: () => void;
  onCancel: () => void;
}

export function RoleModal({ visible, role, onSuccess, onCancel }: RoleModalProps) {
  const [form] = Form.useForm();
  const isEdit = !!role;

  useEffect(() => {
    if (visible) {
      if (role) {
        // Edit mode: populate form with role data
        form.setFieldsValue({
          roleName: role.roleName,
          roleKey: role.roleKey,
          dataScope: role.dataScope,
          sort: role.sort,
          status: role.status,
          remark: role.remark,
        });
      } else {
        // Create mode: reset form
        form.resetFields();
      }
    }
  }, [visible, role, form]);

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      
      if (isEdit) {
        // Update existing role
        await updateRole({
          id: role.id,
          ...values,
        });
        message.success('角色更新成功');
      } else {
        // Create new role
        await createRole(values);
        message.success('角色创建成功');
      }
      
      onSuccess();
    } catch (error: any) {
      if (error.errorFields) {
        // Form validation error
        return;
      }
      message.error(isEdit ? '角色更新失败' : '角色创建失败');
    }
  };

  return (
    <Modal
      title={isEdit ? '编辑角色' : '新建角色'}
      open={visible}
      onOk={handleSubmit}
      onCancel={onCancel}
      width={600}
      destroyOnClose
    >
      <Form
        form={form}
        layout="vertical"
        initialValues={{
          dataScope: 'all',
          sort: 0,
          status: true,
        }}
      >
        <Form.Item
          label="角色名称"
          name="roleName"
          rules={[
            { required: true, message: '请输入角色名称' },
            { max: 50, message: '角色名称长度不能超过 50 个字符' },
          ]}
        >
          <Input placeholder="请输入角色名称" />
        </Form.Item>

        <Form.Item
          label="角色标识"
          name="roleKey"
          rules={[
            { required: true, message: '请输入角色标识' },
            { max: 50, message: '角色标识长度不能超过 50 个字符' },
            { pattern: /^[a-zA-Z0-9_]+$/, message: '角色标识只能包含字母、数字和下划线' },
          ]}
        >
          <Input placeholder="请输入角色标识" disabled={isEdit} />
        </Form.Item>

        <Form.Item
          label="数据权限"
          name="dataScope"
          rules={[{ required: true, message: '请选择数据权限' }]}
        >
          <Select placeholder="请选择数据权限">
            <Select.Option value="all">全部数据</Select.Option>
            <Select.Option value="dept">部门数据</Select.Option>
            <Select.Option value="self">个人数据</Select.Option>
          </Select>
        </Form.Item>

        <Form.Item
          label="排序"
          name="sort"
          rules={[{ required: true, message: '请输入排序' }]}
        >
          <InputNumber
            placeholder="请输入排序"
            min={0}
            style={{ width: '100%' }}
          />
        </Form.Item>

        <Form.Item
          label="状态"
          name="status"
          rules={[{ required: true, message: '请选择状态' }]}
        >
          <Select>
            <Select.Option value={true}>启用</Select.Option>
            <Select.Option value={false}>禁用</Select.Option>
          </Select>
        </Form.Item>

        <Form.Item
          label="备注"
          name="remark"
          rules={[{ max: 255, message: '备注长度不能超过 255 个字符' }]}
        >
          <Input.TextArea
            placeholder="请输入备注"
            rows={4}
            maxLength={255}
            showCount
          />
        </Form.Item>
      </Form>
    </Modal>
  );
}

export default RoleModal;
