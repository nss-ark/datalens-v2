import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { ErrorBoundary, GlobalErrorFallback, ToastContainer } from '@datalens/shared';
import { PortalLayout } from './components/PortalLayout';
import { AuthLayout } from './components/AuthLayout';
import { PortalProtectedRoute } from './components/PortalProtectedRoute';

// Pages
import PortalLogin from './pages/Login';
import PortalDashboard from './pages/Dashboard';
import History from './pages/History';
import Requests from './pages/Requests';
import RequestNew from './pages/RequestNew';
import Profile from './pages/Profile';
import ConsentManage from './pages/ConsentManage';
import SubmitGrievance from './pages/Grievance/SubmitGrievance';
import MyGrievances from './pages/Grievance/MyGrievances';
import BreachNotifications from './pages/BreachNotifications';

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
                    {/* Portal Login */}
                    <Route path="/login" element={
                        <ErrorBoundary FallbackComponent={GlobalErrorFallback}>
                            <AuthLayout>
                                <PortalLogin />
                            </AuthLayout>
                        </ErrorBoundary>
                    } />

                    {/* Protected Portal Routes */}
                    <Route element={<PortalProtectedRoute />}>
                        <Route path="/" element={<Navigate to="/dashboard" replace />} />
                        <Route path="/dashboard" element={
                            <ErrorBoundary FallbackComponent={GlobalErrorFallback}>
                                <PortalLayout>
                                    <PortalDashboard />
                                </PortalLayout>
                            </ErrorBoundary>
                        } />
                        <Route path="/history" element={
                            <ErrorBoundary FallbackComponent={GlobalErrorFallback}>
                                <PortalLayout>
                                    <History />
                                </PortalLayout>
                            </ErrorBoundary>
                        } />
                        <Route path="/requests" element={
                            <ErrorBoundary FallbackComponent={GlobalErrorFallback}>
                                <PortalLayout>
                                    <Requests />
                                </PortalLayout>
                            </ErrorBoundary>
                        } />
                        <Route path="/requests/new" element={
                            <ErrorBoundary FallbackComponent={GlobalErrorFallback}>
                                <PortalLayout>
                                    <RequestNew />
                                </PortalLayout>
                            </ErrorBoundary>
                        } />
                        <Route path="/profile" element={
                            <ErrorBoundary FallbackComponent={GlobalErrorFallback}>
                                <PortalLayout>
                                    <Profile />
                                </PortalLayout>
                            </ErrorBoundary>
                        } />
                        <Route path="/consent" element={
                            <ErrorBoundary FallbackComponent={GlobalErrorFallback}>
                                <PortalLayout>
                                    <ConsentManage />
                                </PortalLayout>
                            </ErrorBoundary>
                        } />
                        <Route path="/grievance/new" element={
                            <ErrorBoundary FallbackComponent={GlobalErrorFallback}>
                                <PortalLayout>
                                    <SubmitGrievance />
                                </PortalLayout>
                            </ErrorBoundary>
                        } />
                        <Route path="/grievance/list" element={
                            <ErrorBoundary FallbackComponent={GlobalErrorFallback}>
                                <PortalLayout>
                                    <MyGrievances />
                                </PortalLayout>
                            </ErrorBoundary>
                        } />
                        <Route path="/notifications/breach" element={
                            <ErrorBoundary FallbackComponent={GlobalErrorFallback}>
                                <PortalLayout>
                                    <BreachNotifications />
                                </PortalLayout>
                            </ErrorBoundary>
                        } />
                    </Route>

                    <Route path="*" element={<Navigate to="/login" replace />} />
                </Routes>
            </BrowserRouter >
        </QueryClientProvider >
    );
}

export default App;
