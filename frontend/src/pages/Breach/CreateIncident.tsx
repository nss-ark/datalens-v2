import { useNavigate } from 'react-router-dom';
import { ArrowLeft } from 'lucide-react';
import { Button } from '../../components/common/Button';
import { BreachForm } from '../../components/Breach/BreachForm';
import { useCreateIncident } from '../../hooks/useBreach';
import { toast } from 'react-toastify';
import type { CreateIncidentInput } from '../../types/breach';

const CreateIncident = () => {
    const navigate = useNavigate();
    const createMutation = useCreateIncident();

    const handleCreate = (incidentData: CreateIncidentInput) => {
        createMutation.mutate(incidentData, {
            onSuccess: () => {
                toast.success('Incident reported successfully');
                navigate('/breach');
            },
            onError: () => toast.error('Failed to report incident')
        });
    };

    return (
        <div className="p-6 max-w-4xl mx-auto">
            <Button variant="ghost" onClick={() => navigate('/breach')} className="mb-4 pl-0">
                <ArrowLeft size={16} className="mr-2" /> Back to Dashboard
            </Button>

            <h1 className="text-2xl font-bold mb-2">Report New Security Incident</h1>
            <p className="text-gray-500 mb-8">
                Provide details about the potential breach. High severity incidents will trigger mandatory reporting workflows.
            </p>

            <BreachForm
                onSubmit={handleCreate}
                isLoading={createMutation.isPending}
            />
        </div>
    );
};

export default CreateIncident;
