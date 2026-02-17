import { cloneDeep } from 'lodash';

/**
 * Deep clone an object
 */
export const deepClone = <T>(obj: T): T => {
  return cloneDeep(obj);
};

/**
 * Generate unique ID
 */
export const generateId = (): string => {
  return `${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
};

/**
 * Debounce function
 */
export const debounce = <T extends (...args: any[]) => any>(
  func: T,
  wait: number
): ((...args: Parameters<T>) => void) => {
  let timeout: NodeJS.Timeout | null = null;
  return (...args: Parameters<T>) => {
    if (timeout) clearTimeout(timeout);
    timeout = setTimeout(() => func(...args), wait);
  };
};

/**
 * Throttle function
 */
export const throttle = <T extends (...args: any[]) => any>(
  func: T,
  limit: number
): ((...args: Parameters<T>) => void) => {
  let inThrottle: boolean;
  return (...args: Parameters<T>) => {
    if (!inThrottle) {
      func(...args);
      inThrottle = true;
      setTimeout(() => (inThrottle = false), limit);
    }
  };
};

/**
 * Download file from URL
 */
export const downloadFile = (url: string, filename: string): void => {
  const link = document.createElement('a');
  link.href = url;
  link.download = filename;
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
};

/**
 * Copy text to clipboard
 */
export const copyToClipboard = async (text: string): Promise<boolean> => {
  try {
    await navigator.clipboard.writeText(text);
    return true;
  } catch (error) {
    console.error('Failed to copy to clipboard:', error);
    return false;
  }
};

/**
 * Get query parameter from URL
 */
export const getQueryParam = (param: string): string | null => {
  const urlParams = new URLSearchParams(window.location.search);
  return urlParams.get(param);
};

/**
 * Build tree structure from flat array
 */
export interface TreeNode {
  id: number;
  parent_id: number;
  children?: TreeNode[];
  [key: string]: any;
}

export const buildTree = <T extends TreeNode>(
  items: T[],
  parentId: number = 0
): T[] => {
  return items
    .filter((item) => item.parent_id === parentId)
    .map((item) => ({
      ...item,
      children: buildTree(items, item.id),
    }));
};

/**
 * Flatten tree structure to array
 */
export const flattenTree = <T extends TreeNode>(tree: T[]): T[] => {
  const result: T[] = [];
  const traverse = (nodes: T[]) => {
    nodes.forEach((node) => {
      const { children, ...rest } = node;
      result.push(rest as T);
      if (children && children.length > 0) {
        traverse(children as T[]);
      }
    });
  };
  traverse(tree);
  return result;
};
