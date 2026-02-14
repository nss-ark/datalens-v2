import { useForm } from 'react-hook-form';
import { z } from 'zod';
import { Button } from '@datalens/shared';
import { consentService } from '../../services/consent';
import type { ConsentNotice, UpdateNoticeInput } from '../../types/consent';
import { useQueryClient, useMutation } from '@tanstack/react-query';
import { toast } from 'react-toastify';

export const noticeSchema = z.object({
    title: z.string().min(1, 'Title is required'),
    content: z.string().min(1, 'Content is required'),
    regulation: z.string().min(1, 'Regulation is required'),
    purposes: z.array(z.string()).min(1, 'At least one purpose is required'),
});

type NoticeFormData = z.infer<typeof noticeSchema>;

interface NoticeFormProps {
    initialData?: ConsentNotice;
    onSuccess: () => void;
    onCancel: () => void;
}

const AVAILABLE_PURPOSES = [
    { id: 'essential', name: 'Essential' },
    { id: 'functional', name: 'Functional' },
    { id: 'analytics', name: 'Analytics' },
    { id: 'marketing', name: 'Marketing' },
    { id: 'advertising', name: 'Advertising' },
    { id: 'security', name: 'Security' },
];

export function NoticeForm({ initialData, onSuccess, onCancel }: NoticeFormProps) {
    const queryClient = useQueryClient();

    const {
        register,
        handleSubmit,
        setValue,
        watch,
        formState: { errors, isSubmitting },
    } = useForm<NoticeFormData>({
        defaultValues: {
            title: initialData?.title || '',
            content: initialData?.content || '',
            regulation: initialData?.regulation || '',
            purposes: initialData?.purposes || [],
        },
    });

    const createMutation = useMutation({
        mutationFn: consentService.createNotice,
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['consent-notices'] });
            toast.success('Notice created successfully');
            onSuccess();
        },
        onError: () => toast.error('Failed to create notice'),
    });

    const updateMutation = useMutation({
        mutationFn: (data: UpdateNoticeInput) => consentService.updateNotice(initialData!.id, data),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['consent-notices'] });
            toast.success('Notice updated successfully');
            onSuccess();
        },
        onError: () => toast.error('Failed to update notice'),
    });

    const onSubmit = (data: NoticeFormData) => {
        if (initialData) {
            updateMutation.mutate({
                id: initialData.id,
                ...data,
            });
        } else {
            createMutation.mutate(data);
        }
    };

    // Purpose selection helper
    // eslint-disable-next-line
    const selectedPurposes: string[] = watch('purposes') || [];

    const togglePurpose = (id: string) => {
        const current = selectedPurposes;
        if (current.includes(id)) {
            setValue('purposes', current.filter((p) => p !== id));
        } else {
            setValue('purposes', [...current, id]);
        }
    };

    return (
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
            <div>
                <label className="block text-sm font-medium text-gray-700">Title</label>
                <input
                    {...register('title', { required: 'Title is required' })}
                    className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm p-2 border"
                    placeholder="e.g., General Privacy Notice"
                />
                {errors.title && <p className="text-red-500 text-xs mt-1">{errors.title.message}</p>}
            </div>

            <div>
                <label className="block text-sm font-medium text-gray-700">Regulation</label>
                <input
                    {...register('regulation', { required: 'Regulation is required' })}
                    className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm p-2 border"
                    placeholder="e.g., GDPR, CCPA, DPDP"
                />
                {errors.regulation && <p className="text-red-500 text-xs mt-1">{errors.regulation.message}</p>}
            </div>

            <div>
                <label className="block text-sm font-medium text-gray-700">Content</label>
                <textarea
                    {...register('content', { required: 'Content is required' })}
                    rows={10}
                    className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm p-2 border font-mono"
                    placeholder="Enter policy content..."
                />
                {errors.content && <p className="text-red-500 text-xs mt-1">{errors.content.message}</p>}
            </div>

            <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">Purposes</label>
                <div className="grid grid-cols-2 gap-2 max-h-40 overflow-y-auto border p-2 rounded bg-gray-50">
                    {AVAILABLE_PURPOSES.map((p) => (
                        <label key={p.id} className="flex items-center space-x-2">
                            <input
                                type="checkbox"
                                value={p.id}
                                checked={selectedPurposes?.includes(p.id)}
                                onChange={() => togglePurpose(p.id)}
                                className="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                            />
                            <span className="text-sm text-gray-700">{p.name}</span>
                        </label>
                    ))}
                </div>
                {errors.purposes && <p className="text-red-500 text-xs mt-1">{errors.purposes.message}</p>}
            </div>

            <div className="flex justify-end space-x-3 pt-4">
                <Button variant="secondary" onClick={onCancel} type="button">
                    Cancel
                </Button>
                <Button type="submit" isLoading={isSubmitting}>
                    {initialData ? 'Update Notice' : 'Create Notice'}
                </Button>
            </div>
        </form>
    );
}
