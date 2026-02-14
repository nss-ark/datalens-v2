import { useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import {
    ArrowLeft, CheckCircle, XCircle, Play, Clock, AlertTriangle,
} from 'lucide-react';
import { Button } from '@datalens/shared';
import { StatusBadge } from '@datalens/shared';
import { DataTable } from '@datalens/shared';
import { Modal } from '@datalens/shared';
import { useDSR, useApproveDSR, useRejectDSR, useExecuteDSR } from '../hooks/useDSR';
import { useToastStore } from '@datalens/shared';
import type { DSRTask } from '../types/dsr';
import type { Column } from '@datalens/shared';

// SLA helpers
function getDaysRemaining(deadline: string): number {
    const now = new Date();
    const sla = new Date(deadline);
    return Math.ceil((sla.getTime() - now.getTime()) / (1000 * 60 * 60 * 24));
}

const DSRDetail = () => {
    const { id } = useParams<{ id: string }>();
    const navigate = useNavigate();
    const addToast = useToastStore((s) => s.addToast);
    const { data: dsr, isLoading } = useDSR(id || '');
    const approveMutation = useApproveDSR();
    const rejectMutation = useRejectDSR();
    const executeMutation = useExecuteDSR();

    const [showRejectModal, setShowRejectModal] = useState(false);
    const [rejectReason, setRejectReason] = useState('');

    const handleApprove = () => {
        if (!id) return;
        approveMutation.mutate(id, {
            onSuccess: () => addToast({ title: 'DSR approved', variant: 'success' }),
            onError: () => addToast({ title: 'Failed to approve DSR', variant: 'error' }),
        });
    };

    const handleReject = () => {
        if (!id || !rejectReason.trim()) return;
        rejectMutation.mutate({ id, reason: rejectReason.trim() }, {
            onSuccess: () => {
                addToast({ title: 'DSR rejected', variant: 'success' });
                setShowRejectModal(false);
                setRejectReason('');
            },
            onError: () => addToast({ title: 'Failed to reject DSR', variant: 'error' }),
        });
    };

    const handleExecute = () => {
        if (!id) return;
        executeMutation.mutate(id, {
            onSuccess: () => addToast({ title: 'DSR execution started', variant: 'success' }),
            onError: () => addToast({ title: 'Failed to execute DSR', variant: 'error' }),
        });
    };

    if (isLoading) {
        return (
            <div style={{ padding: '24px 32px' }}>
                <div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
                    {[1, 2, 3].map(i => (
                        <div key={i} style={{ height: '60px', backgroundColor: 'var(--bg-tertiary)', borderRadius: '8px', animation: 'pulse 1.5s ease-in-out infinite' }} />
                    ))}
                </div>
            </div>
        );
    }

    if (!dsr) {
        return (
            <div style={{ padding: '24px 32px', textAlign: 'center' }}>
                <h2 style={{ color: 'var(--text-primary)' }}>DSR Not Found</h2>
                <Button variant="ghost" onClick={() => navigate('/dsr')} style={{ marginTop: '12px' }}>
                    Back to DSR List
                </Button>
            </div>
        );
    }

    const days = getDaysRemaining(dsr.sla_deadline);
    const isOverdue = days < 0;
    const isPending = dsr.status === 'PENDING';
    const isApproved = dsr.status === 'APPROVED';
    const isInProgress = dsr.status === 'IN_PROGRESS';
    const isCompleted = dsr.status === 'COMPLETED';
    const isTerminal = ['COMPLETED', 'REJECTED', 'FAILED'].includes(dsr.status);

    // Calculate progress
    let progress = 0;
    if (isCompleted) progress = 100;
    else if (isInProgress) {
        if (dsr.tasks && dsr.tasks.length > 0) {
            const completed = dsr.tasks.filter(t => t.status === 'COMPLETED' || t.status === 'VERIFIED').length;
            progress = Math.round((completed / dsr.tasks.length) * 100);
        } else {
            progress = 50;
        }
    } else if (isApproved) progress = 10;
    else if (isPending) progress = 0;

    const taskColumns: Column<DSRTask>[] = [
        {
            key: 'data_source_id',
            header: 'Data Source',
            render: (row) => (
                <span style={{ fontSize: '0.8125rem', fontFamily: 'monospace', color: 'var(--text-primary)' }}>
                    {row.data_source_id.slice(0, 8)}…
                </span>
            ),
        },
        {
            key: 'task_type',
            header: 'Type',
            render: (row) => <span style={{ fontSize: '0.8125rem' }}>{row.task_type}</span>,
        },
        {
            key: 'status',
            header: 'Status',
            render: (row) => <StatusBadge label={row.status} />,
        },
        {
            key: 'error',
            header: 'Error',
            render: (row) => row.error ? (
                <span style={{ fontSize: '0.75rem', color: 'var(--status-danger)' }}>{row.error}</span>
            ) : <span style={{ color: 'var(--text-tertiary)' }}>—</span>,
        },
        {
            key: 'created_at',
            header: 'Created',
            render: (row) => (
                <span style={{ fontSize: '0.8125rem', color: 'var(--text-secondary)' }}>
                    {new Date(row.created_at).toLocaleString()}
                </span>
            ),
        },
    ];

    const cardStyle: React.CSSProperties = {
        backgroundColor: 'var(--bg-primary)',
        border: '1px solid var(--border-primary)',
        borderRadius: '8px',
        padding: '20px',
    };

    const labelStyle: React.CSSProperties = {
        fontSize: '0.75rem',
        fontWeight: 500,
        color: 'var(--text-tertiary)',
        textTransform: 'uppercase',
        letterSpacing: '0.05em',
        marginBottom: '4px',
    };

    return (
        <div style={{ padding: '24px 32px' }}>
            {/* Back Button */}
            <button
                onClick={() => navigate('/dsr')}
                style={{
                    display: 'flex', alignItems: 'center', gap: '6px', background: 'none',
                    border: 'none', cursor: 'pointer', color: 'var(--text-secondary)',
                    fontSize: '0.875rem', marginBottom: '16px', padding: 0,
                }}
            >
                <ArrowLeft size={16} /> Back to DSR Requests
            </button>

            {/* Header Card */}
            <div style={{ ...cardStyle, marginBottom: '20px' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', flexWrap: 'wrap', gap: '16px' }}>
                    <div style={{ flex: 1 }}>
                        <h1 style={{ fontSize: '1.375rem', fontWeight: 700, color: 'var(--text-primary)', marginBottom: '4px' }}>
                            {dsr.subject_name}
                        </h1>
                        <p style={{ fontSize: '0.875rem', color: 'var(--text-secondary)' }}>{dsr.subject_email}</p>
                        <div style={{ display: 'flex', gap: '8px', marginTop: '12px', flexWrap: 'wrap' }}>
                            <StatusBadge label={dsr.status} />
                            <span style={{
                                padding: '2px 10px', borderRadius: '4px', fontSize: '0.75rem', fontWeight: 600,
                                backgroundColor: '#dbeafe', color: '#1d4ed8',
                            }}>
                                {dsr.request_type}
                            </span>
                            <span style={{
                                padding: '2px 10px', borderRadius: '4px', fontSize: '0.75rem', fontWeight: 600,
                                backgroundColor: dsr.priority === 'HIGH' ? '#fee2e2' : dsr.priority === 'MEDIUM' ? '#fef3c7' : '#f1f5f9',
                                color: dsr.priority === 'HIGH' ? '#b91c1c' : dsr.priority === 'MEDIUM' ? '#b45309' : '#64748b',
                            }}>
                                {dsr.priority} Priority
                            </span>
                        </div>
                    </div>

                    {/* Progress Bar */}
                    <div style={{ width: '200px', display: 'flex', flexDirection: 'column', gap: '4px' }}>
                        <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '0.75rem', color: 'var(--text-secondary)' }}>
                            <span>Progress</span>
                            <span>{progress}%</span>
                        </div>
                        <div style={{ width: '100%', height: '8px', backgroundColor: 'var(--bg-tertiary)', borderRadius: '4px', overflow: 'hidden' }}>
                            <div style={{
                                width: `${progress}%`, height: '100%',
                                backgroundColor: isTerminal && dsr.status !== 'COMPLETED' ? 'var(--status-danger)' : 'var(--accent-primary)',
                                transition: 'width 0.5s ease-out'
                            }} />
                        </div>
                    </div>

                    {/* Actions */}
                    {!isTerminal && (
                        <div style={{ display: 'flex', gap: '8px' }}>
                            {isPending && (
                                <>
                                    <Button
                                        icon={<CheckCircle size={16} />}
                                        onClick={handleApprove}
                                        isLoading={approveMutation.isPending}
                                    >
                                        Approve
                                    </Button>
                                    <Button
                                        variant="danger"
                                        icon={<XCircle size={16} />}
                                        onClick={() => setShowRejectModal(true)}
                                    >
                                        Reject
                                    </Button>
                                </>
                            )}
                            {isApproved && (
                                <Button
                                    icon={<Play size={16} />}
                                    onClick={handleExecute}
                                    isLoading={executeMutation.isPending}
                                >
                                    Execute
                                </Button>
                            )}
                        </div>
                    )}
                </div>
            </div>

            {/* Info Grid */}
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', gap: '16px', marginBottom: '20px' }}>
                {/* SLA */}
                <div style={cardStyle}>
                    <div style={labelStyle}>SLA Deadline</div>
                    <div style={{
                        display: 'flex', alignItems: 'center', gap: '6px',
                        fontSize: '1rem', fontWeight: 600,
                        color: isOverdue ? 'var(--status-danger)' : days <= 3 ? 'var(--status-warning)' : 'var(--text-primary)',
                    }}>
                        {isOverdue ? <AlertTriangle size={16} /> : <Clock size={16} />}
                        {isOverdue ? `${Math.abs(days)} days overdue` : `${days} days remaining`}
                    </div>
                    <div style={{ fontSize: '0.8125rem', color: 'var(--text-tertiary)', marginTop: '4px' }}>
                        Due: {new Date(dsr.sla_deadline).toLocaleDateString()}
                    </div>
                </div>

                {/* Created */}
                <div style={cardStyle}>
                    <div style={labelStyle}>Created</div>
                    <div style={{ fontSize: '1rem', fontWeight: 500, color: 'var(--text-primary)' }}>
                        {new Date(dsr.created_at).toLocaleDateString()}
                    </div>
                    <div style={{ fontSize: '0.8125rem', color: 'var(--text-tertiary)', marginTop: '4px' }}>
                        {new Date(dsr.created_at).toLocaleTimeString()}
                    </div>
                </div>

                {/* Completed */}
                {dsr.completed_at && (
                    <div style={cardStyle}>
                        <div style={labelStyle}>Completed</div>
                        <div style={{ fontSize: '1rem', fontWeight: 500, color: 'var(--text-primary)' }}>
                            {new Date(dsr.completed_at).toLocaleDateString()}
                        </div>
                    </div>
                )}

                {/* Identifiers */}
                <div style={cardStyle}>
                    <div style={labelStyle}>Identifiers</div>
                    {Object.entries(dsr.subject_identifiers || {}).length > 0 ? (
                        Object.entries(dsr.subject_identifiers).map(([k, v]) => (
                            <div key={k} style={{ fontSize: '0.8125rem', color: 'var(--text-primary)' }}>
                                <span style={{ color: 'var(--text-tertiary)' }}>{k}:</span> {v}
                            </div>
                        ))
                    ) : (
                        <span style={{ fontSize: '0.8125rem', color: 'var(--text-tertiary)' }}>None</span>
                    )}
                </div>
            </div>

            {/* Timeline (Conceptual) */}
            <div style={{ ...cardStyle, marginBottom: '20px' }}>
                <h3 style={{ fontSize: '1rem', fontWeight: 600, color: 'var(--text-primary)', marginBottom: '16px' }}>Timeline</h3>
                <div style={{ display: 'flex', alignItems: 'center', gap: '0' }}>
                    <TimelineStep label="Received" date={dsr.created_at} completed={true} />
                    <TimelineDivider completed={isApproved || isInProgress || isCompleted} />
                    <TimelineStep label="Approved" date={isApproved || isInProgress || isCompleted ? dsr.updated_at : undefined} completed={isApproved || isInProgress || isCompleted} />
                    <TimelineDivider completed={isInProgress || isCompleted} />
                    <TimelineStep label="In Progress" completed={isInProgress || isCompleted} />
                    <TimelineDivider completed={isCompleted} />
                    <TimelineStep label="Completed" date={dsr.completed_at} completed={isCompleted} isLast />
                </div>
            </div>

            {/* Rejection Reason */}
            {dsr.reason && (
                <div style={{ ...cardStyle, marginBottom: '20px', borderColor: 'var(--status-danger)', backgroundColor: '#fef2f2' }}>
                    <div style={{ ...labelStyle, color: 'var(--status-danger)' }}>Rejection Reason</div>
                    <p style={{ color: 'var(--text-primary)', fontSize: '0.875rem' }}>{dsr.reason}</p>
                </div>
            )}

            {/* Task Breakdown */}
            <div style={{ ...cardStyle, marginBottom: '20px' }}>
                <h3 style={{ fontSize: '1rem', fontWeight: 600, color: 'var(--text-primary)', marginBottom: '16px' }}>
                    Task Breakdown ({dsr.tasks?.length || 0} tasks)
                </h3>
                <DataTable
                    columns={taskColumns}
                    data={dsr.tasks || []}
                    isLoading={false}
                    keyExtractor={(row) => row.id}
                    emptyTitle="No Tasks Yet"
                    emptyDescription="Tasks are created when a DSR is approved."
                />
            </div>

            {/* Result Panel (for completed DSRs) */}
            {dsr.status === 'COMPLETED' && dsr.tasks?.some(t => t.result) && (
                <div style={cardStyle}>
                    <h3 style={{ fontSize: '1rem', fontWeight: 600, color: 'var(--text-primary)', marginBottom: '12px' }}>
                        Execution Results
                    </h3>
                    {dsr.tasks.filter(t => t.result).map(task => (
                        <div key={task.id} style={{
                            marginBottom: '12px', padding: '12px', backgroundColor: 'var(--bg-secondary)',
                            borderRadius: '6px', border: '1px solid var(--border-primary)',
                        }}>
                            <div style={{ fontSize: '0.75rem', color: 'var(--text-tertiary)', marginBottom: '4px' }}>
                                Source: {task.data_source_id.slice(0, 8)}…
                            </div>
                            <pre style={{
                                fontSize: '0.75rem', color: 'var(--text-primary)',
                                whiteSpace: 'pre-wrap', wordBreak: 'break-all', margin: 0,
                            }}>
                                {JSON.stringify(task.result, null, 2)}
                            </pre>
                        </div>
                    ))}
                </div>
            )}

            {/* Reject Modal */}
            <Modal
                open={showRejectModal}
                onClose={() => setShowRejectModal(false)}
                title="Reject DSR"
                footer={
                    <div style={{ display: 'flex', gap: '8px', justifyContent: 'flex-end' }}>
                        <Button variant="ghost" onClick={() => setShowRejectModal(false)}>Cancel</Button>
                        <Button
                            variant="danger"
                            onClick={handleReject}
                            isLoading={rejectMutation.isPending}
                            disabled={!rejectReason.trim()}
                        >
                            Reject
                        </Button>
                    </div>
                }
            >
                <div>
                    <label style={{ display: 'block', fontSize: '0.8125rem', fontWeight: 500, color: 'var(--text-secondary)', marginBottom: '6px' }}>
                        Reason for rejection *
                    </label>
                    <textarea
                        value={rejectReason}
                        onChange={e => setRejectReason(e.target.value)}
                        rows={4}
                        placeholder="Explain why this DSR is being rejected..."
                        style={{
                            width: '100%', padding: '10px 12px',
                            border: '1px solid var(--border-primary)',
                            borderRadius: '6px', fontSize: '0.875rem',
                            backgroundColor: 'var(--bg-secondary)', color: 'var(--text-primary)',
                            resize: 'vertical',
                        }}
                    />
                </div>
            </Modal>
        </div>
    );
};

// Timeline Components
function TimelineStep({ label, date, completed }: { label: string; date?: string; completed: boolean; isLast?: boolean }) {
    return (
        <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', position: 'relative', zIndex: 1, minWidth: '80px' }}>
            <div style={{
                width: '24px', height: '24px', borderRadius: '50%',
                backgroundColor: completed ? 'var(--accent-primary)' : 'var(--bg-tertiary)',
                border: completed ? 'none' : '2px solid var(--border-primary)',
                display: 'flex', alignItems: 'center', justifyContent: 'center',
                color: 'white', marginBottom: '8px'
            }}>
                {completed && <CheckCircle size={14} />}
            </div>
            <div style={{ fontSize: '0.75rem', fontWeight: 600, color: completed ? 'var(--text-primary)' : 'var(--text-tertiary)' }}>{label}</div>
            {date && <div style={{ fontSize: '0.65rem', color: 'var(--text-tertiary)' }}>{new Date(date).toLocaleDateString()}</div>}
        </div>
    );
}

function TimelineDivider({ completed }: { completed: boolean }) {
    return (
        <div style={{ flex: 1, height: '2px', backgroundColor: completed ? 'var(--accent-primary)' : 'var(--border-primary)', transform: 'translateY(-14px)', margin: '0 -10px', zIndex: 0 }} />
    );
}

export default DSRDetail;
