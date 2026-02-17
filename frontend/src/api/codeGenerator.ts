import request from '../utils/request';

/**
 * Code Generator API definitions
 */

export interface FieldConfig {
  columnName: string;
  fieldName: string;
  fieldType: string;
  jsonTag: string;
  gormTag: string;
  comment: string;
}

export interface TableMetadata {
  tableName: string;
  fields: FieldConfig[];
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

export interface GenerateConfig {
  tableName: string;
  structName: string;
  packageName: string;
  frontendPath: string;
  fields: FieldConfig[];
  options: GenerateOptions;
}

export interface GeneratedFiles {
  [filePath: string]: string; // file path -> file content
}

// Get table metadata
export const getTableMetadata = (tableName: string): Promise<TableMetadata> => {
  return request.get(`/api/v1/tools/gen/metadata/${tableName}`);
};

// Preview generated code
export const previewCode = (config: GenerateConfig): Promise<GeneratedFiles> => {
  return request.post('/api/v1/tools/gen/preview', config);
};

// Generate code and write to files
export const generateCode = (config: GenerateConfig): Promise<GeneratedFiles> => {
  return request.post('/api/v1/tools/gen/generate', config);
};

// Create table from field definitions
export interface CreateTableRequest {
  tableName: string;
  fields: FieldConfig[];
}

export const createTable = (data: CreateTableRequest): Promise<void> => {
  return request.post('/api/v1/tools/gen/create-table', data);
};
