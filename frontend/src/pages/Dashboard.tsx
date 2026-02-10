import { Users, Database, ShieldCheck, AlertTriangle } from 'lucide-react';
import { Button } from '../components/common/Button';

// Temporary StatCard component until fully implemented
const StatCard = ({ title, value, icon: Icon, color }: any) => (
    <div style={{
        backgroundColor: 'white',
        padding: '1.5rem',
        borderRadius: 'var(--radius-lg)',
        border: '1px solid var(--border-color)',
        display: 'flex',
        alignItems: 'center',
        gap: '1rem',
        boxShadow: 'var(--shadow-sm)'
    }}>
        <div style={{
            padding: '0.75rem',
            borderRadius: 'var(--radius-md)',
            backgroundColor: `var(--${color}-50)`,
            color: `var(--${color}-600)`
        }}>
            <Icon size={24} />
        </div>
        <div>
            <div style={{ fontSize: '0.875rem', color: 'var(--text-secondary)', fontWeight: 500 }}>{title}</div>
            <div style={{ fontSize: '1.5rem', fontWeight: 600, color: 'var(--text-primary)' }}>{value}</div>
        </div>
    </div>
);

const Dashboard = () => {
    return (
        <div>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '2rem' }}>
                <div>
                    <h1 style={{ fontSize: '1.875rem', fontWeight: 600, color: 'var(--text-primary)', marginBottom: '0.5rem' }}>
                        Good afternoon, User
                    </h1>
                    <p style={{ color: 'var(--text-secondary)' }}>
                        Here's what's happening with your data compliance today.
                    </p>
                </div>
                <div style={{ display: 'flex', gap: '0.75rem' }}>
                    <Button variant="outline">Download Report</Button>
                    <Button>Add New Data Source</Button>
                </div>
            </div>

            <div style={{
                display: 'grid',
                gridTemplateColumns: 'repeat(auto-fit, minmax(240px, 1fr))',
                gap: '1.5rem',
                marginBottom: '2rem'
            }}>
                <StatCard title="Total Data Sources" value="12" icon={Database} color="primary" />
                <StatCard title="Total PII Found" value="1,248" icon={ShieldCheck} color="success" />
                <StatCard title="Pending Review" value="45" icon={AlertTriangle} color="warning" />
                <StatCard title="Active Users" value="8" icon={Users} color="info" />
            </div>

            <div style={{
                backgroundColor: 'white',
                borderRadius: 'var(--radius-lg)',
                border: '1px solid var(--border-color)',
                padding: '1.5rem',
                minHeight: '300px'
            }}>
                <h2 style={{ fontSize: '1.125rem', fontWeight: 600, marginBottom: '1rem' }}>Compliance Trends</h2>
                <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', height: '200px', color: 'var(--text-tertiary)' }}>
                    Chart placeholder
                </div>
            </div>
        </div>
    );
};

export default Dashboard;
