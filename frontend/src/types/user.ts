export interface UserInfo {
  id: number;
  username: string;
  nickname: string;
  headerImg: string;
  phone: string;
  email: string;
  roleId: number;
  active: boolean;
  role?: RoleInfo;
  createdAt: string;
  updatedAt: string;
}

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

export interface LoginRequest {
  username: string;
  password: string;
}

export interface LoginResponse {
  accessToken: string;
  refreshToken: string;
  user: UserInfo;
}
