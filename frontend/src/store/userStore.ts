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
        try {
          const response = await request.post<{
            access_token: string;
            refresh_token: string;
            user: UserInfo;
          }>('/v1/system/user/login', {
            username,
            password,
          });

          const { access_token, refresh_token, user } = response;

          // Store tokens
          setToken(access_token);
          setRefreshToken(refresh_token);

          // Update store
          set({
            accessToken: access_token,
            refreshToken: refresh_token,
            userInfo: user,
          });

          // Fetch user menu
          await get().fetchUserMenu();
        } catch (error) {
          console.error('Login failed:', error);
          throw error;
        }
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
        try {
          const { refreshToken } = get();
          if (!refreshToken) {
            throw new Error('No refresh token');
          }

          const response = await request.post<{ access_token: string }>('/v1/system/user/refresh', {
            refresh_token: refreshToken,
          });

          const { access_token } = response;

          // Store new token
          setToken(access_token);

          // Update store
          set({ accessToken: access_token });
        } catch (error) {
          console.error('Token refresh failed:', error);
          get().logout();
          throw error;
        }
      },

      fetchUserMenu: async () => {
        try {
          const { userInfo } = get();
          if (!userInfo) {
            throw new Error('No user info');
          }

          const menuTree = await request.get<MenuItem[]>(`/v1/system/menu/tree?role_id=${userInfo.role_id}`);

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
          extractPermissions(menuTree);

          // Update store
          set({
            menuTree,
            permissions,
          });
        } catch (error) {
          console.error('Fetch user menu failed:', error);
          throw error;
        }
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
      }),
    }
  )
);
