import { lazy, type ComponentType } from 'react';
import type { RouteObject } from 'react-router-dom';
import type { MenuItem } from '@/types/menu';

/**
 * Dynamically load component with React.lazy
 * @param componentPath - Component path relative to views directory
 * @returns Lazy-loaded component
 */
export function loadComponent(componentPath: string): ComponentType<any> {
  // Remove leading slash if present
  const path = componentPath.startsWith('/') ? componentPath.slice(1) : componentPath;
  
  // Dynamically import component from views directory
  return lazy(() => import(`../views/${path}`));
}

/**
 * Generate route objects from menu items
 * @param menus - Menu items from backend
 * @returns React Router route objects
 */
export function generateRoutes(menus: MenuItem[]): RouteObject[] {
  const routes: RouteObject[] = [];

  for (const menu of menus) {
    // Skip hidden menus
    if (menu.meta?.hidden) {
      continue;
    }

    const route: RouteObject = {
      path: menu.path,
      element: undefined,
    };

    // Load component if specified
    if (menu.component && menu.component !== 'Layout') {
      try {
        const Component = loadComponent(menu.component);
        route.element = <Component />;
      } catch (error) {
        console.error(`Failed to load component: ${menu.component}`, error);
      }
    }

    // Recursively handle children
    if (menu.children && menu.children.length > 0) {
      route.children = generateRoutes(menu.children);
    }

    routes.push(route);
  }

  return routes;
}

/**
 * Flatten menu tree to get all paths
 * @param menus - Menu items
 * @returns Array of all menu paths
 */
export function flattenMenuPaths(menus: MenuItem[]): string[] {
  const paths: string[] = [];

  for (const menu of menus) {
    if (!menu.meta?.hidden) {
      paths.push(menu.path);
    }

    if (menu.children && menu.children.length > 0) {
      paths.push(...flattenMenuPaths(menu.children));
    }
  }

  return paths;
}
