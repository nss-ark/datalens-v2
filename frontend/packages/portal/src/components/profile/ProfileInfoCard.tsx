import { MoreHorizontal, Edit2 } from 'lucide-react';
import type { PortalProfile } from '@/types/portal';
import { format } from 'date-fns';

interface ProfileInfoCardProps {
    profile: PortalProfile;
}

export const ProfileInfoCard = ({ profile }: ProfileInfoCardProps) => {
    const initials = profile.email?.substring(0, 1).toUpperCase() || 'U';
    const name = profile.email?.split('@')[0] || 'User';

    return (
        <div className="bg-white dark:bg-slate-900 rounded-2xl p-10 border border-gray-200/60 dark:border-gray-800 shadow-sm hover:shadow-md transition-all duration-300 flex flex-col justify-between relative group h-full">
            <div className="absolute top-6 right-6 text-gray-400 hover:text-gray-600 cursor-pointer">
                <MoreHorizontal size={20} />
            </div>

            <div className="mb-10 text-center md:text-left">
                <div className="w-32 h-32 rounded-full bg-slate-100 dark:bg-slate-800 flex items-center justify-center text-slate-500 dark:text-slate-400 text-5xl font-light mb-8 shadow-inner mx-auto md:mx-0">
                    {initials}
                </div>
                <h2 className="text-3xl font-bold text-gray-900 dark:text-white mb-2">{name}</h2>
                <p className="text-gray-500 dark:text-gray-400 font-medium text-lg">Data Principal</p>
            </div>

            <div className="space-y-10">
                <div className="group/field">
                    <label className="block text-xs font-semibold uppercase tracking-wider text-gray-400 mb-1">Email Address</label>
                    <div className="flex items-center justify-between py-2 border-b border-gray-100 dark:border-gray-800 group-hover/field:border-gray-300 dark:group-hover/field:border-gray-700 transition-colors">
                        <span className="text-gray-900 dark:text-white font-medium truncate max-w-[200px]" title={profile.email}>{profile.email}</span>
                        <Edit2 size={14} className="text-gray-300 opacity-0 group-hover/field:opacity-100 transition-opacity cursor-pointer" />
                    </div>
                </div>

                <div className="group/field">
                    <label className="block text-xs font-semibold uppercase tracking-wider text-gray-400 mb-1">Phone Number</label>
                    <div className="flex items-center justify-between py-2 border-b border-gray-100 dark:border-gray-800 group-hover/field:border-gray-300 dark:group-hover/field:border-gray-700 transition-colors">
                        <span className="text-gray-900 dark:text-white font-medium">{profile.phone || '-'}</span>
                        <Edit2 size={14} className="text-gray-300 opacity-0 group-hover/field:opacity-100 transition-opacity cursor-pointer" />
                    </div>
                </div>

                <div>
                    <label className="block text-xs font-semibold uppercase tracking-wider text-gray-400 mb-1">Joined</label>
                    <div className="text-gray-900 dark:text-white font-medium">
                        {profile.created_at ? format(new Date(profile.created_at), 'MMM d, yyyy') : '-'}
                    </div>
                </div>
            </div>
        </div>
    );
};
