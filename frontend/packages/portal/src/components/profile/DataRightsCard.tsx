import { Gavel } from 'lucide-react';

export const DataRightsCard = ({ activeRequestsCount = 12 }: { activeRequestsCount?: number }) => {
    return (
        <div className="bg-white dark:bg-slate-900 rounded-2xl p-8 border border-gray-200/60 dark:border-gray-800 shadow-sm hover:shadow-md transition-all duration-300 h-full">
            <div className="flex items-center justify-between mb-6">
                <h3 className="text-sm font-bold text-gray-900 dark:text-white uppercase tracking-wider">Data Rights</h3>
                <Gavel className="text-gray-400 w-5 h-5" />
            </div>
            <div className="flex justify-between items-end">
                <div>
                    <div className="text-3xl font-bold text-gray-900 dark:text-white">{activeRequestsCount}</div>
                    <div className="text-xs text-gray-500 dark:text-gray-400 mt-1">Active Requests</div>
                </div>
                <div className="flex -space-x-2">
                    <div className="w-8 h-8 rounded-full bg-blue-100 dark:bg-blue-900 border-2 border-white dark:border-gray-800 flex items-center justify-center text-[10px] text-blue-700 font-bold">A</div>
                    <div className="w-8 h-8 rounded-full bg-indigo-100 dark:bg-indigo-900 border-2 border-white dark:border-gray-800 flex items-center justify-center text-[10px] text-indigo-700 font-bold">M</div>
                    <div className="w-8 h-8 rounded-full bg-gray-100 dark:bg-gray-700 border-2 border-white dark:border-gray-800 flex items-center justify-center text-[10px] text-gray-500 font-bold">+2</div>
                </div>
            </div>
        </div>
    );
};
