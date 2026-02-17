import request from '../utils/request';

/**
 * Database Inspector API definitions
 */

export interface ColumnInfo {
  name: string;
  type: string;
  nullable: boolean;
  key: string;
  default: string;
  extra: string;
  comment: string;
}

export interface TableDataResponse {
  data: Record<string, any>[];
  total: number;
}

// Get all tables
export const getTables = (): Promise<string[]> => {
  return request.get('/tools/db/tables');
};

// Get table schema
export const getTableSchema = (tableName: string): Promise<ColumnInfo[]> => {
  return request.get(`/tools/db/schema/${tableName}`);
};

// Get table data with pagination
export interface GetTableDataParams {
  tableName: string;
  page: number;
  pageSize: number;
}

export const getTableData = (params: GetTableDataParams): Promise<TableDataResponse> => {
  return request.get('/tools/db/data', { params });
};

// Execute SQL
export interface ExecuteSQLRequest {
  sql: string;
  readOnly: boolean;
}

export const executeSQL = (data: ExecuteSQLRequest): Promise<any> => {
  return request.post('/tools/db/execute', data);
};

// Create record
export interface CreateRecordRequest {
  tableName: string;
  data: Record<string, any>;
}

export const createRecord = (data: CreateRecordRequest): Promise<void> => {
  return request.post('/tools/db/record', data);
};

// Update record
export interface UpdateRecordRequest {
  tableName: string;
  id: any;
  data: Record<string, any>;
}

export const updateRecord = (data: UpdateRecordRequest): Promise<void> => {
  return request.put(`/tools/db/record/${data.tableName}/${data.id}`, data.data);
};

// Delete record
export const deleteRecord = (tableName: string, id: any): Promise<void> => {
  return request.delete(`/tools/db/record/${tableName}/${id}`);
};
