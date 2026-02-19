import { useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
    ArrowLeft, Building2, Globe, MapPin, Shield, Zap, Calendar,
    Save, ToggleLeft, ToggleRight, CreditCard, Package
} from 'lucide-react';
import { adminService } from '@/services/adminService';
import { toast, Button, StatusBadge } from '@datalens/shared';
import type { Tenant, Subscription, ModuleAccess, ModuleName, ModuleAccessInput } from '@/types/admin';

type Tab = 'overview' | 'subscription' | 'modules';

const MODULE_LABELS: Record<ModuleName, { label: string; description: string; icon: string }> = {
    PII_DISCOVERY: { label: 'PII Discovery', description: 'Scan & classify personal data across sources', icon: '🔍' },
    DSR_MANAGEMENT: { label: 'DSR Management', description: 'Handle data subject access & deletion requests', icon: '📋' },
    CONSENT_MANAGER: { label: 'Consent Manager', description: 'Track & manage user consent lifecycle', icon: '✅' },
    BREACH_TRACKER: { label: 'Breach Tracker', description: 'Detect, respond to & report data breaches', icon: '🚨' },
    DATA_GOVERNANCE: { label: 'Data Governance', description: 'Policies, lineage & compliance mapping', icon: '📊' },
    AI_CLASSIFICATION: { label: 'AI Classification', description: 'ML-powered PII detection & categorization', icon: '🤖' },
    ADVANCED_ANALYTICS: { label: 'Advanced Analytics', description: 'Deep insights, trends & compliance scoring', icon: '📈' },
    AUDIT_TRAIL: { label: 'Audit Trail', description: 'Immutable, tamper-proof event logging', icon: '🔒' },
};

const PLAN_COLORS: Record<string, string> = {
    FREE: 'from-gray-500 to-gray-600',
    STARTER: 'from-blue-500 to-blue-600',
    PROFESSIONAL: 'from-purple-500 to-purple-600',
    ENTERPRISE: 'from-amber-500 to-amber-600',
};

export default function TenantDetail() {
    const { id } = useParams<{ id: string }>();
    const navigate = useNavigate();
    const queryClient = useQueryClient();
    const [activeTab, setActiveTab] = useState<Tab>('overview');

    // Queries
    const { data: tenant, isLoading: loadingTenant } = useQuery({
        queryKey: ['tenant', id],
        queryFn: () => adminService.getTenant(id!),
        enabled: !!id,
    });

    const { data: subscription, isLoading: loadingSub } = useQuery({
        queryKey: ['tenant-subscription', id],
        queryFn: () => adminService.getSubscription(id!),
        enabled: !!id,
    });

    const { data: modules, isLoading: loadingModules } = useQuery({
        queryKey: ['tenant-modules', id],
        queryFn: () => adminService.getModuleAccess(id!),
        enabled: !!id,
    });

    // Mutations
    const updateTenantMut = useMutation({
        mutationFn: (data: Partial<Tenant>) => adminService.updateTenant(id!, data),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['tenant', id] });
            toast.success('Tenant updated');
        },
        onError: () => toast.error('Failed to update tenant'),
    });

    const updateSubMut = useMutation({
        mutationFn: (data: Partial<Subscription>) => adminService.updateSubscription(id!, data),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['tenant-subscription', id] });
            queryClient.invalidateQueries({ queryKey: ['tenant-modules', id] });
            toast.success('Subscription updated');
        },
        onError: () => toast.error('Failed to update subscription'),
    });

    const updateModulesMut = useMutation({
        mutationFn: (data: ModuleAccessInput[]) => adminService.updateModuleAccess(id!, data),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['tenant-modules', id] });
            toast.success('Module access updated');
        },
        onError: () => toast.error('Failed to update modules'),
    });

    const handleModuleToggle = (moduleName: ModuleName, currentEnabled: boolean) => {
        if (!modules) return;
        const updated: ModuleAccessInput[] = modules.map(m => ({
            module_name: m.module_name,
            enabled: m.module_name === moduleName ? !currentEnabled : m.enabled,
        }));
        updateModulesMut.mutate(updated);
    };

    if (loadingTenant || !tenant) {
        return (
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', minHeight: '60vh' }}>
                <div style={{ color: 'var(--text-secondary)' }}>Loading tenant...</div>
            </div>
        );
    }

    const tabs: { key: Tab; label: string; icon: React.ReactNode }[] = [
        { key: 'overview', label: 'Overview', icon: <Building2 size={16} /> },
        { key: 'subscription', label: 'Subscription', icon: <CreditCard size={16} /> },
        { key: 'modules', label: 'Modules', icon: <Package size={16} /> },
    ];

    return (
        <div style={{ padding: '1.5rem 2rem' }}>
            {/* Header */}
            <div style={{ display: 'flex', alignItems: 'center', gap: '1rem', marginBottom: '1.5rem' }}>
                <button
                    onClick={() => navigate('/tenants')}
                    style={{
                        background: 'none', border: 'none', cursor: 'pointer',
                        color: 'var(--text-secondary)', display: 'flex', alignItems: 'center',
                        padding: '0.5rem', borderRadius: '0.5rem',
                    }}
                    onMouseOver={e => (e.currentTarget.style.background = 'var(--bg-secondary)')}
                    onMouseOut={e => (e.currentTarget.style.background = 'none')}
                >
                    <ArrowLeft size={20} />
                </button>
                <div>
                    <h1 style={{ fontSize: '1.5rem', fontWeight: 700, color: 'var(--text-primary)', margin: 0 }}>
                        {tenant.name}
                    </h1>
                    <p style={{ fontSize: '0.875rem', color: 'var(--text-secondary)', margin: 0 }}>
                        {tenant.domain} · {tenant.industry} · {tenant.country}
                    </p>
                </div>
                <div style={{ marginLeft: 'auto' }}>
                    <StatusBadge label={tenant.status} />
                </div>
            </div>

            {/* Tabs */}
            <div style={{
                display: 'flex', gap: '0.25rem', marginBottom: '1.5rem',
                borderBottom: '1px solid var(--border-primary)', paddingBottom: '0',
            }}>
                {tabs.map(tab => (
                    <button
                        key={tab.key}
                        onClick={() => setActiveTab(tab.key)}
                        style={{
                            display: 'flex', alignItems: 'center', gap: '0.5rem',
                            padding: '0.75rem 1.25rem', background: 'none', border: 'none',
                            cursor: 'pointer', fontSize: '0.875rem', fontWeight: 500,
                            color: activeTab === tab.key ? 'var(--accent-primary)' : 'var(--text-secondary)',
                            borderBottom: activeTab === tab.key ? '2px solid var(--accent-primary)' : '2px solid transparent',
                            marginBottom: '-1px', transition: 'all 0.2s',
                        }}
                    >
                        {tab.icon} {tab.label}
                    </button>
                ))}
            </div>

            {/* Tab Content */}
            {activeTab === 'overview' && (
                <OverviewTab tenant={tenant} onSave={data => updateTenantMut.mutate(data)} saving={updateTenantMut.isPending} />
            )}
            {activeTab === 'subscription' && (
                <SubscriptionTab
                    subscription={subscription}
                    loading={loadingSub}
                    onSave={data => updateSubMut.mutate(data)}
                    saving={updateSubMut.isPending}
                />
            )}
            {activeTab === 'modules' && (
                <ModulesTab
                    modules={modules ?? []}
                    loading={loadingModules}
                    onToggle={handleModuleToggle}
                    saving={updateModulesMut.isPending}
                />
            )}
        </div>
    );
}

// ===========================================================================
// Overview Tab
// ===========================================================================

function OverviewTab({ tenant, onSave, saving }: { tenant: Tenant; onSave: (data: Partial<Tenant>) => void; saving: boolean }) {
    const [name, setName] = useState(tenant.name);
    const [industry, setIndustry] = useState(tenant.industry);
    const [country, setCountry] = useState(tenant.country);
    const [status, setStatus] = useState(tenant.status);

    const dirty = name !== tenant.name || industry !== tenant.industry || country !== tenant.country || status !== tenant.status;

    return (
        <div style={{ maxWidth: '640px' }}>
            <div style={{ display: 'grid', gap: '1.25rem' }}>
                <FieldGroup label="Organization Name" icon={<Building2 size={16} />}>
                    <input
                        value={name} onChange={e => setName(e.target.value)}
                        style={inputStyle}
                    />
                </FieldGroup>
                <FieldGroup label="Domain" icon={<Globe size={16} />}>
                    <input value={tenant.domain} disabled style={{ ...inputStyle, opacity: 0.6, cursor: 'not-allowed' }} />
                </FieldGroup>
                <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem' }}>
                    <FieldGroup label="Industry" icon={<Shield size={16} />}>
                        <select value={industry} onChange={e => setIndustry(e.target.value)} style={inputStyle}>
                            {['GENERAL', 'HEALTHCARE', 'FINANCE', 'EDUCATION', 'TECHNOLOGY', 'RETAIL', 'GOVERNMENT'].map(v => (
                                <option key={v} value={v}>{v}</option>
                            ))}
                        </select>
                    </FieldGroup>
                    <FieldGroup label="Country" icon={<MapPin size={16} />}>
                        <select value={country} onChange={e => setCountry(e.target.value)} style={inputStyle}>
                            {['IN', 'US', 'UK', 'DE', 'SG', 'AU'].map(v => (
                                <option key={v} value={v}>{v}</option>
                            ))}
                        </select>
                    </FieldGroup>
                </div>
                <FieldGroup label="Status" icon={<Zap size={16} />}>
                    <select value={status} onChange={e => setStatus(e.target.value as Tenant['status'])} style={inputStyle}>
                        <option value="ACTIVE">Active</option>
                        <option value="SUSPENDED">Suspended</option>
                        <option value="DELETED">Deleted</option>
                    </select>
                </FieldGroup>
                {dirty && (
                    <Button onClick={() => onSave({ name, industry, country, status })} disabled={saving}>
                        <Save size={16} style={{ marginRight: '0.5rem' }} />
                        {saving ? 'Saving...' : 'Save Changes'}
                    </Button>
                )}
            </div>
        </div>
    );
}

// ===========================================================================
// Subscription Tab
// ===========================================================================

function SubscriptionTab({
    subscription, loading, onSave, saving,
}: { subscription?: Subscription; loading: boolean; onSave: (data: Partial<Subscription>) => void; saving: boolean }) {
    const [plan, setPlan] = useState(subscription?.plan ?? 'FREE');
    const [autoRevoke, setAutoRevoke] = useState(subscription?.auto_revoke ?? true);
    const [billingStart, setBillingStart] = useState(subscription?.billing_start?.split('T')[0] ?? '');
    const [billingEnd, setBillingEnd] = useState(subscription?.billing_end?.split('T')[0] ?? '');

    if (loading) return <div style={{ color: 'var(--text-secondary)' }}>Loading subscription...</div>;

    const dirty = plan !== subscription?.plan ||
        autoRevoke !== subscription?.auto_revoke ||
        billingStart !== (subscription?.billing_start?.split('T')[0] ?? '') ||
        billingEnd !== (subscription?.billing_end?.split('T')[0] ?? '');

    return (
        <div style={{ maxWidth: '640px' }}>
            {/* Plan selector */}
            <div style={{ marginBottom: '1.5rem' }}>
                <label style={labelStyle}>Current Plan</label>
                <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: '0.75rem', marginTop: '0.5rem' }}>
                    {(['FREE', 'STARTER', 'PROFESSIONAL', 'ENTERPRISE'] as const).map(p => (
                        <button
                            key={p}
                            onClick={() => setPlan(p)}
                            style={{
                                padding: '1rem', borderRadius: '0.75rem', cursor: 'pointer',
                                border: plan === p ? '2px solid var(--accent-primary)' : '1px solid var(--border-primary)',
                                background: plan === p ? 'var(--bg-tertiary)' : 'var(--bg-secondary)',
                                textAlign: 'center', transition: 'all 0.2s',
                            }}
                        >
                            <div style={{
                                fontSize: '0.75rem', fontWeight: 700, letterSpacing: '0.05em',
                                background: `linear-gradient(135deg, ${PLAN_COLORS[p]?.split(' ')[0]?.replace('from-', '') || 'gray'})`,
                                WebkitBackgroundClip: 'text',
                                color: plan === p ? 'var(--accent-primary)' : 'var(--text-secondary)',
                            }}>
                                {p}
                            </div>
                        </button>
                    ))}
                </div>
            </div>

            {/* Billing info */}
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem', marginBottom: '1.5rem' }}>
                <FieldGroup label="Billing Start" icon={<Calendar size={16} />}>
                    <input
                        type="date"
                        value={billingStart}
                        onChange={e => setBillingStart(e.target.value)}
                        style={inputStyle}
                        placeholder="Select start date"
                    />
                </FieldGroup>
                <FieldGroup label="Billing End" icon={<Calendar size={16} />}>
                    <input
                        type="date"
                        value={billingEnd}
                        onChange={e => setBillingEnd(e.target.value)}
                        style={inputStyle}
                        placeholder="Select end date"
                    />
                </FieldGroup>
            </div>

            {/* Auto-revoke toggle */}
            <div style={{
                display: 'flex', alignItems: 'center', justifyContent: 'space-between',
                padding: '1rem 1.25rem', borderRadius: '0.75rem',
                background: 'var(--bg-secondary)', border: '1px solid var(--border-primary)',
                marginBottom: '1.5rem',
            }}>
                <div>
                    <div style={{ fontWeight: 600, color: 'var(--text-primary)', fontSize: '0.875rem' }}>
                        Auto-Revoke on Expiry
                    </div>
                    <div style={{ fontSize: '0.75rem', color: 'var(--text-secondary)' }}>
                        Automatically suspend tenant when billing period ends
                    </div>
                </div>
                <button
                    onClick={() => setAutoRevoke(!autoRevoke)}
                    style={{ background: 'none', border: 'none', cursor: 'pointer', color: autoRevoke ? 'var(--accent-primary)' : 'var(--text-secondary)' }}
                >
                    {autoRevoke ? <ToggleRight size={28} /> : <ToggleLeft size={28} />}
                </button>
            </div>

            {dirty && (
                <Button onClick={() => onSave({
                    plan,
                    auto_revoke: autoRevoke,
                    billing_start: billingStart ? new Date(billingStart).toISOString() : undefined,
                    billing_end: billingEnd ? new Date(billingEnd).toISOString() : undefined
                })} disabled={saving}>
                    <Save size={16} style={{ marginRight: '0.5rem' }} />
                    {saving ? 'Saving...' : 'Update Subscription'}
                </Button>
            )}
        </div>
    );
}

// ===========================================================================
// Modules Tab
// ===========================================================================

function ModulesTab({
    modules, loading, onToggle, saving,
}: { modules: ModuleAccess[]; loading: boolean; onToggle: (name: ModuleName, enabled: boolean) => void; saving: boolean }) {
    if (loading) return <div style={{ color: 'var(--text-secondary)' }}>Loading modules...</div>;

    return (
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(280px, 1fr))', gap: '1rem' }}>
            {modules.map(m => {
                const meta = MODULE_LABELS[m.module_name] ?? { label: m.module_name, description: '', icon: '📦' };
                return (
                    <div
                        key={m.module_name}
                        style={{
                            padding: '1.25rem', borderRadius: '0.75rem',
                            background: 'var(--bg-secondary)',
                            border: m.enabled ? '1px solid var(--accent-primary)' : '1px solid var(--border-primary)',
                            opacity: m.enabled ? 1 : 0.7,
                            transition: 'all 0.2s',
                        }}
                    >
                        <div style={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between' }}>
                            <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
                                <span style={{ fontSize: '1.5rem' }}>{meta.icon}</span>
                                <div>
                                    <div style={{ fontWeight: 600, color: 'var(--text-primary)', fontSize: '0.875rem' }}>
                                        {meta.label}
                                    </div>
                                    <div style={{ fontSize: '0.75rem', color: 'var(--text-secondary)', marginTop: '0.25rem' }}>
                                        {meta.description}
                                    </div>
                                </div>
                            </div>
                            <button
                                onClick={() => onToggle(m.module_name, m.enabled)}
                                disabled={saving}
                                style={{
                                    background: 'none', border: 'none', cursor: saving ? 'wait' : 'pointer',
                                    color: m.enabled ? 'var(--accent-primary)' : 'var(--text-secondary)',
                                    flexShrink: 0,
                                }}
                            >
                                {m.enabled ? <ToggleRight size={24} /> : <ToggleLeft size={24} />}
                            </button>
                        </div>
                    </div>
                );
            })}
        </div>
    );
}

// ===========================================================================
// Helpers
// ===========================================================================

function FieldGroup({ label, icon, children }: { label: string; icon: React.ReactNode; children: React.ReactNode }) {
    return (
        <div>
            <label style={labelStyle}>
                <span style={{ display: 'inline-flex', alignItems: 'center', gap: '0.375rem' }}>
                    {icon} {label}
                </span>
            </label>
            {children}
        </div>
    );
}

const labelStyle: React.CSSProperties = {
    display: 'block', fontSize: '0.75rem', fontWeight: 600,
    color: 'var(--text-secondary)', marginBottom: '0.375rem',
    textTransform: 'uppercase', letterSpacing: '0.05em',
};

const inputStyle: React.CSSProperties = {
    width: '100%', padding: '0.625rem 0.875rem', borderRadius: '0.5rem',
    border: '1px solid var(--border-primary)', background: 'var(--bg-secondary)',
    color: 'var(--text-primary)', fontSize: '0.875rem', outline: 'none',
    boxSizing: 'border-box',
};
