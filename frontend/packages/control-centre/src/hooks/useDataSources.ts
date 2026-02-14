import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { dataSourceService } from '../services/datasource';
import type { CreateDataSourceInput, UpdateDataSourceInput, M365ScopeConfig, GoogleScopeConfig } from '../types/datasource';
import type { ID } from '@datalens/shared';

const QUERY_KEY = ['dataSources'];

export function useDataSources() {
    return useQuery({
        queryKey: QUERY_KEY,
        queryFn: () => dataSourceService.list(),
        staleTime: 30 * 1000,
    });
}

export function useDataSource(id: ID) {
    return useQuery({
        queryKey: [...QUERY_KEY, id],
        queryFn: () => dataSourceService.getById(id),
        enabled: !!id,
    });
}

export function useCreateDataSource() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: (data: CreateDataSourceInput) => dataSourceService.create(data),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: QUERY_KEY });
        },
    });
}

export function useUpdateDataSource() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: ({ id, data }: { id: ID; data: UpdateDataSourceInput }) =>
            dataSourceService.update({ ...data, id }),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: QUERY_KEY });
        },
    });
}

export function useDeleteDataSource() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: (id: ID) => dataSourceService.delete(id),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: QUERY_KEY });
        },
    });
}

export function useScanDataSource() {
    return useMutation({
        mutationFn: (id: ID) => dataSourceService.scan(id),
    });
}

export function useScanStatus(id: ID, enabled: boolean = false) {
    return useQuery({
        queryKey: ['scanStatus', id],
        queryFn: () => dataSourceService.getScanStatus(id),
        enabled: !!id && enabled,
        refetchInterval: (query) => {
            const status = query.state.data?.status;
            if (status === 'COMPLETED' || status === 'FAILED') {
                return false;
            }
            return 3000;
        },
    });
}

export function useScanHistory(id: ID) {
    return useQuery({
        queryKey: ['scanHistory', id],
        queryFn: () => dataSourceService.getScanHistory(id),
        enabled: !!id,
    });
}

// --- M365 Scope Hooks ---

export function useM365Users(id: ID) {
    return useQuery({
        queryKey: ['m365Users', id],
        queryFn: () => dataSourceService.getM365Users(id),
        enabled: !!id,
    });
}

export function useSharePointSites(id: ID) {
    return useQuery({
        queryKey: ['sharePointSites', id],
        queryFn: () => dataSourceService.getSharePointSites(id),
        enabled: !!id,
    });
}

export function useUpdateScope() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: ({ id, config }: { id: ID; config: M365ScopeConfig | GoogleScopeConfig }) =>
            dataSourceService.updateScope(id, config),
        onSuccess: (_data, variables) => {
            queryClient.invalidateQueries({ queryKey: [...QUERY_KEY, variables.id] });
            queryClient.invalidateQueries({ queryKey: ['m365Users', variables.id] });
            queryClient.invalidateQueries({ queryKey: ['sharePointSites', variables.id] });
        },
    });
}
