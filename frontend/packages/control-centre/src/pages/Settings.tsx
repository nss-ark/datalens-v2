import { Bell, Lock } from 'lucide-react';
import { useAuthStore, Button, Headline01, Card09, Card08 } from '@datalens/shared';

const Settings = () => {
    const { user } = useAuthStore();

    return (
        <div className="space-y-8 max-w-4xl mx-auto">
            <div>
                <Headline01 title="Settings" subtitle="Manage your account settings and preferences" />
            </div>

            {/* Profile Section */}
            <Card09
                name={user?.name || 'User'}
                role={user?.role_ids?.[0] === 'ae5a2f18-4153-4bc3-8f18-2f47303d207c' || user?.email?.includes('admin') ? 'System Admin' : 'Member'}
                stats={[
                    { label: 'Role', value: user?.email?.includes('admin') ? 'System Admin' : 'Member' },
                    { label: 'Status', value: 'Active' },
                ]}
                actions={[
                    { label: 'Edit Profile', onClick: () => { }, variant: 'outline' },
                ]}
                className="w-full"
            />

            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                {/* Preferences Section */}
                <Card08 title="Preferences">
                    <div className="space-y-4">
                        <div className="flex items-center gap-2 text-gray-700">
                            <Bell className="h-5 w-5 text-gray-500" />
                            <span className="font-medium">Notifications</span>
                        </div>
                        <p className="text-sm text-gray-500">
                            Manage your notification and theme preferences.
                            <br />
                            <span className="italic text-xs">(Coming soon)</span>
                        </p>
                    </div>
                </Card08>

                {/* Security Section */}
                <Card08 title="Security">
                    <div className="space-y-4">
                        <div className="flex items-center gap-2 text-gray-700">
                            <Lock className="h-5 w-5 text-gray-500" />
                            <span className="font-medium">Password & Auth</span>
                        </div>
                        <div className="flex items-center justify-between p-3 bg-gray-50 rounded-lg border border-gray-200">
                            <div>
                                <h4 className="text-sm font-medium text-gray-900">Password</h4>
                                <p className="text-xs text-gray-500">Last changed 30 days ago</p>
                            </div>
                            <Button variant="outline" size="sm" disabled>Change</Button>
                        </div>
                    </div>
                </Card08>
            </div>

            <Card08 title="System Information">
                <div className="space-y-2 text-sm text-gray-600">
                    <div className="flex justify-between border-b border-gray-100 pb-2">
                        <span>User ID</span>
                        <code className="bg-gray-100 px-2 py-0.5 rounded text-xs">{user?.id}</code>
                    </div>
                    <div className="flex justify-between pt-2">
                        <span>Email</span>
                        <span className="font-medium">{user?.email}</span>
                    </div>
                </div>
            </Card08>
        </div>
    );
};

export default Settings;
