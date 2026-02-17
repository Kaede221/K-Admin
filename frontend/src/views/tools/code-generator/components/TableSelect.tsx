import { useState, useEffect } from 'react';
import { Radio, Button, Space, Table, message, Spin, Card } from 'antd';
import { getTables } from '@/api/dbInspector';
import { getTableMetadata, FieldConfig } from '@/api/codeGenerator';
import type { ColumnsType } from 'antd/es/table';

interface TableSelectProps {
  selectedTable?: string;
  onSelect: (tableName: string, fields: FieldConfig[]) => void;
  onNext: () => void;
}

export default function TableSelect({ selectedTable, onSelect, onNext }: TableSelectProps) {
  const [tables, setTables] = useState<string[]>([]);
  const [loading, setLoading] = useState(false);
  const [metadataLoading, setMetadataLoading] = useState(false);
  const [currentTable, setCurrentTable] = useState(selectedTable || '');
  const [metadata, setMetadata] = useState<FieldConfig[]>([]);

  useEffect(() => {
    loadTables();
  }, []);

  useEffect(() => {
    if (currentTable) {
      loadMetadata(currentTable);
    }
  }, [currentTable]);

  const loadTables = async () => {
    setLoading(true);
    try {
      const data = await getTables();
      setTables(data);
      if (data.length > 0 && !currentTable) {
        setCurrentTable(data[0]);
      }
    } catch (error) {
      message.error('Failed to load tables');
    } finally {
      setLoading(false);
    }
  };

  const loadMetadata = async (tableName: string) => {
    setMetadataLoading(true);
    try {
      const data = await getTableMetadata(tableName);
      setMetadata(data.fields);
    } catch (error) {
      message.error('Failed to load table metadata');
    } finally {
      setMetadataLoading(false);
    }
  };

  const handleNext = () => {
    if (!currentTable) {
      message.warning('Please select a table');
      return;
    }
    onSelect(currentTable, metadata);
    onNext();
  };

  const columns: ColumnsType<FieldConfig> = [
    {
      title: 'Column Name',
      dataIndex: 'columnName',
      key: 'columnName',
    },
    {
      title: 'Field Name',
      dataIndex: 'fieldName',
      key: 'fieldName',
    },
    {
      title: 'Type',
      dataIndex: 'fieldType',
      key: 'fieldType',
    },
    {
      title: 'Comment',
      dataIndex: 'comment',
      key: 'comment',
      ellipsis: true,
    },
  ];

  return (
    <div>
      <Spin spinning={loading}>
        <Card title="Select Table" style={{ marginBottom: 16 }}>
          <Radio.Group
            value={currentTable}
            onChange={e => setCurrentTable(e.target.value)}
            style={{ width: '100%' }}
          >
            <Space direction="vertical" style={{ width: '100%' }}>
              {tables.map(table => (
                <Radio key={table} value={table}>
                  {table}
                </Radio>
              ))}
            </Space>
          </Radio.Group>
        </Card>
      </Spin>

      {currentTable && (
        <Card title="Table Metadata Preview">
          <Spin spinning={metadataLoading}>
            <Table
              columns={columns}
              dataSource={metadata}
              rowKey="columnName"
              pagination={false}
              size="small"
            />
          </Spin>
        </Card>
      )}

      <div style={{ marginTop: 24, textAlign: 'right' }}>
        <Button type="primary" onClick={handleNext} disabled={!currentTable}>
          Next
        </Button>
      </div>
    </div>
  );
}
