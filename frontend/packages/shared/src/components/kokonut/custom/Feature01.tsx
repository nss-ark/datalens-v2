import { ArrowRight, type LucideIcon } from 'lucide-react';
import { cn } from '../../../lib/utils';

interface FeatureCardProps {
    title: string;
    description: string;
    icon: LucideIcon;
    href?: string;
    onClick?: () => void;
    className?: string;
}

export const FeatureCard = ({
    title,
    description,
    icon: Icon,
    // href,
    onClick,
    className
}: FeatureCardProps) => {
    return (
        <div
            onClick={onClick}
            className={cn(
                "group relative p-6 bg-white rounded-2xl border border-slate-100 shadow-sm hover:shadow-md hover:border-blue-100 transition-all duration-300 cursor-pointer overflow-hidden",
                className
            )}
        >
            <div className="absolute top-0 right-0 w-32 h-32 bg-gradient-to-br from-blue-50 to-transparent opacity-0 group-hover:opacity-100 transition-opacity rounded-bl-full" />

            <div className="relative z-10">
                <div className="h-12 w-12 bg-blue-50 rounded-xl flex items-center justify-center mb-4 group-hover:scale-110 transition-transform duration-300">
                    <Icon className="h-6 w-6 text-blue-600" />
                </div>

                <h3 className="text-lg font-semibold text-slate-900 mb-2 group-hover:text-blue-700 transition-colors">
                    {title}
                </h3>

                <p className="text-slate-500 text-sm mb-4 line-clamp-2">
                    {description}
                </p>

                <div className="flex items-center text-sm font-medium text-blue-600 opacity-0 group-hover:opacity-100 transform translate-y-2 group-hover:translate-y-0 transition-all duration-300">
                    View Details <ArrowRight className="ml-1 w-4 h-4" />
                </div>
            </div>
        </div>
    );
};

interface FeatureGridProps {
    features: Omit<FeatureCardProps, 'className'>[];
    className?: string;
}

export const Feature01 = ({ features, className }: FeatureGridProps) => {
    return (
        <div className={cn("grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6", className)}>
            {features.map((feature, index) => (
                <FeatureCard
                    key={index}
                    {...feature}
                    className="animate-fade-in-up"
                // Add staggered delay via style if needed, or rely on CSS/Library
                />
            ))}
        </div>
    );
};
