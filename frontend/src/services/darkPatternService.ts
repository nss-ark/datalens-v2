import { api } from './api';
import type { ApiResponse } from '../types/common';
import type { DarkPatternAnalysisResult, AnalyzeContentRequest } from '../types/darkPattern';

export const darkPatternService = {
    async analyzeContent(data: AnalyzeContentRequest): Promise<DarkPatternAnalysisResult> {
        const res = await api.post<ApiResponse<DarkPatternAnalysisResult>>('/analytics/dark-pattern/analyze', data);
        return res.data.data;
    },
};
