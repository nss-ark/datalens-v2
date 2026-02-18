import { motion } from "framer-motion";
import { cn } from "../../lib/utils";
import type { LucideIcon } from "lucide-react";

interface BentoItemProps {
    title: string;
    value: string | number;
    icon: LucideIcon;
    description?: string;
    color?: "primary" | "info" | "danger" | "warning";
    className?: string;
    delay?: number;
}

const BentoItem = ({ title, value, icon: Icon, description, color = "primary", className, delay = 0 }: BentoItemProps) => {
    const colorStyles = {
        primary: "bg-blue-50 text-blue-700 dark:bg-blue-900/20 dark:text-blue-400",
        info: "bg-indigo-50 text-indigo-700 dark:bg-indigo-900/20 dark:text-indigo-400",
        danger: "bg-red-50 text-red-700 dark:bg-red-900/20 dark:text-red-400",
        warning: "bg-amber-50 text-amber-700 dark:bg-amber-900/20 dark:text-amber-400",
    };

    return (
        <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.4, delay }}
            className={cn(
                "group relative overflow-hidden rounded-2xl bg-white p-6 shadow-sm transition-all hover:shadow-md dark:bg-zinc-900",
                "border border-zinc-200 dark:border-zinc-800",
                className
            )}
        >
            <div className="flex items-center justify-between">
                <div className={cn("rounded-xl p-2.5 transition-colors", colorStyles[color])}>
                    <Icon size={20} />
                </div>
                {description && (
                    <span className="text-xs font-medium text-zinc-500 dark:text-zinc-400">
                        {description}
                    </span>
                )}
            </div>
            <div className="mt-4">
                <div className="text-3xl font-bold text-zinc-900 dark:text-zinc-50">{value}</div>
                <div className="mt-1 text-sm font-medium text-zinc-500 dark:text-zinc-400">{title}</div>
            </div>

            {/* Decorative background gradient */}
            <div className="absolute -right-4 -top-4 -z-10 h-24 w-24 rounded-full bg-gradient-to-br from-zinc-100 to-transparent opacity-0 transition-opacity group-hover:opacity-100 dark:from-zinc-800" />
        </motion.div>
    );
};

interface Feature09Props {
    items: {
        title: string;
        value: string | number;
        icon: LucideIcon;
        color?: "primary" | "info" | "danger" | "warning";
        description?: string;
    }[];
}

export function Feature09({ items }: Feature09Props) {
    return (
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
            {items.map((item, index) => (
                <BentoItem
                    key={index}
                    {...item}
                    delay={index * 0.1}
                />
            ))}
        </div>
    );
}
