import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { consentService } from '../services/consent';
import type { CreateWidgetInput, UpdateWidgetInput } from '../types/consent';
import type { ID } from '../types/common';

export function useWidgets(params?: { page?: number; page_size?: number }) {
    return useQuery({
        queryKey: ['consent-widgets', params],
        queryFn: () => consentService.listWidgets(params),
    });
}

export function useWidget(id: ID) {
    return useQuery({
        queryKey: ['consent-widgets', id],
        queryFn: () => consentService.getWidget(id),
        enabled: !!id,
    });
}

export function useCreateWidget() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: (data: CreateWidgetInput) => consentService.createWidget(data),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['consent-widgets'] });
        },
    });
}

export function useUpdateWidget() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: ({ id, data }: { id: ID; data: UpdateWidgetInput }) => consentService.updateWidget(id, data),
        onSuccess: (data) => {
            queryClient.invalidateQueries({ queryKey: ['consent-widgets'] });
            queryClient.invalidateQueries({ queryKey: ['consent-widgets', data.id] });
        },
    });
}

export function useDeleteWidget() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: (id: ID) => consentService.deleteWidget(id),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['consent-widgets'] });
        },
    });
}

export function useActivateWidget() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: (id: ID) => consentService.activateWidget(id),
        onSuccess: (data) => {
            queryClient.invalidateQueries({ queryKey: ['consent-widgets'] });
            queryClient.invalidateQueries({ queryKey: ['consent-widgets', data.id] });
        },
    });
}

export function usePauseWidget() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: (id: ID) => consentService.pauseWidget(id),
        onSuccess: (data) => {
            queryClient.invalidateQueries({ queryKey: ['consent-widgets'] });
            queryClient.invalidateQueries({ queryKey: ['consent-widgets', data.id] });
        },
    });
}
