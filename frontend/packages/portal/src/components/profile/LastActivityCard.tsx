import { History } from 'lucide-react';

export const LastActivityCard = () => {
    return (
        <div className="bg-white dark:bg-slate-900 rounded-2xl p-8 border border-gray-200/60 dark:border-gray-800 shadow-sm hover:shadow-md transition-all duration-300 h-full">
            <div className="flex items-center justify-between mb-6">
                <h3 className="text-sm font-bold text-gray-900 dark:text-white uppercase tracking-wider">Last Activity</h3>
                <History className="text-gray-400 w-5 h-5" />
            </div>
            <div className="flex items-start gap-3 mb-2">
                <div className="w-2 h-2 rounded-full bg-green-500 mt-1.5"></div>
                <div>
                    <p className="text-sm font-medium text-gray-900 dark:text-white">Successful Login</p>
                    <p className="text-xs text-gray-500 dark:text-gray-400">MacOS • Chrome</p>
                </div>
            </div>
            <p className="text-xs text-gray-400 mt-2 pl-5">Today, 10:23 AM</p>
        </div>
    );
};
