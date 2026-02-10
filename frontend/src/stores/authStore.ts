import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import type { User } from '../types/auth';

interface AuthState {
    user: User | null;
    token: string | null;
    tenantId: string | null;
    isAuthenticated: boolean;
    login: (user: User, token: string, tenantId: string) => void;
    logout: () => void;
    updateUser: (user: Partial<User>) => void;
}

export const useAuthStore = create<AuthState>()(
    persist(
        (set) => ({
            user: null,
            token: null,
            tenantId: null,
            isAuthenticated: false,
            login: (user, token, tenantId) =>
                set({
                    user,
                    token,
                    tenantId,
                    isAuthenticated: true,
                }),
            logout: () =>
                set({
                    user: null,
                    token: null,
                    tenantId: null,
                    isAuthenticated: false,
                }),
            updateUser: (updates) =>
                set((state) => ({
                    user: state.user ? { ...state.user, ...updates } : null,
                })),
        }),
        {
            name: 'auth-storage', // unique name
        }
    )
);
