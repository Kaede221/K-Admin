import { useState, useEffect } from 'react';
import { Tabs, Button, Space, message, Spin } from 'antd';
import Editor from '@monaco-editor/react';
import { previewCode } from '@/api/codeGenerator';
import { GenerateConfig } from '../index';

interface CodePreviewProps {
  config: GenerateConfig;
  generatedCode: Record<string, string>;
  onCodeGenerated: (code: Record<string, string>) => void;
  onNext: () => void;
  onPrev: () => void;
}

export default function CodePreview({
  config,
  generatedCode,
  onCodeGenerated,
  onNext,
  onPrev,
}: CodePreviewProps) {
  const [loading, setLoading] = useState(false);
  const [code, setCode] = useState(generatedCode);

  useEffect(() => {
    if (Object.keys(code).length === 0) {
      loadPreview();
    }
  }, []);

  const loadPreview = async () => {
    setLoading(true);
    try {
      const data = await previewCode(config);
      setCode(data);
      onCodeGenerated(data);
      message.success('Code preview generated successfully');
    } catch (error) {
      message.error('Failed to generate code preview');
    } finally {
      setLoading(false);
    }
  };

  const getLanguage = (filePath: string): string => {
    if (filePath.endsWith('.go')) return 'go';
    if (filePath.endsWith('.ts') || filePath.endsWith('.tsx')) return 'typescript';
    if (filePath.endsWith('.js') || filePath.endsWith('.jsx')) return 'javascript';
    return 'plaintext';
  };

  const tabItems = Object.entries(code).map(([filePath, content]) => ({
    key: filePath,
    label: filePath.split('/').pop() || filePath,
    children: (
      <div style={{ border: '1px solid #d9d9d9', borderRadius: 4 }}>
        <Editor
          height="500px"
          language={getLanguage(filePath)}
          value={content}
          theme="vs-light"
          options={{
            readOnly: true,
            minimap: { enabled: false },
            fontSize: 14,
            lineNumbers: 'on',
            scrollBeyondLastLine: false,
            automaticLayout: true,
          }}
        />
      </div>
    ),
  }));

  return (
    <Spin spinning={loading}>
      <div>
        {tabItems.length > 0 ? (
          <Tabs items={tabItems} />
        ) : (
          <div style={{ textAlign: 'center', padding: 40 }}>
            <p>No code generated yet</p>
            <Button type="primary" onClick={loadPreview}>
              Generate Preview
            </Button>
          </div>
        )}

        <div style={{ marginTop: 24, textAlign: 'right' }}>
          <Space>
            <Button onClick={onPrev}>Previous</Button>
            <Button type="primary" onClick={onNext} disabled={tabItems.length === 0}>
              Next
            </Button>
          </Space>
        </div>
      </div>
    </Spin>
  );
}
