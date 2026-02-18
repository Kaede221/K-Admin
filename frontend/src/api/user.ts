import request from '../utils/request';
import type { UserInfo, LoginRequest, LoginResponse } from '../types/user';

/**
 * User API definitions
 */

// Login
export const login = (data: LoginRequest): Promise<LoginResponse> => {
  return request.post('/user/login', data);
};

// Get user info
export const getUserInfo = (id: number): Promise<UserInfo> => {
  return request.get(`/user/${id}`);
};

// Get user list with pagination and filters
export interface GetUserListParams {
  page: number;
  pageSize: number;
  username?: string;
  phone?: string;
  email?: string;
  roleId?: number;
  active?: boolean;
}

export interface GetUserListResponse {
  list: UserInfo[];
  total: number;
}

export const getUserList = (params: GetUserListParams): Promise<GetUserListResponse> => {
  return request.get('/user/list', { params });
};

// Create user
export interface CreateUserRequest {
  username: string;
  password: string;
  nickname?: string;
  phone?: string;
  email?: string;
  roleId: number;
  headerImg?: string;
  active?: boolean;
}

export const createUser = (data: CreateUserRequest): Promise<UserInfo> => {
  return request.post('/user', data);
};

// Update user
export interface UpdateUserRequest {
  id: number;
  username: string;
  password?: string;
  nickname?: string;
  phone?: string;
  email?: string;
  roleId: number;
  headerImg?: string;
  active: boolean;
}

export const updateUser = (data: UpdateUserRequest): Promise<UserInfo> => {
  return request.put('/user', data);
};

// Delete user
export const deleteUser = (id: number): Promise<void> => {
  return request.delete(`/user/${id}`);
};

// Change password
export interface ChangePasswordRequest {
  userId: number;
  oldPassword: string;
  newPassword: string;
}

export const changePassword = (data: ChangePasswordRequest): Promise<void> => {
  return request.post('/user/change-password', data);
};

// Reset password (admin only)
export interface ResetPasswordRequest {
  userId: number;
  newPassword: string;
}

export const resetPassword = (data: ResetPasswordRequest): Promise<void> => {
  return request.post('/user/reset-password', data);
};

// Toggle user status
export const toggleUserStatus = (userId: number, active: boolean): Promise<void> => {
  return request.post('/user/toggle-status', { userId, active });
};
