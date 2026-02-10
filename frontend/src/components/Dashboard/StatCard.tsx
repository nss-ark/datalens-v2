import { type LucideIcon, ArrowUpRight, ArrowDownRight, Minus } from 'lucide-react';
import { cn } from '../../utils/cn';

interface StatCardProps {
    title: string;
    value: string | number;
    icon: LucideIcon;
    color: 'primary' | 'success' | 'warning' | 'danger' | 'info';
    trend?: {
        value: number;
        label: string;
        direction: 'up' | 'down' | 'neutral';
    };
    loading?: boolean;
}

export const StatCard = ({ title, value, icon: Icon, color, trend, loading = false }: StatCardProps) => {
    if (loading) {
        return (
            <div className="bg-white p-6 rounded-lg border border-gray-200 shadow-sm animate-pulse h-[140px]">
                <div className="flex items-center gap-4 mb-4">
                    <div className="w-12 h-12 rounded-md bg-gray-100" />
                    <div className="space-y-2 flex-1">
                        <div className="h-4 bg-gray-100 rounded w-1/2" />
                        <div className="h-8 bg-gray-100 rounded w-3/4" />
                    </div>
                </div>
            </div>
        );
    }

    return (
        <div className="bg-white p-6 rounded-lg border border-gray-200 shadow-sm transition-all duration-200 hover:shadow-md">
            <div className="flex items-start justify-between">
                <div>
                    <p className="text-sm font-medium text-gray-500 mb-1">{title}</p>
                    <h3 className="text-2xl font-bold text-gray-900">{value}</h3>
                </div>
                <div className={cn(
                    "p-3 rounded-md",
                    color === 'primary' && "bg-blue-50 text-blue-600",
                    color === 'success' && "bg-emerald-50 text-emerald-600",
                    color === 'warning' && "bg-amber-50 text-amber-600",
                    color === 'danger' && "bg-red-50 text-red-600",
                    color === 'info' && "bg-indigo-50 text-indigo-600",
                )}>
                    <Icon size={24} />
                </div>
            </div>

            {trend && (
                <div className="mt-4 flex items-center gap-2 text-sm">
                    <span className={cn(
                        "flex items-center font-medium",
                        trend.direction === 'up' && "text-emerald-600",
                        trend.direction === 'down' && "text-red-600",
                        trend.direction === 'neutral' && "text-gray-500",
                    )}>
                        {trend.direction === 'up' && <ArrowUpRight size={16} className="mr-1" />}
                        {trend.direction === 'down' && <ArrowDownRight size={16} className="mr-1" />}
                        {trend.direction === 'neutral' && <Minus size={16} className="mr-1" />}
                        {Math.abs(trend.value)}%
                    </span>
                    <span className="text-gray-500">{trend.label}</span>
                </div>
            )}
        </div>
    );
};
