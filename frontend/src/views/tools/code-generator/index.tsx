import { useState } from 'react';
import { Steps, Card, Button } from 'antd';
import { PlusOutlined } from '@ant-design/icons';
import TableSelect from './components/TableSelect';
import ConfigForm from './components/ConfigForm';
import CodePreview from './components/CodePreview';
import GenerateConfirm from './components/GenerateConfirm';
import TableCreate from './components/TableCreate';
import './index.css';

const { Step } = Steps;

export interface GenerateConfig {
  tableName: string;
  structName: string;
  packageName: string;
  frontendPath: string;
  fields: FieldConfig[];
  options: GenerateOptions;
}

export interface FieldConfig {
  columnName: string;
  fieldName: string;
  fieldType: string;
  jsonTag: string;
  gormTag: string;
  comment: string;
}

export interface GenerateOptions {
  generateModel: boolean;
  generateService: boolean;
  generateAPI: boolean;
  generateRouter: boolean;
  generateFrontendAPI: boolean;
  generateFrontendTypes: boolean;
  generateFrontendPage: boolean;
}

export default function CodeGenerator() {
  const [current, setCurrent] = useState(0);
  const [config, setConfig] = useState<Partial<GenerateConfig>>({
    options: {
      generateModel: true,
      generateService: true,
      generateAPI: true,
      generateRouter: true,
      generateFrontendAPI: true,
      generateFrontendTypes: true,
      generateFrontendPage: true,
    },
  });
  const [generatedCode, setGeneratedCode] = useState<Record<string, string>>({});
  const [createTableVisible, setCreateTableVisible] = useState(false);

  const steps = [
    {
      title: 'Select Table',
      content: (
        <TableSelect
          selectedTable={config.tableName}
          onSelect={(tableName, fields) => {
            setConfig({ ...config, tableName, fields });
          }}
          onNext={() => setCurrent(1)}
        />
      ),
    },
    {
      title: 'Configure',
      content: (
        <ConfigForm
          config={config as GenerateConfig}
          onChange={setConfig}
          onNext={() => setCurrent(2)}
          onPrev={() => setCurrent(0)}
        />
      ),
    },
    {
      title: 'Preview',
      content: (
        <CodePreview
          config={config as GenerateConfig}
          generatedCode={generatedCode}
          onCodeGenerated={setGeneratedCode}
          onNext={() => setCurrent(3)}
          onPrev={() => setCurrent(1)}
        />
      ),
    },
    {
      title: 'Generate',
      content: (
        <GenerateConfirm
          config={config as GenerateConfig}
          generatedCode={generatedCode}
          onPrev={() => setCurrent(2)}
          onReset={() => {
            setCurrent(0);
            setConfig({
              options: {
                generateModel: true,
                generateService: true,
                generateAPI: true,
                generateRouter: true,
                generateFrontendAPI: true,
                generateFrontendTypes: true,
                generateFrontendPage: true,
              },
            });
            setGeneratedCode({});
          }}
        />
      ),
    },
  ];

  return (
    <div className="code-generator">
      <Card
        extra={
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={() => setCreateTableVisible(true)}
          >
            Create Table
          </Button>
        }
      >
        <Steps current={current} style={{ marginBottom: 24 }}>
          {steps.map(item => (
            <Step key={item.title} title={item.title} />
          ))}
        </Steps>
        <div className="steps-content">{steps[current].content}</div>
      </Card>
      <TableCreate
        visible={createTableVisible}
        onClose={() => setCreateTableVisible(false)}
        onSuccess={() => {
          // Optionally reload tables in TableSelect component
        }}
      />
    </div>
  );
}
