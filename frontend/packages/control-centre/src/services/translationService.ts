import { api } from '@datalens/shared';
import type { ApiResponse } from '@datalens/shared';
import type { Translation, TranslationResult } from '../types/translation';

export const translationService = {
    async translateNotice(noticeId: string): Promise<TranslationResult[]> {
        const res = await api.post<ApiResponse<TranslationResult[]>>(`/consent/notices/${noticeId}/translate`);
        return res.data.data;
    },

    async getTranslations(noticeId: string): Promise<Translation[]> {
        const res = await api.get<ApiResponse<Translation[]>>(`/consent/notices/${noticeId}/translations`);
        return res.data.data;
    },

    async overrideTranslation(noticeId: string, language: string, content: Record<string, string>): Promise<void> {
        await api.put(`/consent/notices/${noticeId}/translations/${language}`, { content });
    }
};
