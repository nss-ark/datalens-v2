import { useState } from 'react';
import { Check, ChevronRight, ChevronLeft, Palette, Globe, Shield, Code } from 'lucide-react';
import { Button } from '@datalens/shared';
import { Input } from '@datalens/shared';
import { Label } from '@datalens/shared';
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from '@datalens/shared';
import { useCreateWidget } from '../../hooks/useConsent';
import type { CreateWidgetInput } from '../../types/consent';
import { cn } from '@datalens/shared';

interface WidgetBuilderProps {
    onClose: () => void;
}

const STEPS = [
    { title: 'Basics', icon: Shield },
    { title: 'Appearance', icon: Palette },
    { title: 'Configuration', icon: Globe },
    { title: 'Review', icon: Code },
];

export const WidgetBuilder = ({ onClose }: WidgetBuilderProps) => {
    const [step, setStep] = useState(0);
    const { mutateAsync: createWidget, isPending } = useCreateWidget();

    // Form State (Single big object for simplicity across steps)
    const [formData, setFormData] = useState<CreateWidgetInput>({
        name: '',
        type: 'BANNER',
        domain: '',
        allowed_origins: [],
        config: {
            theme: {
                primary_color: '#3B82F6',
                background_color: '#FFFFFF',
                text_color: '#1F2937',
                font_family: 'Inter, sans-serif',
                border_radius: '8px',
            },
            layout: 'BOTTOM_BAR',
            purpose_ids: [], // TODO: Fetch from API
            default_state: 'OPT_OUT',
            show_categories: true,
            granular_toggle: true,
            block_until_consent: false,
            languages: ['en'],
            default_language: 'en',
            translations: {
                en: {
                    title: 'We value your privacy',
                    description: 'We use cookies to improve your experience.',
                    accept_all: 'Accept All',
                    reject_all: 'Reject All',
                    customize: 'Customize',
                },
            },
            regulation_ref: 'DPDPA',
            require_explicit: true,
            consent_expiry_days: 365,
        },
    });

    const handleNext = () => setStep((prev) => Math.min(prev + 1, STEPS.length - 1));
    const handleBack = () => setStep((prev) => Math.max(prev - 1, 0));

    const handleSubmit = async () => {
        try {
            await createWidget(formData);
            onClose();
        } catch (error) {
            console.error('Failed to create widget', error);
        }
    };

    const updateField = <K extends keyof CreateWidgetInput>(field: K, value: CreateWidgetInput[K]) => {
        setFormData(prev => ({ ...prev, [field]: value }));
    };

    const updateConfig = <K extends keyof CreateWidgetInput['config']>(field: K, value: CreateWidgetInput['config'][K]) => {
        setFormData(prev => ({
            ...prev,
            config: { ...prev.config, [field]: value }
        }));
    };

    const updateTheme = <K extends keyof CreateWidgetInput['config']['theme']>(field: K, value: CreateWidgetInput['config']['theme'][K]) => {
        setFormData(prev => ({
            ...prev,
            config: {
                ...prev.config,
                theme: { ...prev.config.theme, [field]: value }
            }
        }));
    };

    return (
        <div className="flex flex-col h-full min-h-[500px]">
            {/* Stepper Header */}
            <div className="flex justify-between items-center px-8 py-6 bg-gray-50/50 border-b">
                {STEPS.map((s, i) => (
                    <div key={i} className={cn("flex items-center", i <= step ? 'text-primary' : 'text-muted-foreground')}>
                        <div className={cn(
                            "w-8 h-8 rounded-full flex items-center justify-center border-2 transition-colors",
                            i < step ? "bg-primary border-primary text-primary-foreground" :
                                i === step ? "border-primary text-primary" : "border-border"
                        )}>
                            {i < step ? <Check size={16} /> : <s.icon size={16} />}
                        </div>
                        <span className="ml-3 text-sm font-medium hidden sm:block">{s.title}</span>
                        {i < STEPS.length - 1 && <div className="w-12 h-[2px] mx-4 bg-border" />}
                    </div>
                ))}
            </div>

            {/* Step Content */}
            <div className="flex-1 overflow-y-auto px-8 py-6">
                {step === 0 && (
                    <div className="space-y-6 max-w-lg mx-auto">
                        <div className="space-y-2">
                            <Label>Widget Name</Label>
                            <Input
                                value={formData.name}
                                onChange={(e) => updateField('name', e.target.value)}
                                placeholder="e.g. Main Website Banner"
                            />
                        </div>
                        <div className="space-y-2">
                            <Label>Domain</Label>
                            <Input
                                value={formData.domain}
                                onChange={(e) => updateField('domain', e.target.value)}
                                placeholder="e.g. example.com"
                            />
                        </div>
                        <div className="space-y-2">
                            <Label>Type</Label>
                            <Select
                                value={formData.type}
                                onValueChange={(val) => updateField('type', val as CreateWidgetInput['type'])}
                            >
                                <SelectTrigger>
                                    <SelectValue placeholder="Select type" />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="BANNER">Cookie Banner</SelectItem>
                                    <SelectItem value="PREFERENCE_CENTER">Preference Center</SelectItem>
                                    <SelectItem value="INLINE_FORM">Inline Form</SelectItem>
                                    <SelectItem value="PORTAL">Full Portal (Iframe)</SelectItem>
                                </SelectContent>
                            </Select>
                        </div>
                    </div>
                )}

                {step === 1 && (
                    <div className="space-y-6 max-w-2xl mx-auto">
                        <div className="space-y-3">
                            <Label>Layout Position</Label>
                            <div className="grid grid-cols-2 gap-4">
                                {['BOTTOM_BAR', 'TOP_BAR', 'MODAL', 'SIDEBAR'].map((layout) => (
                                    <div
                                        key={layout}
                                        className={cn(
                                            "border rounded-lg p-4 cursor-pointer hover:border-primary transition-all",
                                            formData.config.layout === layout ? 'border-primary bg-primary/5 ring-1 ring-primary' : 'border-border'
                                        )}
                                        onClick={() => updateConfig('layout', layout as CreateWidgetInput['config']['layout'])}
                                    >
                                        <div className="text-sm font-medium text-center">{layout.replace('_', ' ')}</div>
                                    </div>
                                ))}
                            </div>
                        </div>
                        <div className="grid grid-cols-2 gap-6">
                            <div className="space-y-2">
                                <Label>Primary Color</Label>
                                <div className="flex gap-2">
                                    <div className="relative">
                                        <Input
                                            type="color"
                                            className="h-10 w-12 p-1 cursor-pointer"
                                            value={formData.config.theme.primary_color}
                                            onChange={(e) => updateTheme('primary_color', e.target.value)}
                                        />
                                    </div>
                                    <Input
                                        type="text"
                                        value={formData.config.theme.primary_color}
                                        onChange={(e) => updateTheme('primary_color', e.target.value)}
                                        className="font-mono uppercase"
                                    />
                                </div>
                            </div>
                            <div className="space-y-2">
                                <Label>Text Color</Label>
                                <div className="flex gap-2">
                                    <div className="relative">
                                        <Input
                                            type="color"
                                            className="h-10 w-12 p-1 cursor-pointer"
                                            value={formData.config.theme.text_color}
                                            onChange={(e) => updateTheme('text_color', e.target.value)}
                                        />
                                    </div>
                                    <Input
                                        type="text"
                                        value={formData.config.theme.text_color}
                                        onChange={(e) => updateTheme('text_color', e.target.value)}
                                        className="font-mono uppercase"
                                    />
                                </div>
                            </div>
                        </div>
                    </div>
                )}

                {step === 2 && (
                    <div className="space-y-6 max-w-lg mx-auto">
                        <div className="bg-amber-50 p-4 rounded-md border border-amber-200">
                            <h4 className="text-sm font-medium text-amber-800">DPDPA Compliance Note</h4>
                            <p className="text-sm text-amber-700 mt-1">
                                Under DPDPA 2023, consent must be explicit. "Opt-out" defaults are generally non-compliant for sensitive data.
                                Ensure your purposes are clearly defined.
                            </p>
                        </div>

                        <div className="space-y-2">
                            <Label>Default State</Label>
                            <Select
                                value={formData.config.default_state}
                                onValueChange={(val) => updateConfig('default_state', val as CreateWidgetInput['config']['default_state'])}
                            >
                                <SelectTrigger>
                                    <SelectValue placeholder="Select default state" />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="OPT_OUT">Opt-Out (Data is collected unless user declines)</SelectItem>
                                    <SelectItem value="OPT_IN">Opt-In (Strict - No data until explicit consent)</SelectItem>
                                </SelectContent>
                            </Select>
                        </div>

                        <div className="flex items-center space-x-3 p-3 border rounded-lg hover:bg-gray-50 transition-colors">
                            <input
                                type="checkbox"
                                id="granular"
                                className="h-4 w-4 text-primary focus:ring-primary border-gray-300 rounded"
                                checked={formData.config.granular_toggle}
                                onChange={(e) => updateConfig('granular_toggle', e.target.checked)}
                            />
                            <Label htmlFor="granular" className="cursor-pointer font-normal">Enable Granular Purpose Toggles</Label>
                        </div>

                        <div className="flex items-center space-x-3 p-3 border rounded-lg hover:bg-gray-50 transition-colors">
                            <input
                                type="checkbox"
                                id="block"
                                className="h-4 w-4 text-primary focus:ring-primary border-gray-300 rounded"
                                checked={formData.config.block_until_consent}
                                onChange={(e) => updateConfig('block_until_consent', e.target.checked)}
                            />
                            <Label htmlFor="block" className="cursor-pointer font-normal">Block Site Until Consent is Given</Label>
                        </div>
                    </div>
                )}

                {step === 3 && (
                    <div className="space-y-6 max-w-2xl mx-auto">
                        <h3 className="text-lg font-medium text-foreground">Review Configuration</h3>

                        <div className="bg-slate-50 p-4 rounded-md border border-slate-200 max-h-[300px] overflow-y-auto">
                            <pre className="text-xs text-slate-700 font-mono">
                                {JSON.stringify(formData, null, 2)}
                            </pre>
                        </div>

                        <div className="flex items-center p-4 bg-blue-50 text-blue-700 rounded-md border border-blue-100">
                            <Code size={20} className="mr-3" />
                            <div className="text-sm">
                                Embed code will be generated after creation.
                            </div>
                        </div>
                    </div>
                )}
            </div>

            {/* Footer */}
            <div className="flex justify-between items-center py-4 px-8 border-t bg-gray-50/50">
                <Button onClick={onClose} variant="ghost">Cancel</Button>
                <div className="flex space-x-3">
                    {step > 0 && (
                        <Button onClick={handleBack} variant="outline">
                            <ChevronLeft size={16} className="mr-1" /> Back
                        </Button>
                    )}
                    {step < STEPS.length - 1 ? (
                        <Button onClick={handleNext}>
                            Next <ChevronRight size={16} className="ml-1" />
                        </Button>
                    ) : (
                        <Button onClick={handleSubmit} disabled={isPending}>
                            {isPending && <div className="mr-2 h-4 w-4 animate-spin rounded-full border-2 border-current border-t-transparent" />}
                            Create Widget
                        </Button>
                    )}
                </div>
            </div>
        </div>
    );
};
export default WidgetBuilder;
