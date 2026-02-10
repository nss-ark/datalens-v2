import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { AppLayout } from './components/Layout/AppLayout';
import { ProtectedRoute } from './components/common/ProtectedRoute';
import { ToastContainer } from './components/common/Toast';
import Login from './pages/Login';
import Register from './pages/Register';
import Dashboard from './pages/Dashboard';
import DataSources from './pages/DataSources';
import PIIDiscovery from './pages/PIIDiscovery';
import DSRList from './pages/DSRList';
import DSRDetail from './pages/DSRDetail';

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

          {/* Protected routes */}
          <Route element={
            <ProtectedRoute>
              <AppLayout />
            </ProtectedRoute>
          }>
            <Route path="/" element={<Navigate to="/dashboard" replace />} />
            <Route path="/dashboard" element={<Dashboard />} />
            <Route path="/datasources" element={<DataSources />} />

            {/* Placeholder routes */}
            <Route path="/agents" element={<PlaceholderPage title="Agents" />} />
            {/* Active routes */}
            <Route path="/pii/review" element={<PIIDiscovery />} />

            {/* Placeholder routes */}
            <Route path="/agents" element={<PlaceholderPage title="Agents" />} />
            <Route path="/pii/inventory" element={<PlaceholderPage title="PII Inventory" />} />
            <Route path="/lineage" element={<PlaceholderPage title="Data Lineage" />} />
            <Route path="/subjects" element={<PlaceholderPage title="Data Subjects" />} />
            <Route path="/dsr" element={<DSRList />} />
            <Route path="/dsr/:id" element={<DSRDetail />} />
            <Route path="/consent" element={<PlaceholderPage title="Consent Records" />} />
            <Route path="/consent/analytics" element={<PlaceholderPage title="Consent Analytics" />} />
            <Route path="/grievances" element={<PlaceholderPage title="Grievances" />} />
            <Route path="/nominations" element={<PlaceholderPage title="Nominations" />} />
            <Route path="/purposes" element={<PlaceholderPage title="Purposes" />} />
            <Route path="/departments" element={<PlaceholderPage title="Departments" />} />
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
