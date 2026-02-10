import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { discoveryService, type ClassificationFilters } from '../services/discovery';
import type { SubmitFeedbackInput, DetectionMethod } from '../types/discovery';

const CLS_KEY = ['classifications'];

export function useClassifications(filters?: ClassificationFilters) {
    return useQuery({
        queryKey: [...CLS_KEY, filters],
        queryFn: () => discoveryService.listClassifications(filters),
        staleTime: 30_000,
    });
}

export function useSubmitFeedback() {
    const qc = useQueryClient();
    return useMutation({
        mutationFn: (input: SubmitFeedbackInput) => discoveryService.submitFeedback(input),
        onSuccess: () => {
            qc.invalidateQueries({ queryKey: CLS_KEY });
        },
    });
}

export function useAccuracyStats(method: DetectionMethod) {
    return useQuery({
        queryKey: ['accuracyStats', method],
        queryFn: () => discoveryService.getAccuracyStats(method),
        staleTime: 60_000,
    });
}
