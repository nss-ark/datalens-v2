import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { ErrorBoundary, GlobalErrorFallback, ToastContainer } from '@datalens/shared';
import { AdminLayout } from './components/Layout/AdminLayout';
import { AdminRoute } from './components/AdminRoute';

// Pages
import Login from './pages/Login';
import AdminDashboard from './pages/Dashboard';
import TenantList from './pages/Tenants/TenantList';
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

function PlaceholderPage({ title }: { title: string }) {
    return (
        <div style={{ padding: '2rem' }}>
            <h1 style={{ fontSize: '1.5rem', fontWeight: 700, color: 'var(--text-primary)', marginBottom: '0.5rem' }}>
                {title}
            </h1>
            <p style={{ color: 'var(--text-secondary)' }}>
                This page is under construction.
            </p>
        </div>
    );
}

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
                        <Route path="/users" element={<UserList />} />
                        <Route path="/compliance/dsr" element={<AdminDSRList />} />
                        <Route path="/compliance/dsr/:id" element={<AdminDSRDetail />} />
                        <Route path="/settings" element={<PlaceholderPage title="Platform Settings" />} />
                    </Route>

                    <Route path="*" element={<Navigate to="/dashboard" replace />} />
                </Routes>
            </BrowserRouter>
        </QueryClientProvider>
    );
}

export default App;
