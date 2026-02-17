import { Navigate } from 'react-router-dom';
import { useUserStore } from '@/store/userStore';
import type { ReactNode } from 'react';

interface GuardProps {
  children: ReactNode;
}

/**
 * Auth Guard - Protects routes that require authentication
 * Redirects to login if user is not authenticated
 */
export function AuthGuard({ children }: GuardProps) {
  const accessToken = useUserStore((state) => state.accessToken);

  if (!accessToken) {
    return <Navigate to="/login" replace />;
  }

  return <>{children}</>;
}

/**
 * Guest Guard - Protects routes that should only be accessible to guests
 * Redirects to dashboard if user is already authenticated
 */
export function GuestGuard({ children }: GuardProps) {
  const accessToken = useUserStore((state) => state.accessToken);

  if (accessToken) {
    return <Navigate to="/dashboard" replace />;
  }

  return <>{children}</>;
}
