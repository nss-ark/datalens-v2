import { useState, useEffect } from 'react';
import { Save, HardDrive, Mail, Users } from 'lucide-react';
import { Button } from '../common/Button';
import { toast } from '../../stores/toastStore';
import { dataSourceService } from '../../services/datasource';
import type { DataSource, GoogleScopeConfig as GoogleConfigType } from '../../types/datasource';

interface GoogleScopeConfigProps {
    dataSource: DataSource;
    onSave?: () => void;
}

const DEFAULT_CONFIG: GoogleConfigType = {
    scanMyDrive: true,
    scanSharedDrives: true,
    scanGmail: false
};

export const GoogleScopeConfig = ({ dataSource, onSave }: GoogleScopeConfigProps) => {
    const [config, setConfig] = useState<GoogleConfigType>(DEFAULT_CONFIG);
    const [isSaving, setIsSaving] = useState(false);

    useEffect(() => {
        if (dataSource.config && typeof dataSource.config === 'string') {
            try {
                const parsed = JSON.parse(dataSource.config);
                setConfig({ ...DEFAULT_CONFIG, ...parsed });
            } catch (e) {
                console.error('Failed to parse Google config', e);
            }
        }
    }, [dataSource.config]);

    const handleSave = async () => {
        setIsSaving(true);
        try {
            await dataSourceService.updateScope(dataSource.id, config);
            toast.success('Configuration Saved', 'Google Workspace scan scope updated.');
            if (onSave) onSave();
        } catch (error) {
            console.error('Failed to save config', error);
            toast.error('Save Failed', 'Could not update configuration.');
        } finally {
            setIsSaving(false);
        }
    };

    const toggle = (field: keyof GoogleConfigType) => {
        setConfig(prev => ({ ...prev, [field]: !prev[field] }));
    };

    return (
        <div className="bg-white border border-gray-200 rounded-xl shadow-sm overflow-hidden max-w-2xl mx-auto mt-8">
            <div className="px-6 py-4 border-b border-gray-100 bg-gray-50 flex justify-between items-center">
                <div>
                    <h3 className="text-lg font-semibold text-gray-900">Scan Scope</h3>
                    <p className="text-sm text-gray-500">Select which Google Workspace services to scan</p>
                </div>
                <Button
                    onClick={handleSave}
                    isLoading={isSaving}
                    icon={<Save size={16} />}
                >
                    Save Changes
                </Button>
            </div>

            <div className="p-6 space-y-6">
                <div className="flex items-center justify-between p-4 border border-gray-100 rounded-lg hover:bg-gray-50 transition-colors">
                    <div className="flex items-center gap-4">
                        <div className="h-10 w-10 rounded-full bg-blue-50 flex items-center justify-center text-blue-600">
                            <HardDrive size={20} />
                        </div>
                        <div>
                            <div className="font-medium text-gray-900">My Drive</div>
                            <div className="text-sm text-gray-500">Scan personal drives of all users</div>
                        </div>
                    </div>
                    <Toggle checked={config.scanMyDrive} onChange={() => toggle('scanMyDrive')} />
                </div>

                <div className="flex items-center justify-between p-4 border border-gray-100 rounded-lg hover:bg-gray-50 transition-colors">
                    <div className="flex items-center gap-4">
                        <div className="h-10 w-10 rounded-full bg-indigo-50 flex items-center justify-center text-indigo-600">
                            <Users size={20} />
                        </div>
                        <div>
                            <div className="font-medium text-gray-900">Shared Drives</div>
                            <div className="text-sm text-gray-500">Scan team and shared drives</div>
                        </div>
                    </div>
                    <Toggle checked={config.scanSharedDrives} onChange={() => toggle('scanSharedDrives')} />
                </div>

                <div className="flex items-center justify-between p-4 border border-gray-100 rounded-lg hover:bg-gray-50 transition-colors">
                    <div className="flex items-center gap-4">
                        <div className="h-10 w-10 rounded-full bg-red-50 flex items-center justify-center text-red-600">
                            <Mail size={20} />
                        </div>
                        <div>
                            <div className="font-medium text-gray-900">Gmail</div>
                            <div className="text-sm text-gray-500">Scan user email content</div>
                        </div>
                    </div>
                    <Toggle checked={config.scanGmail} onChange={() => toggle('scanGmail')} />
                </div>
            </div>
        </div>
    );
};

const Toggle = ({ checked, onChange }: { checked: boolean; onChange: () => void }) => (
    <button
        onClick={onChange}
        className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 ${checked ? 'bg-blue-600' : 'bg-gray-200'
            }`}
    >
        <span
            className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${checked ? 'translate-x-6' : 'translate-x-1'
                }`}
        />
    </button>
);
