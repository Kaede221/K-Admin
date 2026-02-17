import { createBrowserRouter, type RouteObject, Navigate } from 'react-router-dom';
import { lazy } from 'react';
import { AuthGuard, GuestGuard } from './guards';
import { generateRoutes } from './generator';
import { useUserStore } from '@/store/userStore';

// Lazy load pages
const Login = lazy(() => import('@/views/login'));
const NotFound = lazy(() => import('@/views/error/404'));
const Forbidden = lazy(() => import('@/views/error/403'));
const Dashboard = lazy(() => import('@/views/dashboard'));
const Layout = lazy(() => import('@/layout'));

/**
 * Static routes - Always available
 */
const staticRoutes: RouteObject[] = [
  {
    path: '/login',
    element: (
      <GuestGuard>
        <Login />
      </GuestGuard>
    ),
  },
  {
    path: '/403',
    element: <Forbidden />,
  },
  {
    path: '/404',
    element: <NotFound />,
  },
];

/**
 * Create router with static and dynamic routes
 * Dynamic routes are generated from user's menu tree
 */
export function createAppRouter() {
  const menuTree = useUserStore.getState().menuTree;
  
  // Generate dynamic routes from menu tree
  const dynamicRoutes = menuTree.length > 0 ? generateRoutes(menuTree) : [];
  
  // Add dashboard as a fallback route if not in dynamic routes
  const hasDashboard = dynamicRoutes.some(route => route.path === '/dashboard' || route.path === 'dashboard');
  if (!hasDashboard) {
    dynamicRoutes.unshift({
      path: 'dashboard',
      element: <Dashboard />,
    });
  }
  
  // Add index route to redirect to dashboard
  dynamicRoutes.unshift({
    index: true,
    element: <Navigate to="/dashboard" replace />,
  });

  // Protected routes wrapped in Layout and AuthGuard
  const protectedRoutes: RouteObject = {
    path: '/',
    element: (
      <AuthGuard>
        <Layout />
      </AuthGuard>
    ),
    children: dynamicRoutes,
  };

  // Combine all routes
  const routes: RouteObject[] = [
    ...staticRoutes,
    protectedRoutes,
    {
      path: '*',
      element: <NotFound />,
    },
  ];

  return createBrowserRouter(routes);
}

/**
 * Default router instance
 * This will be recreated after user logs in and menu is fetched
 */
export const router = createAppRouter();
