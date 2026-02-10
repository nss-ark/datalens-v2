import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { dataSourceService } from '../services/datasources';
import type { CreateDataSourceInput, UpdateDataSourceInput } from '../types/datasource';
import type { ID } from '../types/common';

const QUERY_KEY = ['dataSources'];

export function useDataSources() {
    return useQuery({
        queryKey: QUERY_KEY,
        queryFn: () => dataSourceService.list(),
        staleTime: 30 * 1000, // 30 seconds
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
            dataSourceService.update(id, data),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: QUERY_KEY });
        },
    });
}

export function useDeleteDataSource() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: (id: ID) => dataSourceService.remove(id),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: QUERY_KEY });
        },
    });
}

export function useScanDataSource() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: (id: ID) => dataSourceService.scan(id),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: QUERY_KEY });
        },
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
            return 3000; // Poll every 3 seconds
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
