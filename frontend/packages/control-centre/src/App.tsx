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
import ConsentRecords from './pages/Consent/ConsentRecords';
import DataSubjects from './pages/DataSubjects';
import AuditLogs from './pages/AuditLogs';
import PIIInventory from './pages/PIIInventory';
import Settings from './pages/Settings';
import Departments from './pages/Departments';
import ThirdParties from './pages/ThirdParties';
import RetentionPolicies from './pages/RetentionPolicies';
import RoPA from './pages/RoPA';
import Reports from './pages/Reports';
import Nominations from './pages/Nominations';
import { Clock } from 'lucide-react';

const queryClient = new QueryClient({
    defaultOptions: {
        queries: {
            retry: 1,
            refetchOnWindowFocus: false,
        },
    },
});

function ComingSoonPage({ title, description }: { title: string; description: string }) {
    return (
        <div style={{
            display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center',
            minHeight: '60vh', padding: '2rem', textAlign: 'center',
        }}>
            <div style={{
                width: 80, height: 80, borderRadius: '50%',
                background: 'linear-gradient(135deg, #eff6ff 0%, #dbeafe 100%)',
                display: 'flex', alignItems: 'center', justifyContent: 'center',
                marginBottom: '24px',
            }}>
                <Clock size={36} style={{ color: '#3b82f6' }} />
            </div>
            <h1 style={{ fontSize: '1.5rem', fontWeight: 700, color: 'var(--text-primary)', marginBottom: '8px' }}>
                {title}
            </h1>
            <p style={{ color: 'var(--text-secondary)', fontSize: '0.9375rem', maxWidth: 400, lineHeight: 1.6, marginBottom: '20px' }}>
                {description}
            </p>
            <span style={{
                display: 'inline-block', padding: '6px 16px', borderRadius: '9999px',
                fontSize: '0.75rem', fontWeight: 600, letterSpacing: '0.04em',
                backgroundColor: '#eff6ff', color: '#2563eb', border: '1px solid #bfdbfe',
            }}>
                Coming in Phase 5
            </span>
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

                        <Route path="/agents" element={<ComingSoonPage title="AI Agents" description="Automated compliance scanning agents that monitor your data landscape in real-time." />} />
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
                        <Route path="/subjects" element={<DataSubjects />} />
                        <Route path="/consent" element={<ConsentRecords />} />
                        <Route path="/nominations" element={<Nominations />} />

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

                        {/* Organization / Third-Parties */}
                        <Route path="/departments" element={<Departments />} />
                        <Route path="/third-parties" element={<ThirdParties />} />
                        <Route path="/retention" element={<RetentionPolicies />} />
                        <Route path="/ropa" element={<RoPA />} />
                        <Route path="/reports" element={<Reports />} />
                        <Route path="/audit-logs" element={<AuditLogs />} />
                        <Route path="/users" element={<ComingSoonPage title="User Management" description="Role-based access control and team management for your organization." />} />
                        <Route path="/settings" element={<Settings />} />
                        <Route path="*" element={<ComingSoonPage title="404 — Page Not Found" description="The page you're looking for doesn't exist or has been moved." />} />
                    </Route>
                </Routes>
            </BrowserRouter>
        </QueryClientProvider>
    );
}

export default App;
