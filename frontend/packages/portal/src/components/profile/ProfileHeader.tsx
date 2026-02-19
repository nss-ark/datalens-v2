import { Settings } from 'lucide-react';

interface ProfileHeaderProps {
    verificationStatus: string;
}

export const ProfileHeader = ({ verificationStatus }: ProfileHeaderProps) => {
    const isVerified = verificationStatus === 'VERIFIED';

    return (
        <div className="flex flex-col md:flex-row md:items-end justify-between mb-12">
            <div>
                <div className="flex items-center gap-3">
                    <h1 className="text-3xl font-bold text-gray-900 dark:text-white tracking-tight">My Profile</h1>
                    {isVerified ? (
                        <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-bold bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300 tracking-wide uppercase border border-green-200 dark:border-green-800">
                            <span className="w-1.5 h-1.5 rounded-full bg-green-600 dark:bg-green-400 mr-1.5"></span>
                            Verified
                        </span>
                    ) : (
                        <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-bold bg-slate-100 text-slate-500 dark:bg-slate-800 dark:text-slate-400 tracking-wide uppercase border border-slate-200 dark:border-slate-700">
                            Unverified
                        </span>
                    )}
                </div>
                <p className="mt-2 text-gray-500 dark:text-gray-400 text-lg max-w-2xl">
                    Manage your personal information, security settings, and verification status.
                </p>
            </div>
            <div className="mt-4 md:mt-0">
                <button className="text-sm text-gray-500 hover:text-gray-900 dark:text-gray-400 dark:hover:text-white font-medium flex items-center gap-1 transition-colors">
                    <Settings className="w-4 h-4" /> Settings
                </button>
            </div>
        </div>
    );
};
