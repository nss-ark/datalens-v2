import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { portalService } from '@/services/portalService';
import { ShieldAlert, FileText, Plus, ExternalLink, ChevronRight } from 'lucide-react';
import { format } from 'date-fns';
import { StatusBadge, Button, toast, Modal } from '@datalens/shared';
import { useNavigate } from 'react-router-dom';
import { AppealModal } from '@/components/AppealModal';
import { RequestSidebar } from '@/components/RequestSidebar';
import { NominationModal } from '@/components/NominationModal';
import { DSRRequestModal } from '@/components/DSRRequestModal';

/* ────────────────────────────────────────
   Inline Styles for Requests Page
   ──────────────────────────────────────── */
const styles = {
    container: {
        maxWidth: '1280px',
        margin: '0 auto',
        padding: '32px 16px',
        animation: 'fadeIn 0.5s ease-out',
    },
    layout: {
        display: 'flex',
        flexDirection: 'column' as const,
        gap: '32px',
    },
    // Desktop layout (applied via media query logic in CSS or inline conditional if sensitive)
    // For simplicity with inline styles, we'll rely on flex-wrap and width control

    mainColumn: {
        flex: 1,
        minWidth: 0, // Prevent flex item overflow
    },
    sidebarColumn: {
        width: '100%',
        maxWidth: '320px',
        flexShrink: 0,
    },

    header: {
        marginBottom: '32px',
    },
    title: {
        fontSize: '30px',
        fontWeight: 700,
        color: '#111827',
        marginBottom: '8px',
        letterSpacing: '-0.02em',
    },
    subtitle: {
        fontSize: '16px',
        color: '#6b7280',
    },

    // Empty State
    emptyState: {
        backgroundColor: '#ffffff',
        borderRadius: '16px',
        border: '1px solid #e5e7eb',
        padding: '64px 24px',
        textAlign: 'center' as const,
        display: 'flex',
        flexDirection: 'column' as const,
        alignItems: 'center',
        justifyContent: 'center',
        minHeight: '400px',
    },
    emptyIcon: {
        width: '64px',
        height: '64px',
        backgroundColor: '#eff6ff', // blue-50
        borderRadius: '20px',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        marginBottom: '24px',
        color: '#3b82f6',
    },
    emptyTitle: {
        fontSize: '20px',
        fontWeight: 600,
        color: '#111827',
        marginBottom: '12px',
    },
    emptyDesc: {
        fontSize: '15px',
        color: '#6b7280', // slate-500
        maxWidth: '420px',
        lineHeight: 1.6,
        marginBottom: '32px',
    },

    // Grievance Banner
    grievanceBanner: {
        marginTop: '32px',
        backgroundColor: '#ffffff',
        borderRadius: '16px',
        border: '1px solid #e5e7eb', // slate-200
        padding: '24px',
        display: 'flex',
        flexDirection: 'column' as const, // stacks on mobile
        gap: '24px',
        alignItems: 'flex-start',
        boxShadow: '0 4px 6px -1px rgba(0, 0, 0, 0.02), 0 2px 4px -1px rgba(0, 0, 0, 0.02)',
    },
    grievanceHeader: {
        display: 'flex',
        gap: '16px',
        alignItems: 'flex-start',
    },
    grievanceIcon: {
        width: '40px',
        height: '40px',
        backgroundColor: '#fff7ed', // orange-50
        borderRadius: '10px',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        color: '#f97316', // orange-500
        flexShrink: 0,
    },

    // Responsive helper
    mobileStack: {
        flexDirection: 'column' as const,
    }
};

const PortalRequests = () => {
    const navigate = useNavigate();
    const queryClient = useQueryClient();
    const [appealId, setAppealId] = useState<string | null>(null);
    const [isAppealOpen, setIsAppealOpen] = useState(false);
    const [isNominationOpen, setIsNominationOpen] = useState(false);
    const [isTypeSelectorOpen, setIsTypeSelectorOpen] = useState(false);

    // DSR Modal State
    const [dsrModal, setDsrModal] = useState<{
        isOpen: boolean;
        type: 'ACCESS' | 'CORRECTION' | 'ERASURE';
    }>({ isOpen: false, type: 'ACCESS' });

    const { data: requests, isLoading } = useQuery({
        queryKey: ['portal-requests'],
        queryFn: () => portalService.listRequests(),
    });

    const items = requests?.items || [];

    // Nomination Mutation
    const nominationMutation = useMutation({
        mutationFn: (data: any) => {
            const description = `NOMINATION REQUEST\n\nNominee Name: ${data.nomineeName}\nRelationship: ${data.relationship}\nContact: ${data.contact}\n\nThis is a formal request to nominate the above individual as a beneficiary for my data rights.`;
            return portalService.createRequest({
                type: 'NOMINATION',
                description: description
            });
        },
        onSuccess: () => {
            toast.success('Nomination request submitted successfully');
            queryClient.invalidateQueries({ queryKey: ['portal-requests'] });
            setIsNominationOpen(false);
        },
        onError: () => {
            toast.error('Failed to submit nomination request');
        }
    });

    // Generic DSR Mutation (Access, Correction, Erasure)
    const dsrMutation = useMutation({
        mutationFn: (data: { description: string; file?: File | null }) => {
            let fullDescription = data.description;
            if (data.file) {
                fullDescription += `\n\n[Attachment: ${data.file.name}]`;
            }

            return portalService.createRequest({
                type: dsrModal.type,
                description: fullDescription
            });
        },
        onSuccess: () => {
            toast.success('Request submitted successfully');
            queryClient.invalidateQueries({ queryKey: ['portal-requests'] });
            setDsrModal(prev => ({ ...prev, isOpen: false }));
        },
        onError: () => {
            toast.error('Failed to submit request');
        }
    });

    const handleOpenAppeal = (id: string) => {
        setAppealId(id);
        setIsAppealOpen(true);
    };

    const handleRequestTypeSelect = (type: 'ACCESS' | 'CORRECTION' | 'ERASURE' | 'NOMINATION' | 'GRIEVANCE') => {
        setIsTypeSelectorOpen(false); // Close selector if open
        if (type === 'NOMINATION') {
            setIsNominationOpen(true);
        } else if (type === 'GRIEVANCE') {
            navigate('/grievance/new');
        } else {
            setDsrModal({ isOpen: true, type });
        }
    };

    const handleNominationSubmit = (data: any) => {
        nominationMutation.mutate(data);
    };

    const handleDSRSubmit = (data: { description: string; file?: File | null }) => {
        dsrMutation.mutate(data);
    };

    return (
        <div style={styles.container}>
            <div className="flex flex-col lg:flex-row gap-8">
                {/* ── Main Content (Left) ── */}
                <div style={styles.mainColumn}>
                    {/* Header */}
                    <div className="flex justify-between items-start mb-8">
                        <div>
                            <h1 style={styles.title}>My Requests</h1>
                            <p style={styles.subtitle}>Track and manage your data privacy requests.</p>
                        </div>
                        {items.length > 0 && (
                            <button
                                onClick={() => setIsTypeSelectorOpen(true)}
                                className="hidden sm:flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors text-sm font-medium shadow-sm"
                            >
                                <Plus size={16} />
                                New Request
                            </button>
                        )}
                    </div>

                    {/* Content Area */}
                    {isLoading ? (
                        <div className="space-y-4">
                            {[1, 2, 3].map(i => (
                                <div key={i} className="h-24 bg-slate-100 rounded-xl animate-pulse" />
                            ))}
                        </div>
                    ) : items.length === 0 ? (
                        /* Empty State */
                        <div style={styles.emptyState}>
                            <div style={styles.emptyIcon}>
                                <FileText size={32} />
                            </div>
                            <h3 style={styles.emptyTitle}>No requests yet</h3>
                            <p style={styles.emptyDesc}>
                                You haven't submitted any data requests. Exercise your rights to access, correct, or erase your personal data securely.
                            </p>
                            <button
                                onClick={() => setIsTypeSelectorOpen(true)}
                                className="flex items-center gap-2 text-blue-600 font-semibold hover:text-blue-700 transition-colors"
                            >
                                Submit your first request
                                <ExternalLink size={16} />
                            </button>
                        </div>
                    ) : (
                        /* List View */
                        <div className="bg-white rounded-xl border border-slate-200 overflow-hidden shadow-sm">
                            <div className="overflow-x-auto">
                                <table className="w-full text-left text-sm">
                                    <thead>
                                        <tr className="border-b border-slate-100 bg-slate-50/80">
                                            <th className="px-6 py-4 font-semibold text-slate-500 uppercase tracking-wider text-xs">Type</th>
                                            <th className="px-6 py-4 font-semibold text-slate-500 uppercase tracking-wider text-xs">Description</th>
                                            <th className="px-6 py-4 font-semibold text-slate-500 uppercase tracking-wider text-xs">Date</th>
                                            <th className="px-6 py-4 font-semibold text-slate-500 uppercase tracking-wider text-xs">Status</th>
                                            <th className="px-6 py-4 font-semibold text-slate-500 uppercase tracking-wider text-xs">Actions</th>
                                        </tr>
                                    </thead>
                                    <tbody className="divide-y divide-slate-100">
                                        {items.map((req) => (
                                            <tr key={req.id} className="hover:bg-slate-50/50 transition-colors">
                                                <td className="px-6 py-4 font-medium text-slate-900 whitespace-nowrap">
                                                    <div className="flex items-center gap-2">
                                                        <div className={`w-2 h-2 rounded-full ${req.type === 'NOMINATION' ? 'bg-orange-500' :
                                                            req.type === 'GRIEVANCE' ? 'bg-red-500' : 'bg-blue-500'
                                                            }`} />
                                                        {req.type}
                                                    </div>
                                                </td>
                                                <td className="px-6 py-4 text-slate-500 max-w-xs truncate" title={req.description}>
                                                    {req.description}
                                                </td>
                                                <td className="px-6 py-4 text-slate-500 whitespace-nowrap">
                                                    {format(new Date(req.submitted_at), 'MMM d, yyyy')}
                                                </td>
                                                <td className="px-6 py-4">
                                                    <StatusBadge label={req.status} />
                                                </td>
                                                <td className="px-6 py-4">
                                                    {/* Action buttons (Appeal, etc.) */}
                                                    {req.status === 'REJECTED' && (
                                                        <Button
                                                            variant="outline"
                                                            size="sm"
                                                            onClick={() => handleOpenAppeal(req.id)}
                                                            className="text-orange-600 border-orange-200 hover:bg-orange-50"
                                                        >
                                                            Appeal
                                                        </Button>
                                                    )}
                                                </td>
                                            </tr>
                                        ))}
                                    </tbody>
                                </table>
                            </div>
                        </div>
                    )}

                    {/* Grievance Banner */}
                    <div style={styles.grievanceBanner} className="flex-col sm:flex-row items-center sm:items-center justify-between">
                        <div style={styles.grievanceHeader} className="flex-1">
                            <div style={styles.grievanceIcon}>
                                <ShieldAlert size={20} />
                            </div>
                            <div>
                                <h4 className="text-base font-semibold text-slate-900 mb-1">Submit Grievance</h4>
                                <p className="text-sm text-slate-600 leading-relaxed max-w-2xl">
                                    If you believe your personal data has been mishandled or your rights have not been respected,
                                    you can raise a formal grievance with our Data Protection Officer.
                                </p>
                            </div>
                        </div>
                        <Button
                            variant="outline"
                            className="shrink-0 mt-4 sm:mt-0 bg-white"
                            onClick={() => navigate('/grievance/new')}
                        >
                            Start Grievance Request
                        </Button>
                    </div>

                </div>

                {/* ── Sidebar (Right) ── */}
                <div style={styles.sidebarColumn} className="hidden lg:block">
                    <RequestSidebar onRequestTypeSelect={handleRequestTypeSelect} />
                </div>
            </div>

            {/* Modals */}
            {appealId && (
                <AppealModal
                    isOpen={isAppealOpen}
                    dprId={appealId}
                    onClose={() => setIsAppealOpen(false)}
                    onSuccess={() => queryClient.invalidateQueries({ queryKey: ['portal-requests'] })}
                />
            )}

            <NominationModal
                isOpen={isNominationOpen}
                onClose={() => setIsNominationOpen(false)}
                onSubmit={handleNominationSubmit}
            />

            <DSRRequestModal
                isOpen={dsrModal.isOpen}
                onClose={() => setDsrModal(prev => ({ ...prev, isOpen: false }))}
                type={dsrModal.type}
                onSubmit={handleDSRSubmit}
                isLoading={dsrMutation.isPending}
            />

            {/* Request Type Selector Modal */}
            <Modal
                open={isTypeSelectorOpen}
                onClose={() => setIsTypeSelectorOpen(false)}
                title="Select Request Type"
            >
                <div style={{ padding: '28px', display: 'flex', flexDirection: 'column', gap: '20px' }}>
                    {/* Subtitle */}
                    <p style={{ fontSize: '14px', color: '#64748b', lineHeight: 1.6, margin: 0 }}>
                        Choose the type of data rights request you'd like to submit. Each request is processed securely under DPDPA.
                    </p>

                    {/* Cards Grid */}
                    <div style={{ display: 'grid', gridTemplateColumns: 'repeat(2, 1fr)', gap: '14px' }}>
                        {[
                            { type: 'ACCESS', label: 'Right to Access', desc: 'Request a copy of your personal data', icon: '📥', bg: '#eef2ff', border: '#c7d2fe', accent: '#4f46e5' },
                            { type: 'CORRECTION', label: 'Right to Correction', desc: 'Update or fix inaccurate data', icon: '✏️', bg: '#ecfdf5', border: '#a7f3d0', accent: '#059669' },
                            { type: 'ERASURE', label: 'Right to Erasure', desc: 'Request deletion of your data', icon: '🗑️', bg: '#fef2f2', border: '#fecaca', accent: '#dc2626' },
                            { type: 'NOMINATION', label: 'Nomination', desc: 'Designate a trusted beneficiary', icon: '👤', bg: '#fff7ed', border: '#fed7aa', accent: '#ea580c' },
                            { type: 'GRIEVANCE', label: 'Submit Grievance', desc: 'Report an issue or concern', icon: '⚠️', bg: '#fefce8', border: '#fde68a', accent: '#ca8a04' },
                        ].map((item) => (
                            <button
                                key={item.type}
                                onClick={() => handleRequestTypeSelect(item.type as any)}
                                style={{
                                    display: 'flex',
                                    alignItems: 'flex-start',
                                    gap: '14px',
                                    padding: '18px',
                                    borderRadius: '14px',
                                    border: `1.5px solid ${item.border}`,
                                    background: item.bg,
                                    cursor: 'pointer',
                                    textAlign: 'left',
                                    transition: 'all 0.2s ease',
                                    width: '100%',
                                }}
                                onMouseEnter={(e) => {
                                    e.currentTarget.style.transform = 'translateY(-2px)';
                                    e.currentTarget.style.boxShadow = `0 8px 25px -5px ${item.accent}22`;
                                    e.currentTarget.style.borderColor = item.accent;
                                }}
                                onMouseLeave={(e) => {
                                    e.currentTarget.style.transform = 'translateY(0)';
                                    e.currentTarget.style.boxShadow = 'none';
                                    e.currentTarget.style.borderColor = item.border;
                                }}
                            >
                                <div style={{
                                    width: '44px',
                                    height: '44px',
                                    borderRadius: '12px',
                                    background: 'white',
                                    display: 'flex',
                                    alignItems: 'center',
                                    justifyContent: 'center',
                                    fontSize: '22px',
                                    flexShrink: 0,
                                    boxShadow: '0 1px 3px rgba(0,0,0,0.06)',
                                    border: '1px solid rgba(0,0,0,0.04)',
                                }}>
                                    {item.icon}
                                </div>
                                <div style={{ display: 'flex', flexDirection: 'column', gap: '4px', minWidth: 0 }}>
                                    <span style={{ fontSize: '14px', fontWeight: 700, color: '#0f172a' }}>{item.label}</span>
                                    <span style={{ fontSize: '12px', color: '#64748b', lineHeight: 1.5 }}>{item.desc}</span>
                                    <div style={{ display: 'flex', alignItems: 'center', gap: '4px', fontSize: '12px', fontWeight: 600, color: item.accent, marginTop: '6px' }}>
                                        Select <ChevronRight size={13} />
                                    </div>
                                </div>
                            </button>
                        ))}
                    </div>
                </div>
            </Modal>
        </div>
    );
};

export default PortalRequests;
