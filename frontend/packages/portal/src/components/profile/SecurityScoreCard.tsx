import { Shield } from 'lucide-react';

export const SecurityScoreCard = () => {
    // Mock data for now
    const score = 85;

    return (
        <div className="bg-white dark:bg-slate-900 rounded-2xl p-8 border border-gray-200/60 dark:border-gray-800 shadow-sm hover:shadow-md transition-all duration-300 h-full">
            <div className="flex items-center justify-between mb-4">
                <h3 className="text-sm font-bold text-gray-900 dark:text-white uppercase tracking-wider">Security Score</h3>
                <Shield className="text-gray-400 w-5 h-5" />
            </div>
            <div className="flex items-center gap-6">
                <div className="relative w-16 h-16 flex items-center justify-center">
                    {/* Simplified circular progress using conic-gradient */}
                    <div
                        className="w-16 h-16 rounded-full flex items-center justify-center"
                        style={{ background: `conic-gradient(#10b981 ${score}%, #e2e8f0 0)` }}
                    >
                        <div className="w-12 h-12 bg-white dark:bg-slate-900 rounded-full flex items-center justify-center">
                            <span className="text-sm font-bold text-gray-900 dark:text-white">{score}</span>
                        </div>
                    </div>
                </div>
                <div>
                    <div className="text-sm font-medium text-emerald-600 dark:text-emerald-400 mb-1">High Security</div>
                    <p className="text-xs text-gray-500 dark:text-gray-400">2 actions suggested</p>
                </div>
            </div>
        </div>
    );
};
