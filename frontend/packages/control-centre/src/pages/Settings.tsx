import { User, Mail, Shield, Bell, Lock } from 'lucide-react';
import { useAuthStore } from '@datalens/shared';
import { Card, CardHeader, CardContent, CardTitle, CardDescription } from '@datalens/shared';
import { Button, Input } from '@datalens/shared';

const Settings = () => {
    const { user } = useAuthStore();

    return (
        <div className="space-y-6 max-w-4xl mx-auto">
            <div>
                <h1 className="text-2xl font-bold text-gray-900 mb-1">Settings</h1>
                <p className="text-sm text-gray-500">
                    Manage your account settings and preferences
                </p>
            </div>

            {/* Profile Section */}
            <Card>
                <CardHeader>
                    <div className="flex items-center gap-2">
                        <User className="h-5 w-5 text-gray-500" />
                        <CardTitle>Profile Information</CardTitle>
                    </div>
                    <CardDescription>
                        Your account details and role information
                    </CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                        <div className="space-y-2">
                            <label className="text-sm font-medium text-gray-700">Full Name</label>
                            <div className="relative">
                                <User className="absolute left-3 top-2.5 h-4 w-4 text-gray-400" />
                                <Input
                                    value={user?.name || ''}
                                    readOnly
                                    className="pl-9 bg-gray-50 text-gray-600 cursor-not-allowed"
                                />
                            </div>
                        </div>
                        <div className="space-y-2">
                            <label className="text-sm font-medium text-gray-700">Email Address</label>
                            <div className="relative">
                                <Mail className="absolute left-3 top-2.5 h-4 w-4 text-gray-400" />
                                <Input
                                    value={user?.email || ''}
                                    readOnly
                                    className="pl-9 bg-gray-50 text-gray-600 cursor-not-allowed"
                                />
                            </div>
                        </div>
                    </div>

                    <div className="pt-4 border-t border-gray-100">
                        <label className="text-sm font-medium text-gray-700 mb-2 block">Roles & Permissions</label>
                        <div className="flex flex-wrap gap-2">
                            {user?.role_ids?.map((role) => (
                                <span key={role} className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800">
                                    <Shield className="w-3 h-3 mr-1" />
                                    {role}
                                </span>
                            )) || <span className="text-sm text-gray-500">No roles assigned</span>}
                        </div>
                    </div>
                </CardContent>
            </Card>

            {/* Preferences Section (Placeholder) */}
            <Card>
                <CardHeader>
                    <div className="flex items-center gap-2">
                        <Bell className="h-5 w-5 text-gray-500" />
                        <CardTitle>Preferences</CardTitle>
                    </div>
                    <CardDescription>
                        Manage your notification and theme preferences
                    </CardDescription>
                </CardHeader>
                <CardContent>
                    <div className="text-sm text-gray-500 italic py-4 text-center border border-dashed border-gray-200 rounded-md">
                        Preference settings are coming soon.
                    </div>
                </CardContent>
            </Card>

            {/* Security Section (Placeholder) */}
            <Card>
                <CardHeader>
                    <div className="flex items-center gap-2">
                        <Lock className="h-5 w-5 text-gray-500" />
                        <CardTitle>Security</CardTitle>
                    </div>
                    <CardDescription>
                        Password and security settings
                    </CardDescription>
                </CardHeader>
                <CardContent>
                    <div className="flex items-center justify-between p-4 bg-gray-50 rounded-lg border border-gray-200">
                        <div>
                            <h4 className="text-sm font-medium text-gray-900">Change Password</h4>
                            <p className="text-xs text-gray-500 mt-1">Update your account password</p>
                        </div>
                        <Button variant="outline" size="sm" disabled>Change</Button>
                    </div>
                </CardContent>
            </Card>
        </div>
    );
};

export default Settings;
