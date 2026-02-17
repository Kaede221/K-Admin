import request from '@/utils/request';

export interface DashboardStats {
  userCount: number;
  roleCount: number;
  menuCount: number;
  configCount: number;
}

/**
 * 获取仪表盘统计数据
 */
export const getDashboardStats = () => {
  return request.get<DashboardStats>('/dashboard/stats');
};
