// UI Kit (shadcn/ui)
export * from './ui/badge';
export { Button as ShadcnButton, buttonVariants } from './ui/button';
export type { ButtonProps as ShadcnButtonProps } from './ui/button';
export * from './ui/card';
export * from './ui/dialog';
export * from './ui/form';
export * from './ui/input';
export * from './ui/label';
export * from './ui/select';
export * from './ui/table';
export * from './ui/textarea';

// Common Components
export { ErrorBoundary } from './components/ErrorBoundary';
export { GlobalErrorFallback, SectionErrorFallback } from './components/ErrorFallbacks';
export { ToastContainer } from './components/Toast';
export { StatusBadge } from './components/StatusBadge';
export { Button, type ButtonProps } from './components/Button';
export { Modal } from './components/Modal';
export { DataTable } from './components/DataTable/DataTable';
export type { Column, SortState, SortDirection } from './components/DataTable/DataTable';
export { Pagination } from './components/DataTable/Pagination';

// Types
export type * from './types/common';
export type * from './types/auth';

// Services
export { api } from './services/api';
export { authService } from './services/auth';

// Stores
export { useAuthStore } from './stores/authStore';
export { useToastStore, toast } from './stores/toastStore';
export type { ToastVariant } from './stores/toastStore';
export { useUIStore } from './stores/uiStore';

// Hooks
export { useLogin, useRegister, useRefreshToken, useCurrentUser, useLogout } from './hooks/useAuth';
export { useMediaQuery } from './hooks/useMediaQuery';

// Utils
export { cn } from './lib/utils';
