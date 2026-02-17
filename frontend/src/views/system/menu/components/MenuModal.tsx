import { useEffect, useState } from 'react';
import { Modal, Form, Input, InputNumber, Select, Switch, TreeSelect, message } from 'antd';
import { createMenu, updateMenu, getAllMenus } from '@/api/menu';
import type { MenuItem } from '@/types/menu';
import type { DataNode } from 'antd/es/tree';

interface MenuModalProps {
  visible: boolean;
  menu?: MenuItem;
  onSuccess: () => void;
  onCancel: () => void;
}

export function MenuModal({ visible, menu, onSuccess, onCancel }: MenuModalProps) {
  const [form] = Form.useForm();
  const [menuTree, setMenuTree] = useState<DataNode[]>([]);
  const isEdit = !!menu;

  useEffect(() => {
    if (visible) {
      loadMenuTree();
      if (menu) {
        // Edit mode: populate form with menu data
        form.setFieldsValue({
          parentId: menu.parent_id || 0,
          path: menu.path,
          name: menu.name,
          component: menu.component,
          sort: menu.sort,
          icon: menu.meta.icon,
          title: menu.meta.title,
          hidden: menu.meta.hidden,
          keepAlive: menu.meta.keep_alive,
          btnPerms: menu.btn_perms?.join(',') || '',
        });
      } else {
        // Create mode: reset form
        form.resetFields();
      }
    }
  }, [visible, menu, form]);

  const loadMenuTree = async () => {
    try {
      const menus = await getAllMenus();
      const treeData = convertMenusToTreeData(menus);
      // Add root option
      setMenuTree([
        { value: 0, title: '根菜单', children: treeData },
      ]);
    } catch (error) {
      message.error('加载菜单树失败');
    }
  };

  const convertMenusToTreeData = (menus: MenuItem[]): DataNode[] => {
    return menus.map((menu) => ({
      value: menu.id,
      title: menu.meta.title,
      children: menu.children ? convertMenusToTreeData(menu.children) : undefined,
    }));
  };

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      
      const menuData = {
        parentId: values.parentId,
        path: values.path,
        name: values.name,
        component: values.component,
        sort: values.sort,
        meta: {
          icon: values.icon,
          title: values.title,
          hidden: values.hidden,
          keepAlive: values.keepAlive,
        },
        btnPerms: values.btnPerms ? values.btnPerms.split(',').map((p: string) => p.trim()).filter(Boolean) : [],
      };

      if (isEdit) {
        // Update existing menu
        await updateMenu({
          id: menu.id,
          ...menuData,
        });
        message.success('菜单更新成功');
      } else {
        // Create new menu
        await createMenu(menuData);
        message.success('菜单创建成功');
      }
      
      onSuccess();
    } catch (error: any) {
      if (error.errorFields) {
        // Form validation error
        return;
      }
      message.error(isEdit ? '菜单更新失败' : '菜单创建失败');
    }
  };

  return (
    <Modal
      title={isEdit ? '编辑菜单' : '新建菜单'}
      open={visible}
      onOk={handleSubmit}
      onCancel={onCancel}
      width={700}
      destroyOnClose
    >
      <Form
        form={form}
        layout="vertical"
        initialValues={{
          parentId: 0,
          sort: 0,
          hidden: false,
          keepAlive: true,
        }}
      >
        <Form.Item
          label="上级菜单"
          name="parentId"
          rules={[{ required: true, message: '请选择上级菜单' }]}
        >
          <TreeSelect
            treeData={menuTree}
            placeholder="请选择上级菜单"
            treeDefaultExpandAll
          />
        </Form.Item>

        <Form.Item
          label="菜单标题"
          name="title"
          rules={[
            { required: true, message: '请输入菜单标题' },
            { max: 50, message: '菜单标题长度不能超过 50 个字符' },
          ]}
        >
          <Input placeholder="请输入菜单标题" />
        </Form.Item>

        <Form.Item
          label="路由路径"
          name="path"
          rules={[
            { required: true, message: '请输入路由路径' },
            { max: 100, message: '路由路径长度不能超过 100 个字符' },
          ]}
        >
          <Input placeholder="请输入路由路径，如：/system/user" />
        </Form.Item>

        <Form.Item
          label="路由名称"
          name="name"
          rules={[
            { required: true, message: '请输入路由名称' },
            { max: 50, message: '路由名称长度不能超过 50 个字符' },
          ]}
        >
          <Input placeholder="请输入路由名称，如：SystemUser" />
        </Form.Item>

        <Form.Item
          label="组件路径"
          name="component"
          rules={[
            { required: true, message: '请输入组件路径' },
            { max: 100, message: '组件路径长度不能超过 100 个字符' },
          ]}
        >
          <Input placeholder="请输入组件路径，如：views/system/user/index" />
        </Form.Item>

        <Form.Item
          label="图标"
          name="icon"
        >
          <Input placeholder="请输入图标名称，如：UserOutlined" />
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
          label="隐藏菜单"
          name="hidden"
          valuePropName="checked"
        >
          <Switch checkedChildren="是" unCheckedChildren="否" />
        </Form.Item>

        <Form.Item
          label="页面缓存"
          name="keepAlive"
          valuePropName="checked"
        >
          <Switch checkedChildren="是" unCheckedChildren="否" />
        </Form.Item>

        <Form.Item
          label="按钮权限"
          name="btnPerms"
          tooltip="多个权限用逗号分隔，如：user:create,user:edit,user:delete"
        >
          <Input.TextArea
            placeholder="请输入按钮权限，多个权限用逗号分隔"
            rows={3}
          />
        </Form.Item>
      </Form>
    </Modal>
  );
}

export default MenuModal;
