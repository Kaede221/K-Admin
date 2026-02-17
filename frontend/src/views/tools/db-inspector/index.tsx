import { useState, useEffect } from 'react';
import { Layout, Select, message } from 'antd';
import { DatabaseOutlined } from '@ant-design/icons';
import TableList from './components/TableList';
import SchemaViewer from './components/SchemaViewer';
import DataBrowser from './components/DataBrowser';
import SQLConsole from './components/SQLConsole';
import { getTables } from '@/api/dbInspector';
import './index.css';

const { Sider, Content } = Layout;

type ViewMode = 'schema' | 'data' | 'sql';

export default function DBInspector() {
  const [tables, setTables] = useState<string[]>([]);
  const [selectedTable, setSelectedTable] = useState<string>('');
  const [viewMode, setViewMode] = useState<ViewMode>('schema');
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    loadTables();
  }, []);

  const loadTables = async () => {
    setLoading(true);
    try {
      const data = await getTables();
      setTables(data);
      if (data.length > 0) {
        setSelectedTable(data[0]);
      }
    } catch (error) {
      message.error('Failed to load tables');
    } finally {
      setLoading(false);
    }
  };

  const handleTableSelect = (table: string) => {
    setSelectedTable(table);
    setViewMode('schema');
  };

  return (
    <div className="db-inspector">
      <Layout style={{ height: '100%' }}>
        <Sider width={250} theme="light" className="db-inspector-sider">
          <div className="db-header">
            <DatabaseOutlined style={{ fontSize: 20, marginRight: 8 }} />
            <span>Database Tables</span>
          </div>
          <TableList
            tables={tables}
            selectedTable={selectedTable}
            onSelect={handleTableSelect}
            loading={loading}
          />
        </Sider>
        <Content className="db-inspector-content">
          <div className="view-mode-selector">
            <Select
              value={viewMode}
              onChange={setViewMode}
              style={{ width: 150 }}
              options={[
                { label: 'Schema', value: 'schema' },
                { label: 'Data', value: 'data' },
                { label: 'SQL Console', value: 'sql' },
              ]}
            />
          </div>
          <div className="content-area">
            {viewMode === 'schema' && selectedTable && (
              <SchemaViewer tableName={selectedTable} />
            )}
            {viewMode === 'data' && selectedTable && (
              <DataBrowser tableName={selectedTable} />
            )}
            {viewMode === 'sql' && <SQLConsole />}
          </div>
        </Content>
      </Layout>
    </div>
  );
}
