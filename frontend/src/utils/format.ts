import dayjs from 'dayjs';

/**
 * Format date to string
 */
export const formatDate = (date: string | Date | number, format = 'YYYY-MM-DD HH:mm:ss'): string => {
  if (!date) return '';
  return dayjs(date).format(format);
};

/**
 * Format number with thousand separators
 */
export const formatNumber = (num: number | string, decimals = 2): string => {
  if (num === null || num === undefined) return '';
  const number = typeof num === 'string' ? parseFloat(num) : num;
  if (isNaN(number)) return '';
  return number.toLocaleString('en-US', {
    minimumFractionDigits: decimals,
    maximumFractionDigits: decimals,
  });
};

/**
 * Format file size
 */
export const formatFileSize = (bytes: number): string => {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${(bytes / Math.pow(k, i)).toFixed(2)} ${sizes[i]}`;
};

/**
 * Format phone number
 */
export const formatPhone = (phone: string): string => {
  if (!phone) return '';
  const cleaned = phone.replace(/\D/g, '');
  const match = cleaned.match(/^(\d{3})(\d{4})(\d{4})$/);
  if (match) {
    return `${match[1]}-${match[2]}-${match[3]}`;
  }
  return phone;
};

/**
 * Truncate text with ellipsis
 */
export const truncate = (text: string, maxLength: number): string => {
  if (!text || text.length <= maxLength) return text;
  return `${text.slice(0, maxLength)}...`;
};
