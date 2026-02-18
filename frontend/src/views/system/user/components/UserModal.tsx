import { useEffect, useState } from 'react';
import { Modal, Form, Input, Select, Upload, message } from 'antd';
import { PlusOutlined } from '@ant-design/icons';
import type { UploadFile } from 'antd';
import { createUser, updateUser } from '@/api/user';
import { getRoleList, type RoleInfo } from '@/api/role';
import type { UserInfo } from '@/types/user';

interface UserModalProps {
  visible: boolean;
  user?: UserInfo;
  onSuccess: () => void;
  onCancel: () => void;
}

export function UserModal({ visible, user, onSuccess, onCancel }: UserModalProps) {
  const [form] = Form.useForm();
  const [roleList, setRoleList] = useState<RoleInfo[]>([]);
  const [loading, setLoading] = useState(false);
  const isEdit = !!user;

  // Load role list
  useEffect(() => {
    const loadRoles = async () => {
      try {
        const result = await getRoleList({ page: 1, pageSize: 100 });
        setRoleList(result.list);
      } catch (error) {
        console.error('Failed to load roles:', error);
      }
    };
    loadRoles();
  }, []);

  useEffect(() => {
    if (visible) {
      if (user) {
        // Edit mode: populate form with user data
        form.setFieldsValue({
          username: user.username,
          nickname: user.nickname,
          phone: user.phone,
          email: user.email,
          roleId: user.roleId,
          active: user.active,
        });
      } else {
        // Create mode: reset form
        form.resetFields();
      }
    }
  }, [visible, user, form]);

  const handleSubmit = async () => {
    try {
      setLoading(true);
      const values = await form.validateFields();
      
      // Remove password field if it's empty in edit mode
      if (isEdit && !values.password) {
        delete values.password;
      }
      
      if (isEdit) {
        // Update existing user
        await updateUser({
          id: user.id,
          username: user.username, // Username cannot be changed
          ...values,
        });
        message.success('用户更新成功');
      } else {
        // Create new user
        await createUser(values);
        message.success('用户创建成功');
      }
      
      onSuccess();
    } catch (error: any) {
      if (error.errorFields) {
        // Form validation error
        return;
      }
      message.error(isEdit ? '用户更新失败' : '用户创建失败');
    } finally {
      setLoading(false);
    }
  };

  return (
    <Modal
      title={isEdit ? '编辑用户' : '新建用户'}
      open={visible}
      onOk={handleSubmit}
      onCancel={onCancel}
      confirmLoading={loading}
      width={600}
      destroyOnClose
    >
      <Form
        form={form}
        layout="vertical"
        initialValues={{
          active: true,
        }}
      >
        <Form.Item
          label="用户名"
          name="username"
          rules={[
            { required: true, message: '请输入用户名' },
            { min: 3, max: 50, message: '用户名长度为 3-50 个字符' },
            { pattern: /^[a-zA-Z0-9_]+$/, message: '用户名只能包含字母、数字和下划线' },
          ]}
        >
          <Input placeholder="请输入用户名" disabled={isEdit} />
        </Form.Item>

        <Form.Item
          label={isEdit ? '密码（留空则不修改）' : '密码'}
          name="password"
          rules={[
            { required: !isEdit, message: '请输入密码' },
            { min: 6, message: '密码长度至少为 6 个字符' },
          ]}
        >
          <Input.Password placeholder={isEdit ? '留空则不修改密码' : '请输入密码'} />
        </Form.Item>

        <Form.Item
          label="昵称"
          name="nickname"
          rules={[{ max: 50, message: '昵称长度不能超过 50 个字符' }]}
        >
          <Input placeholder="请输入昵称" />
        </Form.Item>

        <Form.Item
          label="手机号"
          name="phone"
          rules={[
            { pattern: /^1[3-9]\d{9}$/, message: '请输入有效的手机号' },
          ]}
        >
          <Input placeholder="请输入手机号" />
        </Form.Item>

        <Form.Item
          label="邮箱"
          name="email"
          rules={[
            { type: 'email', message: '请输入有效的邮箱地址' },
            { max: 100, message: '邮箱长度不能超过 100 个字符' },
          ]}
        >
          <Input placeholder="请输入邮箱" />
        </Form.Item>

        <Form.Item
          label="角色"
          name="roleId"
          rules={[{ required: true, message: '请选择角色' }]}
        >
          <Select placeholder="请选择角色" loading={roleList.length === 0}>
            {roleList.map((role) => (
              <Select.Option key={role.id} value={role.id}>
                {role.roleName}
              </Select.Option>
            ))}
          </Select>
        </Form.Item>

        <Form.Item
          label="状态"
          name="active"
          rules={[{ required: true, message: '请选择状态' }]}
        >
          <Select>
            <Select.Option value={true}>启用</Select.Option>
            <Select.Option value={false}>禁用</Select.Option>
          </Select>
        </Form.Item>

        <Form.Item
          label="头像"
          name="headerImg"
          valuePropName="fileList"
          getValueFromEvent={(e) => {
            if (Array.isArray(e)) {
              return e;
            }
            return e?.fileList;
          }}
        >
          <Upload
            listType="picture-card"
            maxCount={1}
            beforeUpload={() => false}
          >
            <div>
              <PlusOutlined />
              <div style={{ marginTop: 8 }}>上传头像</div>
            </div>
          </Upload>
        </Form.Item>
      </Form>
    </Modal>
  );
}

export default UserModal;
