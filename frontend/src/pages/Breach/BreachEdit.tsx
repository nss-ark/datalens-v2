import { useParams, useNavigate } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'react-toastify';
import { BreachForm } from '../../components/Breach/BreachForm';
import { breachService } from '../../services/breach';
import type { UpdateIncidentInput } from '../../types/breach';

const BreachEdit = () => {
    const { id } = useParams<{ id: string }>();
    const navigate = useNavigate();
    const queryClient = useQueryClient();

    const { data, isLoading, isError } = useQuery({
        queryKey: ['breach', id],
        queryFn: () => breachService.getById(id!),
        enabled: !!id
    });

    const updateMutation = useMutation({
        mutationFn: (updateData: UpdateIncidentInput) => breachService.update(id!, updateData),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['breach', id] });
            toast.success('Incident updated successfully');
            navigate(`/breach/${id}`);
        },
        onError: (error) => {
            console.error('Failed to update incident:', error);
            toast.error('Failed to update incident');
        }
    });

    if (isLoading) return <div className="p-8 text-center">Loading incident details...</div>;
    if (isError || !data) return <div className="p-8 text-center text-red-600">Failed to load incident</div>;

    return (
        <div className="p-6 max-w-[1200px] mx-auto">
            <div className="mb-6">
                <h1 className="text-2xl font-bold text-gray-900">Edit Incident</h1>
                <p className="text-gray-500 mt-1">
                    Update incident details.
                </p>
            </div>

            <BreachForm
                initialData={data.incident}
                isEdit={true}
                onSubmit={(formData) => updateMutation.mutate(formData)}
                isLoading={updateMutation.isPending}
            />
        </div>
    );
};

export default BreachEdit;
