import { useNavigate } from 'react-router-dom';
import { BreachForm } from '../../components/Breach/BreachForm';
import { breachService } from '../../services/breach';
import { useMutation } from '@tanstack/react-query';
import { toast } from 'react-toastify';
import type { CreateIncidentInput } from '../../types/breach';

const BreachCreate = () => {
    const navigate = useNavigate();

    const createMutation = useMutation({
        mutationFn: (data: CreateIncidentInput) => breachService.create(data),
        onSuccess: (data) => {
            toast.success('Incident reported successfully');
            navigate(`/breach/${data.id}`);
        },
        onError: (error) => {
            console.error('Failed to create incident:', error);
            // Error is handled globally or we can show specific message here
            toast.error('Failed to report incident');
        }
    });

    const handleSubmit = (data: CreateIncidentInput) => {
        createMutation.mutate(data);
    };

    return (
        <div className="p-6 max-w-[1200px] mx-auto">
            <div className="mb-6">
                <h1 className="text-2xl font-bold text-gray-900">Report Security Incident</h1>
                <p className="text-gray-500 mt-1">
                    Log a new security breach or incident. Please provide as much detail as possible.
                </p>
            </div>

            <BreachForm
                onSubmit={handleSubmit}
                isLoading={createMutation.isPending}
            />
        </div>
    );
};

export default BreachCreate;
