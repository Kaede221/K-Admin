import { useState, useEffect } from 'react';
import { Table, Tag, message, Spin } from 'antd';
import { KeyOutlined, CheckCircleOutlined, CloseCircleOutlined } from '@ant-design/icons';
import { getTableSchema, ColumnInfo } from '@/api/dbInspector';
import type { ColumnsType } from 'antd/es/table';

interface SchemaViewerProps {
  tableName: string;
}

export default function SchemaViewer({ tableName }: SchemaViewerProps) {
  const [schema, setSchema] = useState<ColumnInfo[]>([]);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    loadSchema();
  }, [tableName]);

  const loadSchema = async () => {
    setLoading(true);
    try {
      const data = await getTableSchema(tableName);
      setSchema(data);
    } catch (error) {
      message.error('Failed to load table schema');
    } finally {
      setLoading(false);
    }
  };

  const columns: ColumnsType<ColumnInfo> = [
    {
      title: 'Column Name',
      dataIndex: 'name',
      key: 'name',
      width: 200,
      render: (text, record) => (
        <span>
          {record.key && <KeyOutlined style={{ marginRight: 8, color: '#faad14' }} />}
          {text}
        </span>
      ),
    },
    {
      title: 'Type',
      dataIndex: 'type',
      key: 'type',
      width: 150,
      render: text => <Tag color="blue">{text}</Tag>,
    },
    {
      title: 'Nullable',
      dataIndex: 'nullable',
      key: 'nullable',
      width: 100,
      align: 'center',
      render: nullable =>
        nullable ? (
          <CheckCircleOutlined style={{ color: '#52c41a' }} />
        ) : (
          <CloseCircleOutlined style={{ color: '#ff4d4f' }} />
        ),
    },
    {
      title: 'Key',
      dataIndex: 'key',
      key: 'key',
      width: 100,
      render: key => key && <Tag color="gold">{key}</Tag>,
    },
    {
      title: 'Default',
      dataIndex: 'default',
      key: 'default',
      width: 150,
      render: text => text || '-',
    },
    {
      title: 'Extra',
      dataIndex: 'extra',
      key: 'extra',
      width: 150,
      render: text => text || '-',
    },
    {
      title: 'Comment',
      dataIndex: 'comment',
      key: 'comment',
      ellipsis: true,
      render: text => text || '-',
    },
  ];

  return (
    <Spin spinning={loading}>
      <div className="schema-viewer">
        <h3 style={{ marginBottom: 16 }}>Table: {tableName}</h3>
        <Table
          columns={columns}
          dataSource={schema}
          rowKey="name"
          pagination={false}
          bordered
          size="small"
        />
      </div>
    </Spin>
  );
}
