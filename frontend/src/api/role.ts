import request from '../utils/request';

/**
 * Role API definitions
 */

export interface RoleInfo {
  id: number;
  roleName: string;
  roleKey: string;
  dataScope: string;
  sort: number;
  status: boolean;
  remark: string;
  createdAt: string;
  updatedAt: string;
}

// Get role list with pagination
export interface GetRoleListParams {
  page: number;
  pageSize: number;
}

export interface GetRoleListResponse {
  list: RoleInfo[];
  total: number;
}

export const getRoleList = (params: GetRoleListParams): Promise<GetRoleListResponse> => {
  return request.get('/role/list', { params });
};

// Get role by ID
export const getRoleById = (id: number): Promise<RoleInfo> => {
  return request.get(`/role/${id}`);
};

// Create role
export interface CreateRoleRequest {
  roleName: string;
  roleKey: string;
  dataScope?: string;
  sort?: number;
  status?: boolean;
  remark?: string;
}

export const createRole = (data: CreateRoleRequest): Promise<RoleInfo> => {
  return request.post('/role', data);
};

// Update role
export interface UpdateRoleRequest {
  id: number;
  roleName?: string;
  roleKey?: string;
  dataScope?: string;
  sort?: number;
  status?: boolean;
  remark?: string;
}

export const updateRole = (data: UpdateRoleRequest): Promise<RoleInfo> => {
  return request.put(`/role/${data.id}`, data);
};

// Delete role
export const deleteRole = (id: number): Promise<void> => {
  return request.delete(`/role/${id}`);
};

// Assign menus to role
export interface AssignMenusRequest {
  roleId: number;
  menuIds: number[];
}

export const assignMenus = (data: AssignMenusRequest): Promise<void> => {
  return request.post('/role/assign-menus', data);
};

// Get role menus
export const getRoleMenus = (roleId: number): Promise<number[]> => {
  return request.get(`/role/${roleId}/menus`);
};

// Assign APIs to role (Casbin policies)
export interface AssignAPIsRequest {
  roleId: number;
  policies: string[][]; // [[path, method], ...]
}

export const assignAPIs = (data: AssignAPIsRequest): Promise<void> => {
  return request.post('/role/assign-apis', data);
};

// Get role APIs
export const getRoleAPIs = (roleId: number): Promise<string[][]> => {
  return request.get(`/role/${roleId}/apis`);
};
