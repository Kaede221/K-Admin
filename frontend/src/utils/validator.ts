/**
 * Form validation rules
 */

/**
 * Validate email format
 */
export const isEmail = (email: string): boolean => {
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  return emailRegex.test(email);
};

/**
 * Validate phone number (Chinese format)
 */
export const isPhone = (phone: string): boolean => {
  const phoneRegex = /^1[3-9]\d{9}$/;
  return phoneRegex.test(phone);
};

/**
 * Validate URL format
 */
export const isURL = (url: string): boolean => {
  try {
    new URL(url);
    return true;
  } catch {
    return false;
  }
};

/**
 * Validate password strength
 * At least 8 characters, including uppercase, lowercase, and number
 */
export const isStrongPassword = (password: string): boolean => {
  const passwordRegex = /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)[a-zA-Z\d@$!%*?&]{8,}$/;
  return passwordRegex.test(password);
};

/**
 * Validate username format
 * 4-20 characters, alphanumeric and underscore only
 */
export const isUsername = (username: string): boolean => {
  const usernameRegex = /^[a-zA-Z0-9_]{4,20}$/;
  return usernameRegex.test(username);
};

/**
 * Ant Design form validation rules
 */
export const formRules = {
  required: { required: true, message: 'This field is required' },
  email: {
    validator: (_: any, value: string) => {
      if (!value || isEmail(value)) {
        return Promise.resolve();
      }
      return Promise.reject(new Error('Please enter a valid email'));
    },
  },
  phone: {
    validator: (_: any, value: string) => {
      if (!value || isPhone(value)) {
        return Promise.resolve();
      }
      return Promise.reject(new Error('Please enter a valid phone number'));
    },
  },
  url: {
    validator: (_: any, value: string) => {
      if (!value || isURL(value)) {
        return Promise.resolve();
      }
      return Promise.reject(new Error('Please enter a valid URL'));
    },
  },
  strongPassword: {
    validator: (_: any, value: string) => {
      if (!value || isStrongPassword(value)) {
        return Promise.resolve();
      }
      return Promise.reject(
        new Error('Password must be at least 8 characters with uppercase, lowercase, and number')
      );
    },
  },
  username: {
    validator: (_: any, value: string) => {
      if (!value || isUsername(value)) {
        return Promise.resolve();
      }
      return Promise.reject(
        new Error('Username must be 4-20 characters, alphanumeric and underscore only')
      );
    },
  },
};
