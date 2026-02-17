import { useState } from 'react';
import { Button, Space, List, Card, message, Result } from 'antd';
import { CheckCircleOutlined, FileOutlined } from '@ant-design/icons';
import { generateCode } from '@/api/codeGenerator';
import { GenerateConfig } from '../index';

interface GenerateConfirmProps {
  config: GenerateConfig;
  generatedCode: Record<string, string>;
  onPrev: () => void;
  onReset: () => void;
}

export default function GenerateConfirm({
  config,
  generatedCode,
  onPrev,
  onReset,
}: GenerateConfirmProps) {
  const [loading, setLoading] = useState(false);
  const [success, setSuccess] = useState(false);

  const handleGenerate = async () => {
    setLoading(true);
    try {
      await generateCode(config);
      setSuccess(true);
      message.success('Code generated successfully!');
    } catch (error) {
      message.error('Failed to generate code');
    } finally {
      setLoading(false);
    }
  };

  if (success) {
    return (
      <Result
        status="success"
        title="Code Generated Successfully!"
        subTitle="All files have been written to the project directories."
        extra={[
          <Button type="primary" key="reset" onClick={onReset}>
            Generate Another
          </Button>,
        ]}
      />
    );
  }

  const fileList = Object.keys(generatedCode);

  return (
    <div>
      <Card title="Files to be Generated" style={{ marginBottom: 16 }}>
        <List
          dataSource={fileList}
          renderItem={file => (
            <List.Item>
              <FileOutlined style={{ marginRight: 8 }} />
              {file}
            </List.Item>
          )}
        />
      </Card>

      <Card title="Configuration Summary">
        <List>
          <List.Item>
            <strong>Table Name:</strong> {config.tableName}
          </List.Item>
          <List.Item>
            <strong>Struct Name:</strong> {config.structName}
          </List.Item>
          <List.Item>
            <strong>Package Name:</strong> {config.packageName}
          </List.Item>
          <List.Item>
            <strong>Frontend Path:</strong> {config.frontendPath}
          </List.Item>
          <List.Item>
            <strong>Total Files:</strong> {fileList.length}
          </List.Item>
        </List>
      </Card>

      <div style={{ marginTop: 24, textAlign: 'right' }}>
        <Space>
          <Button onClick={onPrev}>Previous</Button>
          <Button
            type="primary"
            icon={<CheckCircleOutlined />}
            onClick={handleGenerate}
            loading={loading}
          >
            Generate Code
          </Button>
        </Space>
      </div>
    </div>
  );
}
