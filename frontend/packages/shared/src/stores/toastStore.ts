import { create } from 'zustand';

export type ToastVariant = 'success' | 'error' | 'warning' | 'info';

export interface Toast {
    id: string;
    title: string;
    message?: string;
    variant: ToastVariant;
}

interface ToastStore {
    toasts: Toast[];
    addToast: (toast: Omit<Toast, 'id'>) => void;
    removeToast: (id: string) => void;
}

let toastCounter = 0;

export const useToastStore = create<ToastStore>((set) => ({
    toasts: [],
    addToast: (toast) => {
        // Deduplicate: skip if a toast with the same title already exists
        const existing = useToastStore.getState().toasts;
        if (existing.some((t) => t.title === toast.title)) return;

        const id = `toast-${++toastCounter}`;
        set((state) => ({ toasts: [...state.toasts, { ...toast, id }] }));
        // Auto-dismiss after 5 seconds
        setTimeout(() => {
            set((state) => ({ toasts: state.toasts.filter((t) => t.id !== id) }));
        }, 5000);
    },
    removeToast: (id) => {
        set((state) => ({ toasts: state.toasts.filter((t) => t.id !== id) }));
    },
}));

// Convenience helpers
export const toast = {
    success: (title: string, message?: string) =>
        useToastStore.getState().addToast({ title, message, variant: 'success' }),
    error: (title: string, message?: string) =>
        useToastStore.getState().addToast({ title, message, variant: 'error' }),
    warning: (title: string, message?: string) =>
        useToastStore.getState().addToast({ title, message, variant: 'warning' }),
    info: (title: string, message?: string) =>
        useToastStore.getState().addToast({ title, message, variant: 'info' }),
};
