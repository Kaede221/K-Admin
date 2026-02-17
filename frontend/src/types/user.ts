export interface UserInfo {
  id: number;
  username: string;
  nickname: string;
  header_img: string;
  phone: string;
  email: string;
  role_id: number;
  active: boolean;
  role: RoleInfo;
  created_at: string;
  updated_at: string;
}

export interface RoleInfo {
  id: number;
  role_name: string;
  role_key: string;
  data_scope: string;
  sort: number;
  status: boolean;
  remark: string;
  created_at: string;
  updated_at: string;
}

export interface LoginRequest {
  username: string;
  password: string;
}

export interface LoginResponse {
  access_token: string;
  refresh_token: string;
  user: UserInfo;
}
