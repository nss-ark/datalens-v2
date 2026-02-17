import React from 'react';
import { ArrowRight } from 'lucide-react';
import type { LucideIcon } from 'lucide-react';
import { useNavigate } from 'react-router-dom';

interface ActionCardProps {
    title: string;
    description: string;
    icon: LucideIcon;
    to: string;
    color?: 'blue' | 'indigo' | 'purple' | 'emerald';
}

const colorStyles = {
    blue: {
        gradient: 'from-blue-500 to-blue-600',
        glow: 'group-hover:shadow-blue-100',
        accent: 'text-blue-600',
        bg: 'bg-blue-50',
    },
    indigo: {
        gradient: 'from-indigo-500 to-indigo-600',
        glow: 'group-hover:shadow-indigo-100',
        accent: 'text-indigo-600',
        bg: 'bg-indigo-50',
    },
    purple: {
        gradient: 'from-purple-500 to-purple-600',
        glow: 'group-hover:shadow-purple-100',
        accent: 'text-purple-600',
        bg: 'bg-purple-50',
    },
    emerald: {
        gradient: 'from-emerald-500 to-emerald-600',
        glow: 'group-hover:shadow-emerald-100',
        accent: 'text-emerald-600',
        bg: 'bg-emerald-50',
    },
};

export const ActionCard: React.FC<ActionCardProps> = ({ title, description, icon: Icon, to, color = 'blue' }) => {
    const navigate = useNavigate();
    const s = colorStyles[color];

    return (
        <button
            onClick={() => navigate(to)}
            className={`group relative flex flex-col items-start p-6 bg-white border border-slate-200/60 rounded-2xl hover:border-slate-300/80 hover:shadow-xl ${s.glow} transition-all duration-300 text-left w-full h-full min-h-[180px] overflow-hidden`}
        >
            {/* Decorative corner gradient */}
            <div className={`absolute -top-16 -right-16 w-32 h-32 bg-gradient-to-br ${s.gradient} rounded-full opacity-0 group-hover:opacity-[0.08] blur-3xl transition-all duration-500`} />

            <div className="flex w-full items-start justify-between mb-5">
                <div className={`p-2.5 rounded-xl bg-gradient-to-br ${s.gradient} text-white shadow-md transition-all duration-300 group-hover:scale-110 group-hover:rotate-3`}>
                    <Icon className="w-5 h-5" />
                </div>
                <div className={`p-1.5 rounded-full ${s.bg} opacity-0 group-hover:opacity-100 transition-all duration-300 -mr-1`}>
                    <ArrowRight className={`w-4 h-4 ${s.accent} -rotate-45 group-hover:rotate-0 transition-transform duration-300`} />
                </div>
            </div>

            <h3 className="text-lg font-bold text-slate-900 group-hover:text-slate-800 transition-colors mb-2 tracking-tight">
                {title}
            </h3>
            <p className="text-sm text-slate-500 leading-relaxed max-w-[90%]">
                {description}
            </p>
        </button>
    );
};
