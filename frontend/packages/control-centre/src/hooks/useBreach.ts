import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { breachService } from '../services/breach';
import type { BreachFilter, CreateIncidentInput, UpdateIncidentInput } from '../types/breach';
import type { ID } from '@datalens/shared';

export function useBreachList(params?: { page?: number; page_size?: number } & BreachFilter) {
    return useQuery({
        queryKey: ['breach-incidents', params],
        queryFn: () => breachService.list(params),
    });
}

export function useBreach(id: ID) {
    return useQuery({
        queryKey: ['breach-incidents', id],
        queryFn: () => breachService.getById(id),
        enabled: !!id,
    });
}

export function useCreateIncident() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: (data: CreateIncidentInput) => breachService.create(data),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['breach-incidents'] });
        },
    });
}

export function useUpdateIncident() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: ({ id, data }: { id: ID; data: UpdateIncidentInput }) =>
            breachService.update(id, data),
        onSuccess: (data) => {
            queryClient.invalidateQueries({ queryKey: ['breach-incidents'] });
            queryClient.invalidateQueries({ queryKey: ['breach-incidents', data.id] });
        },
    });
}

export function useCertInReport() {
    return useMutation({
        mutationFn: (id: ID) => breachService.generateCertInReport(id),
    });
}
