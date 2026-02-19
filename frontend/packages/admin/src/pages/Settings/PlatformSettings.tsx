import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Settings, Save, Shield, PenTool, Database } from 'lucide-react';
import { adminService } from '@/services/adminService';
import { toast, Button } from '@datalens/shared';
import type { PlatformSettings } from '@/types/admin';

export default function PlatformSettingsPage() {
    const queryClient = useQueryClient();

    const { data: settings, isLoading } = useQuery({
        queryKey: ['platform-settings'],
        queryFn: adminService.getPlatformSettings,
    });

    const updateMut = useMutation({
        mutationFn: adminService.updatePlatformSettings,
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['platform-settings'] });
            toast.success('Settings updated successfully');
        },
        onError: () => toast.error('Failed to update settings'),
    });

    if (isLoading) return <div className="p-8 text-[var(--text-secondary)]">Loading settings...</div>;

    return (
        <div className="p-8 max-w-4xl mx-auto">
            <div className="mb-8">
                <h1 className="text-2xl font-bold text-[var(--text-primary)] flex items-center gap-2">
                    <Settings className="text-[var(--accent-primary)]" />
                    Platform Settings
                </h1>
                <p className="text-[var(--text-secondary)] mt-1">
                    Manage global configuration, branding, and security policies.
                </p>
            </div>

            {settings ? (
                <SettingsForm settings={settings} onSave={(s) => updateMut.mutate(s)} saving={updateMut.isPending} />
            ) : (
                <div className="text-red-500">Failed to load settings.</div>
            )}
        </div>
    );
}

function SettingsForm({ settings, onSave, saving }: { settings: PlatformSettings; onSave: (s: Partial<PlatformSettings>) => void; saving: boolean }) {
    // Local state for form
    const [branding, setBranding] = useState(settings.branding || { company_name: '', primary_color: '#000000', logo_url: '' });
    const [security, setSecurity] = useState(settings.security || { mfa_required: false, session_timeout_minutes: 60 });
    const [maintenance, setMaintenance] = useState(settings.maintenance || { enabled: false, message: '' });

    const handleSave = () => {
        onSave({ branding, security, maintenance });
    };

    return (
        <div className="space-y-6">
            {/* Branding Section */}
            <Section title="Branding & Appearance" icon={<PenTool size={20} />}>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                    <Field label="Company Name">
                        <input
                            value={branding.company_name}
                            onChange={e => setBranding({ ...branding, company_name: e.target.value })}
                            className={inputClass}
                        />
                    </Field>
                    <Field label="Primary Color">
                        <div className="flex gap-2">
                            <input
                                type="color"
                                value={branding.primary_color}
                                onChange={e => setBranding({ ...branding, primary_color: e.target.value })}
                                className="h-10 w-10 p-0 border-0 rounded cursor-pointer"
                            />
                            <input
                                value={branding.primary_color}
                                onChange={e => setBranding({ ...branding, primary_color: e.target.value })}
                                className={inputClass}
                            />
                        </div>
                    </Field>
                    <Field label="Logo URL" className="md:col-span-2">
                        <input
                            value={branding.logo_url}
                            onChange={e => setBranding({ ...branding, logo_url: e.target.value })}
                            className={inputClass}
                            placeholder="/assets/logo.png"
                        />
                    </Field>
                </div>
            </Section>

            {/* Security Section */}
            <Section title="Security Policies" icon={<Shield size={20} />}>
                <div className="space-y-4">
                    <div className="flex items-center justify-between p-4 bg-[var(--bg-secondary)] rounded-lg border border-[var(--border-primary)]">
                        <div>
                            <div className="font-medium text-[var(--text-primary)]">Enforce MFA</div>
                            <div className="text-sm text-[var(--text-secondary)]">Require Multi-Factor Authentication for all admin users</div>
                        </div>
                        <input
                            type="checkbox"
                            checked={security.mfa_required}
                            onChange={e => setSecurity({ ...security, mfa_required: e.target.checked })}
                            className="h-5 w-5 rounded border-gray-300 text-[var(--accent-primary)] focus:ring-[var(--accent-primary)]"
                        />
                    </div>
                    <Field label="Session Timeout (minutes)">
                        <input
                            type="number"
                            value={security.session_timeout_minutes}
                            onChange={e => setSecurity({ ...security, session_timeout_minutes: Number(e.target.value) })}
                            className={inputClass}
                        />
                    </Field>
                </div>
            </Section>

            {/* Maintenance Section */}
            <Section title="System Maintenance" icon={<Database size={20} />}>
                <div className="space-y-4">
                    <div className="flex items-center justify-between p-4 bg-[var(--bg-secondary)] rounded-lg border border-[var(--border-primary)]">
                        <div>
                            <div className="font-medium text-[var(--text-primary)]">Maintenance Mode</div>
                            <div className="text-sm text-[var(--text-secondary)]">Prevent login for non-admin users</div>
                        </div>
                        <input
                            type="checkbox"
                            checked={maintenance.enabled}
                            onChange={e => setMaintenance({ ...maintenance, enabled: e.target.checked })}
                            className="h-5 w-5 rounded border-gray-300 text-[var(--accent-primary)] focus:ring-[var(--accent-primary)]"
                        />
                    </div>
                    {maintenance.enabled && (
                        <Field label="Maintenance Message">
                            <textarea
                                value={maintenance.message}
                                onChange={e => setMaintenance({ ...maintenance, message: e.target.value })}
                                className={`${inputClass} min-h-[80px]`}
                            />
                        </Field>
                    )}
                </div>
            </Section>

            <div className="flex justify-end pt-4">
                <Button onClick={handleSave} disabled={saving} size="lg">
                    <Save size={18} className="mr-2" />
                    {saving ? 'Saving...' : 'Save Configuration'}
                </Button>
            </div>
        </div>
    );
}

function Section({ title, icon, children }: { title: string; icon: React.ReactNode; children: React.ReactNode }) {
    return (
        <div className="bg-[var(--bg-primary)] p-6 rounded-xl border border-[var(--border-primary)] shadow-sm">
            <h3 className="text-lg font-semibold text-[var(--text-primary)] mb-6 flex items-center gap-2 pb-4 border-b border-[var(--border-primary)]">
                <span className="text-[var(--text-secondary)]">{icon}</span>
                {title}
            </h3>
            {children}
        </div>
    );
}

function Field({ label, children, className = '' }: { label: string; children: React.ReactNode; className?: string }) {
    return (
        <div className={className}>
            <label className="block text-xs font-semibold text-[var(--text-secondary)] uppercase mb-2 ml-1">{label}</label>
            {children}
        </div>
    );
}

const inputClass = "w-full px-4 py-2.5 bg-[var(--bg-secondary)] border border-[var(--border-primary)] rounded-lg text-[var(--text-primary)] focus:outline-none focus:ring-2 focus:ring-[var(--accent-primary)] transition-all";
