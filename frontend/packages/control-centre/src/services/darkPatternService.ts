import { api } from '@datalens/shared';
import type { ApiResponse } from '@datalens/shared';
import type { DarkPatternAnalysisResult, AnalyzeContentRequest } from '../types/darkPattern';

export const darkPatternService = {
    async analyzeContent(data: AnalyzeContentRequest): Promise<DarkPatternAnalysisResult> {
        const res = await api.post<ApiResponse<DarkPatternAnalysisResult>>('/analytics/dark-pattern/analyze', data);
        return res.data.data;
    },
};
