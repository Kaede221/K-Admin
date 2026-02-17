import { Button } from 'antd';
import type { ButtonProps } from 'antd';
import type { ReactNode } from 'react';
import { useUserStore } from '@/store/userStore';

interface AuthButtonProps extends ButtonProps {
  perm: string;
  fallback?: ReactNode;
}

/**
 * AuthButton component - Conditionally renders a button based on user permissions
 * 
 * @param perm - The permission string required to render the button
 * @param fallback - Optional fallback content to render when user lacks permission
 * @param children - Button content
 * @param props - Additional Ant Design Button props
 * 
 * @example
 * // Button only visible to users with 'user:delete' permission
 * <AuthButton perm="user:delete" danger>Delete</AuthButton>
 * 
 * @example
 * // Button with fallback for unauthorized users
 * <AuthButton perm="user:edit" fallback={<span>No permission</span>}>
 *   Edit
 * </AuthButton>
 */
export function AuthButton({ perm, fallback, children, ...props }: AuthButtonProps): ReactNode {
  const hasPermission = useUserStore((state) => state.hasPermission);

  // Check if user has the required permission
  if (!hasPermission(perm)) {
    // Return fallback if provided, otherwise return null (hide button)
    return fallback ?? null;
  }

  // Render button if user has permission
  return <Button {...props}>{children}</Button>;
}

export default AuthButton;
