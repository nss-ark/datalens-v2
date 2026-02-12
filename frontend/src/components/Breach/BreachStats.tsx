import { StatCard } from '../Dashboard/StatCard';
import { ShieldAlert, AlertCircle, FileText } from 'lucide-react';
import type { BreachIncident } from '../../types/breach';

interface BreachStatsProps {
    incidents: BreachIncident[];
}

export const BreachStats = ({ incidents }: BreachStatsProps) => {
    const activeIncidents = incidents.filter(i =>
        ['OPEN', 'INVESTIGATING', 'CONTAINED'].includes(i.status)
    ).length;

    const criticalBreaches = incidents.filter(i =>
        i.severity === 'CRITICAL' && i.status !== 'CLOSED'
    ).length;

    // Find next SLA deadline (simplification: just count overdue for now or nearest)
    // In a real app we'd parse all deadlines.
    const openHighSeverity = incidents.filter(i =>
        (i.severity === 'HIGH' || i.severity === 'CRITICAL') && i.status !== 'CLOSED'
    ).length;

    const pendingReports = openHighSeverity;

    return (
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
            <StatCard
                title="Active Incidents"
                value={activeIncidents}
                icon={AlertCircle}
                color="warning"
                trend={{
                    value: 0,
                    label: 'from last week',
                    direction: 'neutral'
                }}
            />
            <StatCard
                title="Critical Breaches"
                value={criticalBreaches}
                icon={ShieldAlert}
                color="danger"
                trend={{
                    value: 0,
                    label: 'requires attention',
                    direction: criticalBreaches > 0 ? 'down' : 'neutral'
                }}
            />
            <StatCard
                title="Pending Reports"
                value={pendingReports}
                icon={FileText}
                color="info"
                trend={{
                    value: 0,
                    label: 'SLA ticking',
                    direction: 'neutral'
                }}
            />
        </div>
    );
};
