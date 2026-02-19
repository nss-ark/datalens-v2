import type { ReactNode } from 'react';
import { Navigate, useLocation } from 'react-router-dom';
import { useAuthStore } from '@datalens/shared';

interface AdminRouteProps {
    children: ReactNode;
}

export function AdminRoute({ children }: AdminRouteProps) {
    const { isAuthenticated } = useAuthStore();
    const location = useLocation();

    if (!isAuthenticated) {
        return <Navigate to="/login" state={{ from: location }} replace />;
    }

    // TODO: Strictly enforce PLATFORM_ADMIN role check once ID is available in shared constants
    // For now, we rely on backend enforcement for data actions.
    // Ideally:
    // const isAdmin = user?.role_ids.includes('platform-admin-uuid');
    // if (!isAdmin) { return <Navigate to="/unauthorized" replace />; }

    return <>{children}</>;
}
