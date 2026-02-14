import { StatusBadge } from '@datalens/shared';
import type { IncidentStatus } from '../../types/breach';

interface BreachStatusBadgeProps {
    status: IncidentStatus;
}

export const BreachStatusBadge = ({ status }: BreachStatusBadgeProps) => {
    let variant: 'success' | 'warning' | 'danger' | 'info' | 'neutral' = 'neutral';

    switch (status) {
        case 'OPEN':
            variant = 'danger';
            break;
        case 'INVESTIGATING':
            variant = 'warning';
            break;
        case 'CONTAINED':
            variant = 'info';
            break;
        case 'RESOLVED':
        case 'REPORTED':
        case 'CLOSED':
            variant = 'neutral';
            break;
        default:
            variant = 'neutral';
            break;
    }

    return <StatusBadge label={status} variant={variant} />;
};
