import { create } from 'zustand';
import { persist, createJSONStorage } from 'zustand/middleware';
import type { PortalProfile } from '@/types/portal';

interface PortalAuthState {
    token: string | null;
    profile: PortalProfile | null;
    isAuthenticated: boolean;
    setAuth: (token: string, profile: PortalProfile) => void;
    logout: () => void;
}

export const usePortalAuthStore = create<PortalAuthState>()(
    persist(
        (set) => ({
            token: null,
            profile: null,
            isAuthenticated: false,
            setAuth: (token, profile) => set({ token, profile, isAuthenticated: true }),
            logout: () => set({ token: null, profile: null, isAuthenticated: false }),
        }),
        {
            name: 'datalens-portal-auth', // unique name
            storage: createJSONStorage(() => sessionStorage), // Use sessionStorage for security
        }
    )
);
