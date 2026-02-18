import { motion } from "framer-motion";
import { cn } from "../../lib/utils";
import { Button } from "../Button";
import { MoreHorizontal } from "lucide-react";

interface Card09Props {
    name: string;
    role: string;
    avatarSrc?: string;
    stats: {
        label: string;
        value: string | number;
    }[];
    actions?: {
        label: string;
        onClick: () => void;
        variant?: "primary" | "secondary" | "outline" | "ghost";
    }[];
    className?: string;
}

export function Card09({ name, role, avatarSrc, stats, actions, className }: Card09Props) {
    return (
        <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.4 }}
            className={cn(
                "relative overflow-hidden rounded-2xl border border-zinc-200 bg-white p-6 shadow-sm dark:border-zinc-800 dark:bg-zinc-900",
                className
            )}
        >
            <div className="flex items-start justify-between">
                <div className="flex gap-4">
                    <div className="relative h-16 w-16 overflow-hidden rounded-full border border-gray-100 dark:border-zinc-800">
                        {avatarSrc ? (
                            <img src={avatarSrc} alt={name} className="h-full w-full object-cover" />
                        ) : (
                            <div className="flex h-full w-full items-center justify-center bg-gray-100 text-lg font-medium text-gray-500 dark:bg-zinc-800 dark:text-zinc-400">
                                {name.charAt(0)}
                            </div>
                        )}
                    </div>
                    <div>
                        <h3 className="text-lg font-semibold text-zinc-900 dark:text-zinc-50">{name}</h3>
                        <p className="text-sm text-zinc-500 dark:text-zinc-400">{role}</p>
                    </div>
                </div>
                <Button variant="ghost" size="icon" className="h-8 w-8">
                    <MoreHorizontal size={16} />
                </Button>
            </div>

            <div className="mt-6 grid grid-cols-2 gap-4 divide-x divide-zinc-100 dark:divide-zinc-800">
                {stats.map((stat, index) => (
                    <div key={index} className={cn("px-4", index === 0 && "pl-0")}>
                        <div className="text-2xl font-bold text-zinc-900 dark:text-zinc-50">{stat.value}</div>
                        <div className="text-xs text-zinc-500 dark:text-zinc-400">{stat.label}</div>
                    </div>
                ))}
            </div>

            {actions && actions.length > 0 && (
                <div className="mt-6 flex gap-3">
                    {actions.map((action, index) => (
                        <Button
                            key={index}
                            className="flex-1"
                            variant={action.variant || "primary"}
                            onClick={action.onClick}
                        >
                            {action.label}
                        </Button>
                    ))}
                </div>
            )}
        </motion.div>
    );
}
