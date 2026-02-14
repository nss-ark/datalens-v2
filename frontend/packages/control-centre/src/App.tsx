import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { ErrorBoundary, GlobalErrorFallback, ToastContainer } from '@datalens/shared';
import { AppLayout } from './components/Layout/AppLayout';
import { ProtectedRoute } from './components/ProtectedRoute';

// Pages
import Login from './pages/Login';
import Register from './pages/Register';
import Dashboard from './pages/Dashboard';
import DataSources from './pages/DataSources';
import DataSourceDetail from './pages/DataSourceDetail';
import DataSourceConfig from './pages/DataSourceConfig';
import PIIDiscovery from './pages/PIIDiscovery';
import DSRList from './pages/DSRList';
import DSRDetail from './pages/DSRDetail';
import ConsentWidgets from './pages/ConsentWidgets';
import WidgetDetail from './pages/WidgetDetail';
import PurposeMapping from './pages/Governance/PurposeMapping';
import PolicyManager from './pages/Governance/PolicyManager';
import Violations from './pages/Governance/Violations';
import DataLineage from './pages/Governance/DataLineage';
import BreachDashboard from './pages/Breach/BreachDashboard';
import BreachCreate from './pages/Breach/BreachCreate';
import BreachDetail from './pages/Breach/BreachDetail';
import BreachEdit from './pages/Breach/BreachEdit';
import IdentitySettings from './pages/Compliance/IdentitySettings';
import Analytics from './pages/Compliance/Analytics';
import DarkPatternLab from './pages/Compliance/DarkPatternLab';
import NotificationHistory from './pages/Compliance/NotificationHistory';
import GrievanceList from './pages/Compliance/GrievanceList';
import GrievanceDetail from './pages/Compliance/GrievanceDetail';
import NoticeManager from './pages/Consent/NoticeManager';
import PIIInventory from './pages/PIIInventory';
import Settings from './pages/Settings';

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
                    {/* Public routes */}
                    <Route path="/login" element={<Login />} />
                    <Route path="/register" element={<Register />} />

                    {/* Protected routes — Control Centre */}
                    <Route element={
                        <ProtectedRoute>
                            <ErrorBoundary FallbackComponent={GlobalErrorFallback}>
                                <AppLayout />
                            </ErrorBoundary>
                        </ProtectedRoute>
                    }>
                        <Route path="/" element={<Navigate to="/dashboard" replace />} />
                        <Route path="/dashboard" element={<Dashboard />} />
                        <Route path="/datasources" element={<DataSources />} />
                        <Route path="/datasources/:id" element={<DataSourceDetail />} />
                        <Route path="/datasources/:id/config" element={<DataSourceConfig />} />

                        <Route path="/agents" element={<PlaceholderPage title="Agents" />} />
                        <Route path="/pii/review" element={<PIIDiscovery />} />
                        <Route path="/dsr" element={<DSRList />} />
                        <Route path="/dsr/:id" element={<DSRDetail />} />
                        <Route path="/consent/notices" element={<NoticeManager />} />
                        <Route path="/consent/widgets" element={<ConsentWidgets />} />
                        <Route path="/consent/widgets/:id" element={<WidgetDetail />} />
                        <Route path="/breach" element={<BreachDashboard />} />
                        <Route path="/breach/new" element={<BreachCreate />} />
                        <Route path="/breach/:id" element={<BreachDetail />} />
                        <Route path="/breach/:id/edit" element={<BreachEdit />} />

                        <Route path="/pii/inventory" element={<PIIInventory />} />
                        <Route path="/governance/lineage" element={<DataLineage />} />
                        <Route path="/subjects" element={<PlaceholderPage title="Data Subjects" />} />
                        <Route path="/consent" element={<PlaceholderPage title="Consent Records" />} />
                        <Route path="/grievances" element={<PlaceholderPage title="Grievances" />} />
                        <Route path="/nominations" element={<PlaceholderPage title="Nominations" />} />

                        {/* Governance Routes */}
                        <Route path="/governance/purposes" element={<PurposeMapping />} />
                        <Route path="/governance/policies" element={<PolicyManager />} />
                        <Route path="/governance/violations" element={<Violations />} />

                        {/* Compliance Settings */}
                        <Route path="/compliance/settings/identity" element={<IdentitySettings />} />
                        <Route path="/compliance/analytics" element={<Analytics />} />
                        <Route path="/compliance/lab" element={<DarkPatternLab />} />
                        <Route path="/compliance/notifications" element={<NotificationHistory />} />
                        <Route path="/compliance/grievances" element={<GrievanceList />} />
                        <Route path="/compliance/grievances/:id" element={<GrievanceDetail />} />

                        {/* Placeholder routes */}
                        <Route path="/department" element={<PlaceholderPage title="Department" />} />
                        <Route path="/third-parties" element={<PlaceholderPage title="Third Parties" />} />
                        <Route path="/retention" element={<PlaceholderPage title="Retention Policies" />} />
                        <Route path="/ropa" element={<PlaceholderPage title="RoPA" />} />
                        <Route path="/reports" element={<PlaceholderPage title="Reports" />} />
                        <Route path="/audit-logs" element={<PlaceholderPage title="Audit Logs" />} />
                        <Route path="/users" element={<PlaceholderPage title="User Management" />} />
                        <Route path="/settings" element={<Settings />} />
                        <Route path="*" element={<PlaceholderPage title="404 — Page Not Found" />} />
                    </Route>
                </Routes>
            </BrowserRouter>
        </QueryClientProvider>
    );
}

export default App;
