import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import type { UserInfo } from '@/types/user';
import type { MenuItem } from '@/types/menu';
import request from '@/utils/request';
import { setToken, setRefreshToken, removeToken, removeRefreshToken, removeUserInfo } from '@/utils/storage';

interface UserState {
  userInfo: UserInfo | null;
  accessToken: string;
  refreshToken: string;
  permissions: string[];
  menuTree: MenuItem[];

  // Actions
  login: (username: string, password: string) => Promise<void>;
  logout: () => void;
  refreshAccessToken: () => Promise<void>;
  fetchUserMenu: () => Promise<void>;
  hasPermission: (perm: string) => boolean;
  updateUserInfo: (info: Partial<UserInfo>) => void;
  setTokens: (accessToken: string, refreshToken: string) => void;
  setUserInfo: (userInfo: UserInfo) => void;
}

export const useUserStore = create<UserState>()(
  persist(
    (set, get) => ({
      userInfo: null,
      accessToken: '',
      refreshToken: '',
      permissions: [],
      menuTree: [],

      login: async (username: string, password: string) => {
        const response = await request.post<{
          accessToken: string;
          refreshToken: string;
          user: UserInfo;
        }>('/user/login', {
          username,
          password,
        });

        const { accessToken, refreshToken, user } = response;

        // Store tokens
        setToken(accessToken);
        setRefreshToken(refreshToken);

        // Update store
        set({
          accessToken: accessToken,
          refreshToken: refreshToken,
          userInfo: user,
        });
      },

      logout: () => {
        // Clear tokens and user info
        removeToken();
        removeRefreshToken();
        removeUserInfo();

        // Reset store
        set({
          userInfo: null,
          accessToken: '',
          refreshToken: '',
          permissions: [],
          menuTree: [],
        });

        // Redirect to login
        window.location.href = '/login';
      },

      refreshAccessToken: async () => {
        const { refreshToken } = get();
        if (!refreshToken) {
          throw new Error('No refresh token');
        }

        const response = await request.post<{ accessToken: string }>('/user/refresh', {
          refresh_token: refreshToken,
        });

        const { accessToken } = response;

        // Store new token
        setToken(accessToken);

        // Update store
        set({ accessToken: accessToken });
      },

      fetchUserMenu: async () => {
        const { userInfo } = get();
        if (!userInfo) {
          throw new Error('No user info');
        }

        const menuTree = await request.get<MenuItem[]>(`/menu/tree?roleId=${userInfo.roleId}`);

        // Ensure menuTree is an array (handle null/undefined from backend)
        const safeMenuTree = Array.isArray(menuTree) ? menuTree : [];

        // Extract button permissions from menu tree
        const permissions: string[] = [];
        const extractPermissions = (menus: MenuItem[]) => {
          menus.forEach((menu) => {
            if (menu.btn_perms && menu.btn_perms.length > 0) {
              permissions.push(...menu.btn_perms);
            }
            if (menu.children && menu.children.length > 0) {
              extractPermissions(menu.children);
            }
          });
        };
        extractPermissions(safeMenuTree);

        // Update store
        set({
          menuTree: safeMenuTree,
          permissions,
        });
      },

      hasPermission: (perm: string) => {
        const { permissions } = get();
        return permissions.includes(perm);
      },

      updateUserInfo: (info: Partial<UserInfo>) => {
        const { userInfo } = get();
        if (userInfo) {
          set({ userInfo: { ...userInfo, ...info } });
        }
      },

      setTokens: (accessToken: string, refreshToken: string) => {
        setToken(accessToken);
        setRefreshToken(refreshToken);
        set({ accessToken, refreshToken });
      },

      setUserInfo: (userInfo: UserInfo) => {
        set({ userInfo });
      },
    }),
    {
      name: 'user-storage',
      partialize: (state) => ({
        accessToken: state.accessToken,
        refreshToken: state.refreshToken,
        userInfo: state.userInfo,
        // 不持久化 menuTree 和 permissions，每次刷新重新获取
      }),
    }
  )
);
