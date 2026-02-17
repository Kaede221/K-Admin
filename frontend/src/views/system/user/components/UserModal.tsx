import { useEffect } from 'react';
import { Modal, Form, Input, Select, Upload, message } from 'antd';
import { PlusOutlined } from '@ant-design/icons';
import type { UploadFile } from 'antd';
import { createUser, updateUser } from '@/api/user';
import type { UserInfo } from '@/types/user';

interface UserModalProps {
  visible: boolean;
  user?: UserInfo;
  onSuccess: () => void;
  onCancel: () => void;
}

export function UserModal({ visible, user, onSuccess, onCancel }: UserModalProps) {
  const [form] = Form.useForm();
  const isEdit = !!user;

  useEffect(() => {
    if (visible) {
      if (user) {
        // Edit mode: populate form with user data
        form.setFieldsValue({
          username: user.username,
          nickname: user.nickname,
          phone: user.phone,
          email: user.email,
          roleId: user.role_id,
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
      const values = await form.validateFields();
      
      if (isEdit) {
        // Update existing user
        await updateUser({
          id: user.id,
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
    }
  };

  return (
    <Modal
      title={isEdit ? '编辑用户' : '新建用户'}
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

        {!isEdit && (
          <Form.Item
            label="密码"
            name="password"
            rules={[
              { required: true, message: '请输入密码' },
              { min: 6, message: '密码长度至少为 6 个字符' },
            ]}
          >
            <Input.Password placeholder="请输入密码" />
          </Form.Item>
        )}

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
          <Select placeholder="请选择角色">
            {/* TODO: Load roles dynamically from API */}
            <Select.Option value={1}>管理员</Select.Option>
            <Select.Option value={2}>普通用户</Select.Option>
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
