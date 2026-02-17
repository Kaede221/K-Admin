import { useState } from 'react';
import { Button, Space, Switch, Table, message, Alert, Modal } from 'antd';
import { PlayCircleOutlined, ClearOutlined, ExclamationCircleOutlined } from '@ant-design/icons';
import Editor from '@monaco-editor/react';
import { executeSQL } from '@/api/dbInspector';

export default function SQLConsole() {
  const [sql, setSql] = useState('');
  const [readOnly, setReadOnly] = useState(true);
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<any>(null);
  const [error, setError] = useState<string>('');

  const isDangerousSQL = (sql: string): boolean => {
    const dangerous = ['DROP', 'TRUNCATE', 'ALTER'];
    const upperSQL = sql.toUpperCase();
    return dangerous.some(keyword => upperSQL.includes(keyword));
  };

  const handleExecute = async () => {
    if (!sql.trim()) {
      message.warning('Please enter SQL statement');
      return;
    }

    if (isDangerousSQL(sql)) {
      Modal.confirm({
        title: 'Dangerous Operation',
        icon: <ExclamationCircleOutlined />,
        content: 'This SQL contains dangerous operations (DROP, TRUNCATE, ALTER). Are you sure to execute?',
        okText: 'Execute',
        okType: 'danger',
        cancelText: 'Cancel',
        onOk: () => executeQuery(),
      });
    } else {
      executeQuery();
    }
  };

  const executeQuery = async () => {
    setLoading(true);
    setError('');
    setResult(null);
    try {
      const data = await executeSQL({ sql, readOnly });
      setResult(data);
      message.success('SQL executed successfully');
    } catch (err: any) {
      const errorMsg = err.response?.data?.msg || err.message || 'SQL execution failed';
      setError(errorMsg);
      message.error('SQL execution failed');
    } finally {
      setLoading(false);
    }
  };

  const handleClear = () => {
    setSql('');
    setResult(null);
    setError('');
  };

  const renderResult = () => {
    if (error) {
      return <Alert message="Error" description={error} type="error" showIcon />;
    }

    if (!result) {
      return null;
    }

    if (Array.isArray(result)) {
      if (result.length === 0) {
        return <Alert message="Query returned no results" type="info" showIcon />;
      }

      const columns = Object.keys(result[0]).map(key => ({
        title: key,
        dataIndex: key,
        key,
        ellipsis: true,
      }));

      return (
        <Table
          columns={columns}
          dataSource={result}
          rowKey={(_, index) => index?.toString() || '0'}
          scroll={{ x: 'max-content' }}
          pagination={{ pageSize: 20 }}
        />
      );
    }

    return <Alert message="Success" description={JSON.stringify(result)} type="success" showIcon />;
  };

  return (
    <div className="sql-console">
      <div style={{ marginBottom: 16 }}>
        <Space>
          <Button
            type="primary"
            icon={<PlayCircleOutlined />}
            onClick={handleExecute}
            loading={loading}
          >
            Execute
          </Button>
          <Button icon={<ClearOutlined />} onClick={handleClear}>
            Clear
          </Button>
          <span style={{ marginLeft: 16 }}>
            Read-Only Mode:
            <Switch
              checked={readOnly}
              onChange={setReadOnly}
              style={{ marginLeft: 8 }}
            />
          </span>
        </Space>
      </div>
      <div style={{ marginBottom: 16, border: '1px solid #d9d9d9', borderRadius: 4 }}>
        <Editor
          height="300px"
          defaultLanguage="sql"
          value={sql}
          onChange={(value) => setSql(value || '')}
          theme="vs-light"
          options={{
            minimap: { enabled: false },
            fontSize: 14,
            lineNumbers: 'on',
            scrollBeyondLastLine: false,
            automaticLayout: true,
          }}
        />
      </div>
      {!readOnly && (
        <Alert
          message="Warning"
          description="Read-only mode is disabled. Be careful with write operations!"
          type="warning"
          showIcon
          style={{ marginBottom: 16 }}
        />
      )}
      <div className="sql-result">{renderResult()}</div>
    </div>
  );
}
