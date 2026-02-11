import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { AppLayout } from './components/Layout/AppLayout';
import { ProtectedRoute } from './components/common/ProtectedRoute';
import { ToastContainer } from './components/common/Toast';
import { ErrorBoundary } from './components/common/ErrorBoundary';
import { GlobalErrorFallback } from './components/common/ErrorFallbacks';
import Login from './pages/Login';
import Register from './pages/Register';
import Dashboard from './pages/Dashboard';
import DataSources from './pages/DataSources';
import PIIDiscovery from './pages/PIIDiscovery';
import DSRList from './pages/DSRList';
import DSRDetail from './pages/DSRDetail';
import ConsentWidgets from './pages/ConsentWidgets';
import WidgetDetail from './pages/WidgetDetail';
import PurposeMapping from './pages/Governance/PurposeMapping';
import PolicyManager from './pages/Governance/PolicyManager';
import Violations from './pages/Governance/Violations';

// Portal Components
import { PortalLayout } from './components/Layout/PortalLayout';
import { PortalProtectedRoute } from './components/Portal/PortalProtectedRoute';
import PortalLogin from './pages/Portal/Login';
import PortalDashboard from './pages/Portal/Dashboard';
import PortalHistory from './pages/Portal/History';
import PortalRequests from './pages/Portal/Requests';

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
          {/* Public routes */}
          <Route path="/login" element={<Login />} />
          <Route path="/register" element={<Register />} />

          {/* Portal Routes - Standalone Layout */}
          <Route path="/portal/login" element={
            <ErrorBoundary FallbackComponent={GlobalErrorFallback}>
              <PortalLayout>
                <PortalLogin />
              </PortalLayout>
            </ErrorBoundary>
          } />

          <Route element={<PortalProtectedRoute />}>
            <Route path="/portal" element={<Navigate to="/portal/dashboard" replace />} />
            <Route path="/portal/dashboard" element={
              <ErrorBoundary FallbackComponent={GlobalErrorFallback}>
                <PortalLayout>
                  <PortalDashboard />
                </PortalLayout>
              </ErrorBoundary>
            } />
            <Route path="/portal/history" element={
              <ErrorBoundary FallbackComponent={GlobalErrorFallback}>
                <PortalLayout>
                  <PortalHistory />
                </PortalLayout>
              </ErrorBoundary>
            } />
            <Route path="/portal/requests" element={
              <ErrorBoundary FallbackComponent={GlobalErrorFallback}>
                <PortalLayout>
                  <PortalRequests />
                </PortalLayout>
              </ErrorBoundary>
            } />
          </Route>

          {/* Protected routes */}
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

            {/* Placeholder routes */}
            <Route path="/agents" element={<PlaceholderPage title="Agents" />} />
            {/* Active routes */}
            <Route path="/pii/review" element={<PIIDiscovery />} />
            <Route path="/dsr" element={<DSRList />} />
            <Route path="/dsr/:id" element={<DSRDetail />} />
            <Route path="/consent/widgets" element={<ConsentWidgets />} />
            <Route path="/consent/widgets/:id" element={<WidgetDetail />} />

            {/* Placeholder routes */}
            <Route path="/agents" element={<PlaceholderPage title="Agents" />} />
            <Route path="/pii/inventory" element={<PlaceholderPage title="PII Inventory" />} />
            <Route path="/lineage" element={<PlaceholderPage title="Data Lineage" />} />
            <Route path="/subjects" element={<PlaceholderPage title="Data Subjects" />} />
            <Route path="/consent" element={<PlaceholderPage title="Consent Records" />} />
            <Route path="/consent/analytics" element={<PlaceholderPage title="Consent Analytics" />} />
            <Route path="/grievances" element={<PlaceholderPage title="Grievances" />} />
            <Route path="/nominations" element={<PlaceholderPage title="Nominations" />} />
            {/* Governance Routes */}
            <Route path="/governance/purposes" element={<PurposeMapping />} />
            <Route path="/governance/policies" element={<PolicyManager />} />
            <Route path="/governance/violations" element={<Violations />} />

            {/* Placeholder routes */}
            <Route path="/department" element={<PlaceholderPage title="Department" />} />
            <Route path="/third-parties" element={<PlaceholderPage title="Third Parties" />} />
            <Route path="/retention" element={<PlaceholderPage title="Retention Policies" />} />
            <Route path="/ropa" element={<PlaceholderPage title="RoPA" />} />
            <Route path="/reports" element={<PlaceholderPage title="Reports" />} />
            <Route path="/audit-logs" element={<PlaceholderPage title="Audit Logs" />} />
            <Route path="/users" element={<PlaceholderPage title="User Management" />} />
            <Route path="/settings" element={<PlaceholderPage title="Settings" />} />
            <Route path="*" element={<PlaceholderPage title="404 — Page Not Found" />} />
          </Route>
        </Routes>
      </BrowserRouter>
    </QueryClientProvider>
  );
}

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

export default App;
