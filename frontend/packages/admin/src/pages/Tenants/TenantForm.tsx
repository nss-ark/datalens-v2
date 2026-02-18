import { useState } from 'react';
import { adminService } from '@/services/adminService';
import { toast } from '@datalens/shared';
import { Button } from '@datalens/shared';
import { cn } from '@datalens/shared';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from '@datalens/shared';
import { Check, Building2, Mail, CreditCard, ArrowRight, ArrowLeft, Sparkles } from 'lucide-react';
import type { CreateTenantInput } from '@/types/admin';

interface TenantFormProps {
    isOpen: boolean;
    onClose: () => void;
    onSuccess: () => void;
}

const PLANS = [
    {
        id: 'FREE',
        name: 'Free',
        price: '$0',
        period: '/mo',
        description: 'For small teams getting started',
        features: ['Up to 3 users', '7-day log retention', 'Basic PII scanning'],
        accent: 'from-zinc-500 to-zinc-600',
        badge: null,
    },
    {
        id: 'STARTER',
        name: 'Starter',
        price: '$49',
        period: '/mo',
        description: 'Growing businesses with more needs',
        features: ['Up to 10 users', '30-day retention', 'DSR management', 'Consent widgets'],
        accent: 'from-emerald-500 to-emerald-600',
        badge: null,
    },
    {
        id: 'PROFESSIONAL',
        name: 'Professional',
        price: '$199',
        period: '/mo',
        description: 'Advanced compliance features',
        features: ['Up to 50 users', '90-day retention', 'Governance suite', 'Dark pattern lab', 'Analytics'],
        accent: 'from-blue-500 to-blue-600',
        badge: 'Popular',
    },
    {
        id: 'ENTERPRISE',
        name: 'Enterprise',
        price: 'Custom',
        period: '',
        description: 'Full compliance suite with SLA',
        features: ['Unlimited users', '365-day retention', 'Full feature set', 'Priority support', 'Custom integrations'],
        accent: 'from-purple-500 to-purple-600',
        badge: null,
    },
] as const;

const STEPS = [
    { id: 1, label: 'Organization' },
    { id: 2, label: 'Select Plan' },
];

function StepIndicator({ currentStep }: { currentStep: number }) {
    return (
        <div className="flex items-center justify-center gap-2 mb-8">
            {STEPS.map((step, index) => (
                <div key={step.id} className="flex items-center gap-2">
                    <div className={cn(
                        "w-8 h-8 rounded-full flex items-center justify-center text-sm font-medium transition-all duration-300",
                        currentStep === step.id
                            ? "bg-blue-500 text-white shadow-md shadow-blue-500/25"
                            : currentStep > step.id
                                ? "bg-emerald-500 text-white"
                                : "bg-zinc-100 text-zinc-400 dark:bg-zinc-800 dark:text-zinc-500"
                    )}>
                        {currentStep > step.id ? <Check className="h-4 w-4" /> : step.id}
                    </div>
                    <span className={cn(
                        "text-sm font-medium transition-colors",
                        currentStep >= step.id
                            ? "text-zinc-900 dark:text-zinc-100"
                            : "text-zinc-400 dark:text-zinc-500"
                    )}>
                        {step.label}
                    </span>
                    {index < STEPS.length - 1 && (
                        <div className={cn(
                            "w-12 h-0.5 mx-2 rounded transition-colors",
                            currentStep > step.id ? "bg-emerald-500" : "bg-zinc-200 dark:bg-zinc-700"
                        )} />
                    )}
                </div>
            ))}
        </div>
    );
}

export function TenantForm({ isOpen, onClose, onSuccess }: TenantFormProps) {
    const [step, setStep] = useState(1);
    const [formData, setFormData] = useState<CreateTenantInput>({
        name: '',
        domain: '',
        admin_email: '',
        plan: 'FREE',
    });
    const [isSubmitting, setIsSubmitting] = useState(false);

    const handleClose = () => {
        onClose();
        // Reset after animation
        setTimeout(() => {
            setStep(1);
            setFormData({ name: '', domain: '', admin_email: '', plan: 'FREE' });
        }, 200);
    };

    const handleSubmit = async () => {
        setIsSubmitting(true);
        try {
            await adminService.createTenant(formData);
            toast.success('Success', `Tenant "${formData.name}" created successfully.`);
            onSuccess();
            handleClose();
        } catch (error) {
            console.error('Failed to create tenant', error);
        } finally {
            setIsSubmitting(false);
        }
    };

    const canProceedToStep2 = formData.name.trim() !== '' && formData.domain.trim() !== '' && formData.admin_email.trim() !== '';

    return (
        <Dialog open={isOpen} onOpenChange={(open) => { if (!open) handleClose(); }}>
            <DialogContent className="sm:max-w-[640px] p-0 overflow-hidden bg-white dark:bg-zinc-900 border-zinc-200 dark:border-zinc-800">
                {/* Header with gradient */}
                <div className="bg-gradient-to-r from-blue-500/5 via-purple-500/5 to-blue-500/5 dark:from-blue-500/10 dark:via-purple-500/10 dark:to-blue-500/10 px-8 pt-8 pb-6">
                    <DialogHeader>
                        <DialogTitle className="text-xl font-bold text-zinc-900 dark:text-zinc-50">
                            Onboard New Tenant
                        </DialogTitle>
                        <DialogDescription className="text-zinc-500 dark:text-zinc-400">
                            Set up a new organization workspace in two quick steps.
                        </DialogDescription>
                    </DialogHeader>
                    <div className="mt-6">
                        <StepIndicator currentStep={step} />
                    </div>
                </div>

                <div className="px-8 pb-8">
                    {/* Step 1: Organization Details */}
                    {step === 1 && (
                        <div className="space-y-5 animate-in fade-in slide-in-from-left-4 duration-300">
                            <div>
                                <label className="block text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-1.5">
                                    <Building2 className="h-4 w-4 inline mr-1.5 text-blue-500" />
                                    Organization Name
                                </label>
                                <input
                                    type="text"
                                    required
                                    className="w-full px-4 py-2.5 bg-zinc-50 dark:bg-zinc-800 border border-zinc-200 dark:border-zinc-700 rounded-lg focus:bg-white dark:focus:bg-zinc-800 focus:outline-none focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 dark:focus:border-blue-400 transition-all text-zinc-900 dark:text-zinc-100 placeholder:text-zinc-400 dark:placeholder:text-zinc-500"
                                    value={formData.name}
                                    onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                                    placeholder="Acme Corp"
                                />
                            </div>

                            <div>
                                <label className="block text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-1.5">
                                    Subdomain
                                </label>
                                <div className="flex">
                                    <input
                                        type="text"
                                        required
                                        pattern="[a-z0-9-]+"
                                        title="Lowercase letters, numbers, and hyphens only"
                                        className="flex-1 px-4 py-2.5 bg-zinc-50 dark:bg-zinc-800 border border-zinc-200 dark:border-zinc-700 rounded-l-lg focus:bg-white dark:focus:bg-zinc-800 focus:outline-none focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 dark:focus:border-blue-400 transition-all text-zinc-900 dark:text-zinc-100 placeholder:text-zinc-400 dark:placeholder:text-zinc-500"
                                        value={formData.domain}
                                        onChange={(e) => setFormData({ ...formData, domain: e.target.value.toLowerCase() })}
                                        placeholder="acme"
                                    />
                                    <span className="inline-flex items-center px-4 rounded-r-lg border border-l-0 border-zinc-200 dark:border-zinc-700 bg-zinc-100 dark:bg-zinc-800 text-zinc-500 dark:text-zinc-400 text-sm font-medium">
                                        .datalens.com
                                    </span>
                                </div>
                            </div>

                            <div>
                                <label className="block text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-1.5">
                                    <Mail className="h-4 w-4 inline mr-1.5 text-blue-500" />
                                    Admin Email
                                </label>
                                <input
                                    type="email"
                                    required
                                    className="w-full px-4 py-2.5 bg-zinc-50 dark:bg-zinc-800 border border-zinc-200 dark:border-zinc-700 rounded-lg focus:bg-white dark:focus:bg-zinc-800 focus:outline-none focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 dark:focus:border-blue-400 transition-all text-zinc-900 dark:text-zinc-100 placeholder:text-zinc-400 dark:placeholder:text-zinc-500"
                                    value={formData.admin_email}
                                    onChange={(e) => setFormData({ ...formData, admin_email: e.target.value })}
                                    placeholder="admin@acme.com"
                                />
                            </div>

                            <div className="pt-4 flex justify-end">
                                <Button
                                    onClick={() => setStep(2)}
                                    disabled={!canProceedToStep2}
                                    className="min-w-[140px]"
                                >
                                    Continue <ArrowRight className="h-4 w-4 ml-2" />
                                </Button>
                            </div>
                        </div>
                    )}

                    {/* Step 2: Plan Selection */}
                    {step === 2 && (
                        <div className="space-y-4 animate-in fade-in slide-in-from-right-4 duration-300">
                            <div className="flex items-center gap-2 mb-2">
                                <CreditCard className="h-5 w-5 text-purple-500" />
                                <h3 className="text-base font-semibold text-zinc-900 dark:text-zinc-50">
                                    Choose a Plan
                                </h3>
                            </div>

                            <div className="grid grid-cols-2 gap-3">
                                {PLANS.map((plan) => (
                                    <div
                                        key={plan.id}
                                        onClick={() => setFormData({ ...formData, plan: plan.id as CreateTenantInput['plan'] })}
                                        className={cn(
                                            "relative flex flex-col p-4 border rounded-xl cursor-pointer transition-all duration-200",
                                            formData.plan === plan.id
                                                ? "border-blue-500 bg-blue-50/50 dark:bg-blue-500/10 ring-1 ring-blue-500 dark:border-blue-400"
                                                : "border-zinc-200 dark:border-zinc-700 bg-white dark:bg-zinc-800/50 hover:border-zinc-300 dark:hover:border-zinc-600"
                                        )}
                                    >
                                        {plan.badge && (
                                            <span className="absolute -top-2.5 left-3 px-2 py-0.5 bg-blue-500 text-white text-[10px] font-bold rounded-full flex items-center gap-1">
                                                <Sparkles className="h-2.5 w-2.5" />
                                                {plan.badge}
                                            </span>
                                        )}
                                        <div className="flex items-center justify-between mb-2">
                                            <span className="font-semibold text-sm text-zinc-900 dark:text-zinc-100">
                                                {plan.name}
                                            </span>
                                            {formData.plan === plan.id && (
                                                <div className="w-5 h-5 rounded-full bg-blue-500 flex items-center justify-center">
                                                    <Check className="h-3 w-3 text-white" />
                                                </div>
                                            )}
                                        </div>
                                        <div className="mb-2">
                                            <span className="text-lg font-bold text-zinc-900 dark:text-zinc-50">{plan.price}</span>
                                            <span className="text-xs text-zinc-500 dark:text-zinc-400">{plan.period}</span>
                                        </div>
                                        <p className="text-xs text-zinc-500 dark:text-zinc-400 mb-3">{plan.description}</p>
                                        <ul className="space-y-1">
                                            {plan.features.slice(0, 3).map((feature) => (
                                                <li key={feature} className="text-xs text-zinc-600 dark:text-zinc-300 flex items-center gap-1.5">
                                                    <Check className="h-3 w-3 text-emerald-500 shrink-0" />
                                                    {feature}
                                                </li>
                                            ))}
                                        </ul>
                                    </div>
                                ))}
                            </div>

                            <div className="pt-4 border-t border-zinc-100 dark:border-zinc-800 flex justify-between">
                                <Button variant="ghost" onClick={() => setStep(1)}>
                                    <ArrowLeft className="h-4 w-4 mr-2" /> Back
                                </Button>
                                <Button onClick={handleSubmit} isLoading={isSubmitting} className="min-w-[160px]">
                                    Create Tenant
                                </Button>
                            </div>
                        </div>
                    )}
                </div>
            </DialogContent>
        </Dialog>
    );
}
