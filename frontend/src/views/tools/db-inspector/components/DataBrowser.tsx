import { useState, useEffect } from 'react';
import { Table, Button, Space, message, Popconfirm, Modal, Form, Input } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined, ReloadOutlined } from '@ant-design/icons';
import { getTableData, deleteRecord, createRecord, updateRecord } from '@/api/dbInspector';
import type { ColumnsType } from 'antd/es/table';

interface DataBrowserProps {
  tableName: string;
}

export default function DataBrowser({ tableName }: DataBrowserProps) {
  const [data, setData] = useState<Record<string, any>[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(false);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [modalVisible, setModalVisible] = useState(false);
  const [editingRecord, setEditingRecord] = useState<Record<string, any> | null>(null);
  const [form] = Form.useForm();

  useEffect(() => {
    loadData();
  }, [tableName, page, pageSize]);

  const loadData = async () => {
    setLoading(true);
    try {
      const result = await getTableData({ tableName, page, pageSize });
      setData(result.data);
      setTotal(result.total);
    } catch (error) {
      message.error('Failed to load table data');
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async (record: Record<string, any>) => {
    try {
      await deleteRecord(tableName, record.id);
      message.success('Record deleted successfully');
      loadData();
    } catch (error) {
      message.error('Failed to delete record');
    }
  };

  const handleEdit = (record: Record<string, any>) => {
    setEditingRecord(record);
    form.setFieldsValue(record);
    setModalVisible(true);
  };

  const handleCreate = () => {
    setEditingRecord(null);
    form.resetFields();
    setModalVisible(true);
  };

  const handleModalOk = async () => {
    try {
      const values = await form.validateFields();
      if (editingRecord) {
        await updateRecord({ tableName, id: editingRecord.id, data: values });
        message.success('Record updated successfully');
      } else {
        await createRecord({ tableName, data: values });
        message.success('Record created successfully');
      }
      setModalVisible(false);
      loadData();
    } catch (error) {
      message.error('Operation failed');
    }
  };

  const columns: ColumnsType<Record<string, any>> = data.length > 0
    ? Object.keys(data[0]).map(key => ({
        title: key,
        dataIndex: key,
        key,
        ellipsis: true,
        width: 150,
      })).concat([
        {
          title: 'Actions',
          key: 'actions',
          fixed: 'right',
          width: 150,
          render: (_, record) => (
            <Space>
              <Button
                type="link"
                size="small"
                icon={<EditOutlined />}
                onClick={() => handleEdit(record)}
              >
                Edit
              </Button>
              <Popconfirm
                title="Are you sure to delete this record?"
                onConfirm={() => handleDelete(record)}
                okText="Yes"
                cancelText="No"
              >
                <Button type="link" size="small" danger icon={<DeleteOutlined />}>
                  Delete
                </Button>
              </Popconfirm>
            </Space>
          ),
        },
      ])
    : [];

  return (
    <div className="data-browser">
      <div style={{ marginBottom: 16 }}>
        <Space>
          <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>
            Create
          </Button>
          <Button icon={<ReloadOutlined />} onClick={loadData}>
            Refresh
          </Button>
        </Space>
      </div>
      <Table
        columns={columns}
        dataSource={data}
        rowKey="id"
        loading={loading}
        scroll={{ x: 'max-content' }}
        pagination={{
          current: page,
          pageSize,
          total,
          showSizeChanger: true,
          showTotal: total => `Total ${total} records`,
          onChange: (page, pageSize) => {
            setPage(page);
            setPageSize(pageSize);
          },
        }}
      />
      <Modal
        title={editingRecord ? 'Edit Record' : 'Create Record'}
        open={modalVisible}
        onOk={handleModalOk}
        onCancel={() => setModalVisible(false)}
        width={600}
      >
        <Form form={form} layout="vertical">
          {data.length > 0 &&
            Object.keys(data[0]).map(key => (
              <Form.Item key={key} label={key} name={key}>
                <Input />
              </Form.Item>
            ))}
        </Form>
      </Modal>
    </div>
  );
}
