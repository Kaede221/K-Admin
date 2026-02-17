import { useState, useImperativeHandle, forwardRef, useEffect, type ReactNode } from 'react';
import { Table, Form, Button, Space, Card } from 'antd';
import { SearchOutlined, ReloadOutlined, DownloadOutlined } from '@ant-design/icons';
import type { TableProps, FormInstance } from 'antd';
import type { TablePaginationConfig } from 'antd/es/table';

export interface ProTableColumn<T = any> extends Omit<TableProps<T>['columns'][number], 'dataIndex'> {
  dataIndex?: string;
  hideInSearch?: boolean;
  hideInTable?: boolean;
  valueType?: 'text' | 'select' | 'date' | 'dateRange';
  valueEnum?: Record<string, { text: string; status?: string }>;
  renderFormItem?: () => ReactNode;
}

export interface ProTableProps<T = any> {
  columns: ProTableColumn<T>[];
  request: (params: any) => Promise<{ data: T[]; total: number }>;
  rowKey?: string | ((record: T) => string);
  toolbar?: ReactNode;
  search?: boolean;
  pagination?: false | TablePaginationConfig;
  onExport?: () => void;
}

export interface ProTableActionRef {
  reload: () => void;
  reset: () => void;
}

function ProTableComponent<T extends Record<string, any>>(
  props: ProTableProps<T>,
  ref: React.Ref<ProTableActionRef>
) {
  const {
    columns,
    request,
    rowKey = 'id',
    toolbar,
    search = true,
    pagination: paginationProp = {},
    onExport,
  } = props;

  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [dataSource, setDataSource] = useState<T[]>([]);
  const [pagination, setPagination] = useState<TablePaginationConfig>({
    current: 1,
    pageSize: 10,
    total: 0,
    showSizeChanger: true,
    showQuickJumper: true,
    showTotal: (total) => `共 ${total} 条`,
    ...paginationProp,
  });

  const fetchData = async (params?: any) => {
    setLoading(true);
    try {
      const searchParams = form.getFieldsValue();
      const result = await request({
        page: pagination.current,
        page_size: pagination.pageSize,
        ...searchParams,
        ...params,
      });

      setDataSource(result.data);
      setPagination((prev) => ({
        ...prev,
        total: result.total,
      }));
    } catch (error) {
      console.error('Failed to fetch data:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleSearch = () => {
    setPagination((prev) => ({ ...prev, current: 1 }));
    fetchData({ page: 1 });
  };

  const handleReset = () => {
    form.resetFields();
    setPagination((prev) => ({ ...prev, current: 1 }));
    fetchData({ page: 1 });
  };

  const handleTableChange = (newPagination: TablePaginationConfig) => {
    setPagination(newPagination);
    fetchData({
      page: newPagination.current,
      page_size: newPagination.pageSize,
    });
  };

  useImperativeHandle(ref, () => ({
    reload: () => fetchData(),
    reset: handleReset,
  }));

  // Fetch data on mount
  useEffect(() => {
    fetchData();
  }, []);

  // Filter columns for search form
  const searchColumns = columns.filter((col) => !col.hideInSearch);
  
  // Filter columns for table
  const tableColumns = columns.filter((col) => !col.hideInTable);

  return (
    <Card>
      {search && searchColumns.length > 0 && (
        <Form
          form={form}
          layout="inline"
          style={{ marginBottom: 16 }}
          onFinish={handleSearch}
        >
          {searchColumns.map((col) => (
            <Form.Item
              key={col.dataIndex as string}
              name={col.dataIndex}
              label={col.title as string}
            >
              {col.renderFormItem ? col.renderFormItem() : <input />}
            </Form.Item>
          ))}
          <Form.Item>
            <Space>
              <Button type="primary" icon={<SearchOutlined />} htmlType="submit">
                搜索
              </Button>
              <Button icon={<ReloadOutlined />} onClick={handleReset}>
                重置
              </Button>
            </Space>
          </Form.Item>
        </Form>
      )}

      <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between' }}>
        <Space>{toolbar}</Space>
        <Space>
          <Button icon={<ReloadOutlined />} onClick={() => fetchData()}>
            刷新
          </Button>
          {onExport && (
            <Button icon={<DownloadOutlined />} onClick={onExport}>
              导出
            </Button>
          )}
        </Space>
      </div>

      <Table
        rowKey={rowKey}
        columns={tableColumns}
        dataSource={dataSource}
        loading={loading}
        pagination={paginationProp === false ? false : pagination}
        onChange={handleTableChange}
      />
    </Card>
  );
}

export const ProTable = forwardRef(ProTableComponent) as <T extends Record<string, any>>(
  props: ProTableProps<T> & { ref?: React.Ref<ProTableActionRef> }
) => React.ReactElement;

export default ProTable;
