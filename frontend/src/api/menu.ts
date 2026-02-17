import request from '../utils/request';
import type { MenuItem } from '../types/menu';

/**
 * Menu API definitions
 */

// Get menu tree for current user
export const getMenuTree = (): Promise<MenuItem[]> => {
  return request.get('/menu/tree');
};

// Get all menus (admin only)
export const getAllMenus = (): Promise<MenuItem[]> => {
  return request.get('/menu/all');
};

// Get menu by ID
export const getMenuById = (id: number): Promise<MenuItem> => {
  return request.get(`/menu/${id}`);
};

// Create menu
export interface CreateMenuRequest {
  parentId: number;
  path: string;
  name: string;
  component: string;
  sort: number;
  meta: {
    icon: string;
    title: string;
    hidden: boolean;
    keepAlive: boolean;
  };
  btnPerms?: string[];
}

export const createMenu = (data: CreateMenuRequest): Promise<MenuItem> => {
  return request.post('/menu', data);
};

// Update menu
export interface UpdateMenuRequest {
  id: number;
  parentId?: number;
  path?: string;
  name?: string;
  component?: string;
  sort?: number;
  meta?: {
    icon?: string;
    title?: string;
    hidden?: boolean;
    keepAlive?: boolean;
  };
  btnPerms?: string[];
}

export const updateMenu = (data: UpdateMenuRequest): Promise<MenuItem> => {
  return request.put(`/menu/${data.id}`, data);
};

// Delete menu
export const deleteMenu = (id: number): Promise<void> => {
  return request.delete(`/menu/${id}`);
};
