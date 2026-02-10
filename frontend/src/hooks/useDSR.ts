import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { dsrService } from '../services/dsr';
import type { ID } from '../types/common';
import type { CreateDSRInput } from '../types/dsr';

const DSR_KEYS = {
    all: ['dsrs'] as const,
    list: (params?: Record<string, unknown>) => [...DSR_KEYS.all, 'list', params] as const,
    detail: (id: ID) => [...DSR_KEYS.all, 'detail', id] as const,
};

export function useDSRs(params?: { page?: number; page_size?: number; status?: string }) {
    return useQuery({
        queryKey: DSR_KEYS.list(params as Record<string, unknown>),
        queryFn: () => dsrService.list(params),
    });
}

export function useDSR(id: ID) {
    return useQuery({
        queryKey: DSR_KEYS.detail(id),
        queryFn: () => dsrService.getById(id),
        enabled: !!id,
    });
}

export function useCreateDSR() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: (data: CreateDSRInput) => dsrService.create(data),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: DSR_KEYS.all });
        },
    });
}

export function useApproveDSR() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: (id: ID) => dsrService.approve(id),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: DSR_KEYS.all });
        },
    });
}

export function useRejectDSR() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: ({ id, reason }: { id: ID; reason: string }) => dsrService.reject(id, reason),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: DSR_KEYS.all });
        },
    });
}

export function useExecuteDSR() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: (id: ID) => dsrService.execute(id),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: DSR_KEYS.all });
        },
    });
}
