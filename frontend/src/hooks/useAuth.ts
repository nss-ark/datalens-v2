import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { authService } from '../services/auth';
import type { LoginRequest, RegisterRequest } from '../services/auth';
import { useAuthStore } from '../stores/authStore';

export function useLogin() {
    const login = useAuthStore((s) => s.login);
    const navigate = useNavigate();

    return useMutation({
        mutationFn: (data: LoginRequest) => authService.login(data),
        onSuccess: async (tokenPair) => {
            // Store token first so getMe call has auth
            useAuthStore.setState({ token: tokenPair.access_token });

            // Fetch full user profile
            try {
                const user = await authService.getMe();
                login(user, tokenPair.access_token, user.tenant_id);
                navigate('/dashboard');
            } catch {
                // If getMe fails, still store what we have
                login(
                    { id: '', name: '', email: '', status: 'ACTIVE', role_ids: [], tenant_id: '', mfa_enabled: false },
                    tokenPair.access_token,
                    ''
                );
                navigate('/dashboard');
            }
        },
    });
}

export function useRegister() {
    return useMutation({
        mutationFn: (data: RegisterRequest) => authService.register(data),
    });
}

export function useRefreshToken() {
    return useMutation({
        mutationFn: (refreshToken: string) => authService.refreshToken(refreshToken),
        onSuccess: (tokenPair) => {
            useAuthStore.setState({ token: tokenPair.access_token });
        },
    });
}

export function useCurrentUser() {
    const isAuthenticated = useAuthStore((s) => s.isAuthenticated);

    return useQuery({
        queryKey: ['currentUser'],
        queryFn: () => authService.getMe(),
        enabled: isAuthenticated,
        staleTime: 5 * 60 * 1000, // 5 minutes
        retry: false,
    });
}

export function useLogout() {
    const logout = useAuthStore((s) => s.logout);
    const navigate = useNavigate();
    const queryClient = useQueryClient();

    return () => {
        logout();
        queryClient.clear();
        navigate('/login');
    };
}
