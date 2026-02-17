/**
 * LocalStorage utility with type-safe JSON serialization
 */

export const storage = {
  /**
   * Get item from localStorage with JSON parsing
   */
  getItem<T = any>(key: string): T | null {
    try {
      const item = localStorage.getItem(key);
      if (item === null) {
        return null;
      }
      return JSON.parse(item) as T;
    } catch (error) {
      console.error(`Error getting item ${key} from localStorage:`, error);
      return null;
    }
  },

  /**
   * Set item to localStorage with JSON serialization
   */
  setItem<T = any>(key: string, value: T): void {
    try {
      localStorage.setItem(key, JSON.stringify(value));
    } catch (error) {
      console.error(`Error setting item ${key} to localStorage:`, error);
    }
  },

  /**
   * Remove item from localStorage
   */
  removeItem(key: string): void {
    try {
      localStorage.removeItem(key);
    } catch (error) {
      console.error(`Error removing item ${key} from localStorage:`, error);
    }
  },

  /**
   * Clear all items from localStorage
   */
  clear(): void {
    try {
      localStorage.clear();
    } catch (error) {
      console.error('Error clearing localStorage:', error);
    }
  },
};

// Type-safe wrappers for common keys
export const TOKEN_KEY = 'access_token';
export const REFRESH_TOKEN_KEY = 'refresh_token';
export const USER_INFO_KEY = 'user_info';
export const THEME_KEY = 'theme';

export const getToken = (): string | null => storage.getItem<string>(TOKEN_KEY);
export const setToken = (token: string): void => storage.setItem(TOKEN_KEY, token);
export const removeToken = (): void => storage.removeItem(TOKEN_KEY);

export const getRefreshToken = (): string | null => storage.getItem<string>(REFRESH_TOKEN_KEY);
export const setRefreshToken = (token: string): void => storage.setItem(REFRESH_TOKEN_KEY, token);
export const removeRefreshToken = (): void => storage.removeItem(REFRESH_TOKEN_KEY);

export const getUserInfo = <T = any>(): T | null => storage.getItem<T>(USER_INFO_KEY);
export const setUserInfo = <T = any>(userInfo: T): void => storage.setItem(USER_INFO_KEY, userInfo);
export const removeUserInfo = (): void => storage.removeItem(USER_INFO_KEY);

export const getTheme = (): 'light' | 'dark' | null => storage.getItem<'light' | 'dark'>(THEME_KEY);
export const setTheme = (theme: 'light' | 'dark'): void => storage.setItem(THEME_KEY, theme);
export const removeTheme = (): void => storage.removeItem(THEME_KEY);

export default storage;
