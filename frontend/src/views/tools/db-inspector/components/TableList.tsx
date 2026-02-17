import { useState } from 'react';
import { Input, List, Spin } from 'antd';
import { SearchOutlined, TableOutlined } from '@ant-design/icons';

interface TableListProps {
  tables: string[];
  selectedTable: string;
  onSelect: (table: string) => void;
  loading: boolean;
}

export default function TableList({ tables, selectedTable, onSelect, loading }: TableListProps) {
  const [searchText, setSearchText] = useState('');

  const filteredTables = tables.filter(table =>
    table.toLowerCase().includes(searchText.toLowerCase())
  );

  return (
    <div className="table-list">
      <div style={{ padding: '8px 16px' }}>
        <Input
          placeholder="Search tables..."
          prefix={<SearchOutlined />}
          value={searchText}
          onChange={e => setSearchText(e.target.value)}
          allowClear
        />
      </div>
      <Spin spinning={loading}>
        <List
          dataSource={filteredTables}
          renderItem={table => (
            <List.Item
              className={`table-item ${selectedTable === table ? 'selected' : ''}`}
              onClick={() => onSelect(table)}
              style={{
                cursor: 'pointer',
                padding: '12px 16px',
                backgroundColor: selectedTable === table ? '#e6f7ff' : 'transparent',
              }}
            >
              <TableOutlined style={{ marginRight: 8 }} />
              {table}
            </List.Item>
          )}
        />
      </Spin>
    </div>
  );
}
