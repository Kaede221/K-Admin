import { Form, Input, Button, Space, Checkbox, Table, Card, message } from 'antd';
import { GenerateConfig, FieldConfig } from '../index';
import type { ColumnsType } from 'antd/es/table';

interface ConfigFormProps {
  config: GenerateConfig;
  onChange: (config: Partial<GenerateConfig>) => void;
  onNext: () => void;
  onPrev: () => void;
}

export default function ConfigForm({ config, onChange, onNext, onPrev }: ConfigFormProps) {
  const [form] = Form.useForm();

  const handleNext = async () => {
    try {
      const values = await form.validateFields();
      onChange({ ...config, ...values });
      onNext();
    } catch (error) {
      message.error('Please fill in all required fields');
    }
  };

  const handleFieldChange = (index: number, field: keyof FieldConfig, value: string) => {
    const newFields = [...(config.fields || [])];
    newFields[index] = { ...newFields[index], [field]: value };
    onChange({ ...config, fields: newFields });
  };

  const columns: ColumnsType<FieldConfig> = [
    {
      title: 'Column Name',
      dataIndex: 'columnName',
      key: 'columnName',
      width: 150,
    },
    {
      title: 'Field Name',
      dataIndex: 'fieldName',
      key: 'fieldName',
      width: 150,
      render: (text, record, index) => (
        <Input
          value={text}
          onChange={e => handleFieldChange(index, 'fieldName', e.target.value)}
          size="small"
        />
      ),
    },
    {
      title: 'Field Type',
      dataIndex: 'fieldType',
      key: 'fieldType',
      width: 120,
      render: (text, record, index) => (
        <Input
          value={text}
          onChange={e => handleFieldChange(index, 'fieldType', e.target.value)}
          size="small"
        />
      ),
    },
    {
      title: 'JSON Tag',
      dataIndex: 'jsonTag',
      key: 'jsonTag',
      width: 120,
      render: (text, record, index) => (
        <Input
          value={text}
          onChange={e => handleFieldChange(index, 'jsonTag', e.target.value)}
          size="small"
        />
      ),
    },
    {
      title: 'GORM Tag',
      dataIndex: 'gormTag',
      key: 'gormTag',
      width: 200,
      render: (text, record, index) => (
        <Input
          value={text}
          onChange={e => handleFieldChange(index, 'gormTag', e.target.value)}
          size="small"
        />
      ),
    },
    {
      title: 'Comment',
      dataIndex: 'comment',
      key: 'comment',
      render: (text, record, index) => (
        <Input
          value={text}
          onChange={e => handleFieldChange(index, 'comment', e.target.value)}
          size="small"
        />
      ),
    },
  ];

  return (
    <div>
      <Card title="Basic Configuration" style={{ marginBottom: 16 }}>
        <Form
          form={form}
          layout="vertical"
          initialValues={{
            structName: config.structName || '',
            packageName: config.packageName || 'business',
            frontendPath: config.frontendPath || 'views/business',
          }}
        >
          <Form.Item
            label="Struct Name"
            name="structName"
            rules={[{ required: true, message: 'Please enter struct name' }]}
          >
            <Input placeholder="e.g., Product" />
          </Form.Item>
          <Form.Item
            label="Package Name"
            name="packageName"
            rules={[{ required: true, message: 'Please enter package name' }]}
          >
            <Input placeholder="e.g., business" />
          </Form.Item>
          <Form.Item
            label="Frontend Path"
            name="frontendPath"
            rules={[{ required: true, message: 'Please enter frontend path' }]}
          >
            <Input placeholder="e.g., views/business" />
          </Form.Item>
        </Form>
      </Card>

      <Card title="Field Mapping" style={{ marginBottom: 16 }}>
        <Table
          columns={columns}
          dataSource={config.fields || []}
          rowKey="columnName"
          pagination={false}
          scroll={{ x: 'max-content' }}
          size="small"
        />
      </Card>

      <Card title="Generation Options">
        <Checkbox.Group
          value={Object.keys(config.options || {}).filter(
            key => config.options?.[key as keyof typeof config.options]
          )}
          onChange={values => {
            const options = {
              generateModel: values.includes('generateModel'),
              generateService: values.includes('generateService'),
              generateAPI: values.includes('generateAPI'),
              generateRouter: values.includes('generateRouter'),
              generateFrontendAPI: values.includes('generateFrontendAPI'),
              generateFrontendTypes: values.includes('generateFrontendTypes'),
              generateFrontendPage: values.includes('generateFrontendPage'),
            };
            onChange({ ...config, options });
          }}
        >
          <Space direction="vertical">
            <Checkbox value="generateModel">Generate Model</Checkbox>
            <Checkbox value="generateService">Generate Service</Checkbox>
            <Checkbox value="generateAPI">Generate API</Checkbox>
            <Checkbox value="generateRouter">Generate Router</Checkbox>
            <Checkbox value="generateFrontendAPI">Generate Frontend API</Checkbox>
            <Checkbox value="generateFrontendTypes">Generate Frontend Types</Checkbox>
            <Checkbox value="generateFrontendPage">Generate Frontend Page</Checkbox>
          </Space>
        </Checkbox.Group>
      </Card>

      <div style={{ marginTop: 24, textAlign: 'right' }}>
        <Space>
          <Button onClick={onPrev}>Previous</Button>
          <Button type="primary" onClick={handleNext}>
            Next
          </Button>
        </Space>
      </div>
    </div>
  );
}
