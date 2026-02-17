import React from 'react';
import type { LucideIcon } from 'lucide-react';
import { TrendingUp, TrendingDown } from 'lucide-react';

interface StatsCardProps {
    title: string;
    value: string | number;
    icon?: LucideIcon;
    trend?: {
        value: string;
        isPositive: boolean;
    };
    color?: 'blue' | 'green' | 'orange' | 'purple' | 'red';
    /** When true, renders a larger variant for bento hero cells */
    large?: boolean;
}

const colorMap = {
    blue: {
        gradient: 'from-blue-500 to-blue-600',
        light: 'bg-blue-50',
        text: 'text-blue-600',
        glow: 'shadow-blue-200/50',
    },
    green: {
        gradient: 'from-emerald-500 to-emerald-600',
        light: 'bg-emerald-50',
        text: 'text-emerald-600',
        glow: 'shadow-emerald-200/50',
    },
    orange: {
        gradient: 'from-orange-500 to-orange-600',
        light: 'bg-orange-50',
        text: 'text-orange-600',
        glow: 'shadow-orange-200/50',
    },
    purple: {
        gradient: 'from-purple-500 to-purple-600',
        light: 'bg-purple-50',
        text: 'text-purple-600',
        glow: 'shadow-purple-200/50',
    },
    red: {
        gradient: 'from-red-500 to-red-600',
        light: 'bg-red-50',
        text: 'text-red-600',
        glow: 'shadow-red-200/50',
    },
};

export const StatsCard: React.FC<StatsCardProps> = ({ title, value, icon: Icon, trend, color = 'blue', large = false }) => {
    const c = colorMap[color];

    return (
        <div className="group relative bg-white rounded-2xl border border-slate-200/60 p-5 hover:shadow-lg hover:border-slate-300/60 transition-all duration-300 overflow-hidden h-full min-h-[140px] flex flex-col justify-between">
            {/* Subtle gradient background on hover */}
            <div className={`absolute inset-0 bg-gradient-to-br ${c.gradient} opacity-0 group-hover:opacity-[0.02] transition-opacity duration-300`} />

            <div className="relative z-10 flex items-start justify-between gap-4">
                <div className="flex flex-col gap-1">
                    <p className="text-xs font-semibold text-slate-500 uppercase tracking-widest">{title}</p>
                    <h3 className={`${large ? 'text-4xl' : 'text-3xl'} font-extrabold text-slate-900 tracking-tight leading-none`}>{value}</h3>
                </div>
                {Icon && (
                    <div className={`p-2.5 rounded-xl bg-gradient-to-br ${c.gradient} text-white shadow-md ${c.glow} transition-transform duration-500 group-hover:scale-110 group-hover:rotate-6`}>
                        <Icon className={`${large ? 'w-6 h-6' : 'w-5 h-5'}`} />
                    </div>
                )}
            </div>

            {trend && (
                <div className={`relative z-10 flex items-center gap-1.5 mt-4 text-xs font-semibold ${trend.isPositive ? 'text-emerald-600' : 'text-red-500'}`}>
                    <span className={`flex items-center justify-center w-5 h-5 rounded-full ${trend.isPositive ? 'bg-emerald-100' : 'bg-red-100'}`}>
                        {trend.isPositive ? <TrendingUp className="w-3 h-3" /> : <TrendingDown className="w-3 h-3" />}
                    </span>
                    <span>{trend.value}</span>
                    <span className="text-slate-400 font-medium">vs last month</span>
                </div>
            )}

            {!trend && <div className="mt-4" />}
        </div>
    );
};
