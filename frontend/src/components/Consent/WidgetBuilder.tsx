import { useState } from 'react';
import { Check, ChevronRight, ChevronLeft, Palette, Globe, Shield, Code } from 'lucide-react';
import { Button } from '../common/Button';
// import { Select } from '../common/Select'; // Assuming Select exists, if not use native select
import { useCreateWidget } from '../../hooks/useConsent';
import type { CreateWidgetInput } from '../../types/consent';

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
            // TODO: Toast error
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
            <div className="flex justify-between items-center mb-8 px-4">
                {STEPS.map((s, i) => (
                    <div key={i} className={`flex items-center ${i <= step ? 'text-blue-600' : 'text-gray-400'}`}>
                        <div className={`w-8 h-8 rounded-full flex items-center justify-center border-2 ${i <= step ? 'border-blue-600 bg-blue-50' : 'border-gray-200'}`}>
                            {i < step ? <Check size={16} /> : <s.icon size={16} />}
                        </div>
                        <span className="ml-2 text-sm font-medium hidden sm:block">{s.title}</span>
                        {i < STEPS.length - 1 && <div className="w-8 h-[2px] mx-2 bg-gray-200" />}
                    </div>
                ))}
            </div>

            {/* Step Content */}
            <div className="flex-1 overflow-y-auto px-4 py-2">
                {step === 0 && (
                    <div className="space-y-4">
                        <div>
                            <label className="block text-sm font-medium text-gray-700">Widget Name</label>
                            <input
                                type="text"
                                className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm border p-2"
                                value={formData.name}
                                onChange={(e) => updateField('name', e.target.value)}
                                placeholder="e.g. Main Website Banner"
                            />
                        </div>
                        <div>
                            <label className="block text-sm font-medium text-gray-700">Domain</label>
                            <input
                                type="text"
                                className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm border p-2"
                                value={formData.domain}
                                onChange={(e) => updateField('domain', e.target.value)}
                                placeholder="e.g. example.com"
                            />
                        </div>
                        <div>
                            <label className="block text-sm font-medium text-gray-700">Type</label>
                            <select
                                className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm border p-2"
                                value={formData.type}
                                onChange={(e) => updateField('type', e.target.value as CreateWidgetInput['type'])}
                            >
                                <option value="BANNER">Cookie Banner</option>
                                <option value="PREFERENCE_CENTER">Preference Center</option>
                                <option value="INLINE_FORM">Inline Form</option>
                                <option value="PORTAL">Full Portal (Iframe)</option>
                            </select>
                        </div>
                    </div>
                )}

                {step === 1 && (
                    <div className="space-y-4">
                        <div>
                            <label className="block text-sm font-medium text-gray-700">Layout Position</label>
                            <div className="grid grid-cols-2 gap-4 mt-2">
                                {['BOTTOM_BAR', 'TOP_BAR', 'MODAL', 'SIDEBAR'].map((layout) => (
                                    <div
                                        key={layout}
                                        className={`border rounded-lg p-4 cursor-pointer hover:border-blue-500 ${formData.config.layout === layout ? 'border-blue-500 bg-blue-50' : 'border-gray-200'}`}
                                        onClick={() => updateConfig('layout', layout as CreateWidgetInput['config']['layout'])}
                                    >
                                        <div className="text-sm font-medium text-center">{layout.replace('_', ' ')}</div>
                                    </div>
                                ))}
                            </div>
                        </div>
                        <div className="grid grid-cols-2 gap-4">
                            <div>
                                <label className="block text-sm font-medium text-gray-700">Primary Color</label>
                                <div className="flex mt-1">
                                    <input
                                        type="color"
                                        className="h-9 w-9 rounded-l border border-gray-300"
                                        value={formData.config.theme.primary_color}
                                        onChange={(e) => updateTheme('primary_color', e.target.value)}
                                    />
                                    <input
                                        type="text"
                                        className="block w-full rounded-r border-gray-300 border-t border-b border-r sm:text-sm p-2"
                                        value={formData.config.theme.primary_color}
                                        onChange={(e) => updateTheme('primary_color', e.target.value)}
                                    />
                                </div>
                            </div>
                            <div>
                                <label className="block text-sm font-medium text-gray-700">Text Color</label>
                                <div className="flex mt-1">
                                    <input
                                        type="color"
                                        className="h-9 w-9 rounded-l border border-gray-300"
                                        value={formData.config.theme.text_color}
                                        onChange={(e) => updateTheme('text_color', e.target.value)}
                                    />
                                    <input
                                        type="text"
                                        className="block w-full rounded-r border-gray-300 border-t border-b border-r sm:text-sm p-2"
                                        value={formData.config.theme.text_color}
                                        onChange={(e) => updateTheme('text_color', e.target.value)}
                                    />
                                </div>
                            </div>
                        </div>
                    </div>
                )}

                {step === 2 && (
                    <div className="space-y-4">
                        <div className="bg-yellow-50 p-4 rounded-md border border-yellow-200">
                            <h4 className="text-sm font-medium text-yellow-800">DPDPA Compliance Note</h4>
                            <p className="text-sm text-yellow-700 mt-1">
                                Under DPDPA 2023, consent must be explicit. "Opt-out" defaults are generally non-compliant for sensitive data.
                                Ensure your purposes are clearly defined.
                            </p>
                        </div>

                        <div>
                            <label className="block text-sm font-medium text-gray-700">Default State</label>
                            <select
                                className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm border p-2"
                                value={formData.config.default_state}
                                onChange={(e) => updateConfig('default_state', e.target.value as CreateWidgetInput['config']['default_state'])}
                            >
                                <option value="OPT_OUT">Opt-Out (Data is collected unless user declines)</option>
                                <option value="OPT_IN">Opt-In (Strict - No data until explicit consent)</option>
                            </select>
                        </div>

                        <div className="flex items-center space-x-2 mt-4">
                            <input
                                type="checkbox"
                                id="granular"
                                className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
                                checked={formData.config.granular_toggle}
                                onChange={(e) => updateConfig('granular_toggle', e.target.checked)}
                            />
                            <label htmlFor="granular" className="text-sm text-gray-700">Enable Granular Purpose Toggles</label>
                        </div>

                        <div className="flex items-center space-x-2">
                            <input
                                type="checkbox"
                                id="block"
                                className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
                                checked={formData.config.block_until_consent}
                                onChange={(e) => updateConfig('block_until_consent', e.target.checked)}
                            />
                            <label htmlFor="block" className="text-sm text-gray-700">Block Site Until Consent is Given</label>
                        </div>
                    </div>
                )}

                {step === 3 && (
                    <div className="space-y-4">
                        <h3 className="text-lg font-medium text-gray-900">Review Configuration</h3>

                        <div className="bg-gray-50 p-4 rounded-md border border-gray-200 max-h-60 overflow-y-auto">
                            <pre className="text-xs text-gray-600 font-mono">
                                {JSON.stringify(formData, null, 2)}
                            </pre>
                        </div>

                        <div className="flex items-center p-4 bg-blue-50 text-blue-700 rounded-md">
                            <Code size={20} className="mr-3" />
                            <div className="text-sm">
                                Embed code will be generated after creation.
                            </div>
                        </div>
                    </div>
                )}
            </div>

            {/* Footer */}
            <div className="flex justify-between mt-auto border-t pt-4 px-4 sticky bottom-0 bg-white">
                <Button onClick={onClose} variant="ghost">Cancel</Button>
                <div className="flex space-x-2">
                    {step > 0 && (
                        <Button onClick={handleBack} variant="secondary">
                            <ChevronLeft size={16} className="mr-1" /> Back
                        </Button>
                    )}
                    {step < STEPS.length - 1 ? (
                        <Button onClick={handleNext}>
                            Next <ChevronRight size={16} className="ml-1" />
                        </Button>
                    ) : (
                        <Button onClick={handleSubmit} isLoading={isPending}>
                            Create Widget
                        </Button>
                    )}
                </div>
            </div>
        </div>
    );
};
export default WidgetBuilder;
