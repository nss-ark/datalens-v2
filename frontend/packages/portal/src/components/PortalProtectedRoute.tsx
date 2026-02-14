import { Navigate, Outlet } from 'react-router-dom';
import { usePortalAuthStore } from '@/stores/portalAuthStore';

export const PortalProtectedRoute = ({ children }: { children?: React.ReactNode }) => {
    const isAuthenticated = usePortalAuthStore((state) => state.isAuthenticated);

    if (!isAuthenticated) {
        return <Navigate to="/portal/login" replace />;
    }

    return children ? <>{children}</> : <Outlet />;
};
