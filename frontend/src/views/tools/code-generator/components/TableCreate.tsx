import { useState } from 'react';
import { Form, Input, Button, Space, Table, Select, Modal, message } from 'antd';
import { PlusOutlined, DeleteOutlined } from '@ant-design/icons';
import { createTable, FieldConfig } from '@/api/codeGenerator';
import type { ColumnsType } from 'antd/es/table';

interface TableCreateProps {
  visible: boolean;
  onClose: () => void;
  onSuccess: () => void;
}

interface FieldRow extends FieldConfig {
  key: string;
}

export default function TableCreate({ visible, onClose, onSuccess }: TableCreateProps) {
  const [form] = Form.useForm();
  const [fields, setFields] = useState<FieldRow[]>([]);
  const [loading, setLoading] = useState(false);

  const handleAddField = () => {
    const newField: FieldRow = {
      key: Date.now().toString(),
      columnName: '',
      fieldName: '',
      fieldType: 'string',
      jsonTag: '',
      gormTag: '',
      comment: '',
    };
    setFields([...fields, newField]);
  };

  const handleDeleteField = (key: string) => {
    setFields(fields.filter(f => f.key !== key));
  };

  const handleFieldChange = (key: string, field: keyof FieldConfig, value: string) => {
    setFields(
      fields.map(f => (f.key === key ? { ...f, [field]: value } : f))
    );
  };

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      if (fields.length === 0) {
        message.warning('Please add at least one field');
        return;
      }

      setLoading(true);
      await createTable({
        tableName: values.tableName,
        fields: fields.map(({ key, ...rest }) => rest),
      });
      message.success('Table created successfully');
      form.resetFields();
      setFields([]);
      onSuccess();
      onClose();
    } catch (error) {
      message.error('Failed to create table');
    } finally {
      setLoading(false);
    }
  };

  const columns: ColumnsType<FieldRow> = [
    {
      title: 'Column Name',
      dataIndex: 'columnName',
      key: 'columnName',
      width: 150,
      render: (text, record) => (
        <Input
          value={text}
          onChange={e => handleFieldChange(record.key, 'columnName', e.target.value)}
          placeholder="column_name"
          size="small"
        />
      ),
    },
    {
      title: 'Field Type',
      dataIndex: 'fieldType',
      key: 'fieldType',
      width: 120,
      render: (text, record) => (
        <Select
          value={text}
          onChange={value => handleFieldChange(record.key, 'fieldType', value)}
          size="small"
          style={{ width: '100%' }}
          options={[
            { label: 'string', value: 'string' },
            { label: 'int', value: 'int' },
            { label: 'uint', value: 'uint' },
            { label: 'bool', value: 'bool' },
            { label: 'float64', value: 'float64' },
            { label: 'time.Time', value: 'time.Time' },
          ]}
        />
      ),
    },
    {
      title: 'Nullable',
      dataIndex: 'gormTag',
      key: 'nullable',
      width: 100,
      render: (text, record) => (
        <Select
          value={text.includes('not null') ? 'no' : 'yes'}
          onChange={value => {
            const tag = value === 'no' ? 'not null' : '';
            handleFieldChange(record.key, 'gormTag', tag);
          }}
          size="small"
          style={{ width: '100%' }}
          options={[
            { label: 'Yes', value: 'yes' },
            { label: 'No', value: 'no' },
          ]}
        />
      ),
    },
    {
      title: 'Default',
      dataIndex: 'gormTag',
      key: 'default',
      width: 120,
      render: (text, record) => (
        <Input
          value={text.match(/default:([^;]+)/)?.[1] || ''}
          onChange={e => {
            const defaultValue = e.target.value;
            const tag = defaultValue ? `default:${defaultValue}` : '';
            handleFieldChange(record.key, 'gormTag', tag);
          }}
          placeholder="default value"
          size="small"
        />
      ),
    },
    {
      title: 'Comment',
      dataIndex: 'comment',
      key: 'comment',
      render: (text, record) => (
        <Input
          value={text}
          onChange={e => handleFieldChange(record.key, 'comment', e.target.value)}
          placeholder="field comment"
          size="small"
        />
      ),
    },
    {
      title: 'Action',
      key: 'action',
      width: 80,
      render: (_, record) => (
        <Button
          type="link"
          danger
          size="small"
          icon={<DeleteOutlined />}
          onClick={() => handleDeleteField(record.key)}
        />
      ),
    },
  ];

  return (
    <Modal
      title="Create Table"
      open={visible}
      onCancel={onClose}
      onOk={handleSubmit}
      width={900}
      confirmLoading={loading}
    >
      <Form form={form} layout="vertical">
        <Form.Item
          label="Table Name"
          name="tableName"
          rules={[{ required: true, message: 'Please enter table name' }]}
        >
          <Input placeholder="e.g., products" />
        </Form.Item>
      </Form>

      <div style={{ marginBottom: 16 }}>
        <Button type="dashed" onClick={handleAddField} icon={<PlusOutlined />} block>
          Add Field
        </Button>
      </div>

      <Table
        columns={columns}
        dataSource={fields}
        pagination={false}
        scroll={{ x: 'max-content' }}
        size="small"
      />
    </Modal>
  );
}
