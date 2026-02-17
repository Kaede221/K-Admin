import { Navigate, type NavigateProps } from 'react-router-dom';
import { useUserStore } from '@/store/userStore';
import { getToken } from '@/utils/storage';

/**
 * Authentication guard - Check if user is logged in
 * Redirects to login page if not authenticated
 */
export function AuthGuard({ children }: { children: React.ReactNode }) {
  const token = getToken();

  if (!token) {
    return <Navigate to="/login" replace />;
  }

  return <>{children}</>;
}

/**
 * Authorization guard - Check if user has permission to access route
 * Redirects to 403 page if unauthorized
 */
export function PermissionGuard({
  children,
  requiredPermission,
}: {
  children: React.ReactNode;
  requiredPermission?: string;
}) {
  const hasPermission = useUserStore((state) => state.hasPermission);

  // If no permission required, allow access
  if (!requiredPermission) {
    return <>{children}</>;
  }

  // Check if user has required permission
  if (!hasPermission(requiredPermission)) {
    return <Navigate to="/403" replace />;
  }

  return <>{children}</>;
}

/**
 * Guest guard - Redirect authenticated users away from guest-only pages
 * Used for login page to prevent logged-in users from accessing it
 */
export function GuestGuard({ children }: { children: React.ReactNode }) {
  const token = getToken();

  if (token) {
    return <Navigate to="/dashboard" replace />;
  }

  return <>{children}</>;
}
