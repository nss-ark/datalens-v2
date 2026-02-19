import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { ErrorBoundary, GlobalErrorFallback, ToastContainer } from '@datalens/shared';
import { AdminLayout } from './components/Layout/AdminLayout';
import { AdminRoute } from './components/AdminRoute';

// Pages
import Login from './pages/Login';
import AdminDashboard from './pages/Dashboard';
import TenantList from './pages/Tenants/TenantList';
import TenantDetail from '@/pages/Tenants/TenantDetail';
import RetentionPolicies from '@/pages/Retention/RetentionPolicies';
import PlatformSettingsPage from '@/pages/Settings/PlatformSettings';
import UserList from './pages/Users/UserList';
import AdminDSRList from './pages/Compliance/DSRList';
import AdminDSRDetail from './pages/Compliance/DSRDetail';

const queryClient = new QueryClient({
    defaultOptions: {
        queries: {
            retry: 1,
            refetchOnWindowFocus: false,
        },
    },
});



function App() {
    return (
        <QueryClientProvider client={queryClient}>
            <ToastContainer />
            <BrowserRouter>
                <Routes>
                    {/* Public route */}
                    <Route path="/login" element={<Login />} />

                    {/* Admin routes */}
                    <Route element={
                        <AdminRoute>
                            <ErrorBoundary FallbackComponent={GlobalErrorFallback}>
                                <AdminLayout />
                            </ErrorBoundary>
                        </AdminRoute>
                    }>
                        <Route index element={<Navigate to="/dashboard" replace />} />
                        <Route path="/dashboard" element={<AdminDashboard />} />
                        <Route path="/tenants" element={<TenantList />} />
                        <Route path="/tenants/:id" element={<TenantDetail />} />
                        <Route path="/users" element={<UserList />} />
                        <Route path="/compliance/dsr" element={<AdminDSRList />} />
                        <Route path="/compliance/dsr/:id" element={<AdminDSRDetail />} />
                        <Route path="/retention-policies" element={<RetentionPolicies />} />
                        <Route path="/settings" element={<PlatformSettingsPage />} />
                    </Route>

                    <Route path="*" element={<Navigate to="/dashboard" replace />} />
                </Routes>
            </BrowserRouter>
        </QueryClientProvider>
    );
}

export default App;
