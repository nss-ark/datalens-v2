import { useState } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { translationService } from '../../services/translationService';
import { Button } from '../common/Button';
import { toast } from 'react-toastify';
import { SUPPORTED_LANGUAGES, type Translation } from '../../types/translation';

interface Props {
    noticeId: string;
    translation: Translation | null;
    languageCode: string; // The language we are editing
    baseContent: Record<string, string>; // Original English content for reference
    onClose: () => void;
}

export function TranslationOverrideModal({ noticeId, translation, languageCode, baseContent, onClose }: Props) {
    const queryClient = useQueryClient();
    const [content, setContent] = useState<Record<string, string>>(() => {
        if (translation?.content) {
            return translation.content;
        }
        // Initialize with empty strings based on reference baseContent keys
        const initial: Record<string, string> = {};
        Object.keys(baseContent).forEach(key => initial[key] = '');
        return initial;
    });

    const languageName = SUPPORTED_LANGUAGES.find(l => l.code === languageCode)?.name || languageCode;

    const mutation = useMutation({
        mutationFn: (data: Record<string, string>) => translationService.overrideTranslation(noticeId, languageCode, data),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['notice-translations', noticeId] });
            toast.success('Translation updated');
            onClose();
        },
        onError: () => toast.error('Failed to update translation')
    });

    const handleSave = () => {
        mutation.mutate(content);
    };

    return (
        <div className="space-y-4">
            <p className="text-sm text-gray-500">
                Manually edit the <strong>{languageName}</strong> translation.
                Use the original English text as a reference.
            </p>

            <div className="max-h-96 overflow-y-auto space-y-4 border p-4 rounded bg-gray-50">
                {Object.entries(baseContent).map(([key, originalText]) => (
                    <div key={key} className="grid grid-cols-1 md:grid-cols-2 gap-4">
                        <div>
                            <label className="block text-xs font-medium text-gray-500 mb-1">Original ({key})</label>
                            <div className="p-2 bg-white border border-gray-200 rounded text-sm text-gray-800">
                                {originalText}
                            </div>
                        </div>
                        <div>
                            <label className="block text-xs font-medium text-gray-700 mb-1">Translation</label>
                            <textarea
                                className="w-full p-2 border border-gray-300 rounded text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                                rows={3}
                                value={content[key] || ''}
                                onChange={(e) => setContent(prev => ({ ...prev, [key]: e.target.value }))}
                            />
                        </div>
                    </div>
                ))}
            </div>

            <div className="flex justify-end space-x-3 pt-2">
                <Button variant="secondary" onClick={onClose}>Cancel</Button>
                <Button onClick={handleSave} isLoading={mutation.isPending}>Save Translation</Button>
            </div>
        </div>
    );
}
