import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { translationService } from '../../services/translationService';
import { Button } from '../common/Button';
import { StatusBadge } from '../common/StatusBadge';
import { Modal } from '../common/Modal';
import { TranslationOverrideModal } from './TranslationOverrideModal';
import { SUPPORTED_LANGUAGES } from '../../types/translation';
import { RefreshCw, Edit3 } from 'lucide-react';
import { toast } from 'react-toastify';
import type { ConsentNotice } from '../../types/consent';

interface Props {
    notice: ConsentNotice;
    onClose: () => void;
}

const baseContent = { "body": "Privacy Notice Content..." };

export function TranslationManagementModal({ notice, onClose }: Props) {
    const queryClient = useQueryClient();
    const [overrideLang, setOverrideLang] = useState<string | null>(null);

    const { data: translations = [] } = useQuery({
        queryKey: ['notice-translations', notice.id],
        queryFn: () => translationService.getTranslations(notice.id),
        enabled: !!notice.id
    });

    const translateMutation = useMutation({
        mutationFn: () => translationService.translateNotice(notice.id),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['notice-translations', notice.id] });
            toast.success('Translation job started');
        },
        onError: () => toast.error('Failed to start translation')
    });

    // Helper to get status of a language
    const getStatus = (code: string) => {
        const t = translations.find(t => t.language === code);
        return t ? t.status : 'PENDING'; // Default to pending/missing if not found
    };

    const getTranslation = (code: string) => translations.find(t => t.language === code) || null;

    return (
        <div className="space-y-6">
            <div className="flex justify-between items-center bg-blue-50 p-4 rounded-lg">
                <div>
                    <h3 className="font-medium text-blue-900">Auto-Translation</h3>
                    <p className="text-sm text-blue-700">
                        Translate this notice into all 22 scheduled languages using AI.
                    </p>
                </div>
                <Button
                    variant="primary"
                    icon={<RefreshCw size={16} />}
                    onClick={() => translateMutation.mutate()}
                    isLoading={translateMutation.isPending}
                >
                    Translate All
                </Button>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3 max-h-[60vh] overflow-y-auto">
                {SUPPORTED_LANGUAGES.map(lang => {
                    const status = getStatus(lang.code);

                    return (
                        <div key={lang.code} className="border rounded-md p-3 flex justify-between items-center hover:bg-gray-50">
                            <div className="flex items-center space-x-2">
                                <div className="bg-gray-200 w-8 h-8 rounded-full flex items-center justify-center text-xs font-bold text-gray-700">
                                    {lang.code.toUpperCase()}
                                </div>
                                <div>
                                    <div className="text-sm font-medium">{lang.name}</div>
                                    <StatusBadge label={status} size="sm" />
                                </div>
                            </div>
                            <button
                                onClick={() => setOverrideLang(lang.code)}
                                className="text-gray-400 hover:text-blue-600 p-1"
                                title="Edit Translation"
                            >
                                <Edit3 size={16} />
                            </button>
                        </div>
                    );
                })}
            </div>

            <div className="flex justify-end pt-4 border-t">
                <Button variant="secondary" onClick={onClose}>Close</Button>
            </div>

            {/* Override Modal */}
            <Modal
                open={!!overrideLang}
                onClose={() => setOverrideLang(null)}
                title={`Edit Translation: ${SUPPORTED_LANGUAGES.find(l => l.code === overrideLang)?.name}`}
                size="lg"
            >
                {overrideLang && (
                    <TranslationOverrideModal
                        key={overrideLang}
                        noticeId={notice.id}
                        languageCode={overrideLang}
                        translation={getTranslation(overrideLang)}
                        baseContent={baseContent} // Passing dummy content for now as we don't have notice content in list item
                        onClose={() => setOverrideLang(null)}
                    />
                )}
            </Modal>
        </div>
    );
}
