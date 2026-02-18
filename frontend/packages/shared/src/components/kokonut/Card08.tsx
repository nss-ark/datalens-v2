import { motion } from "framer-motion";
import { cn } from "../../lib/utils";

interface Card08Props {
    title: string;
    children: React.ReactNode;
    action?: React.ReactNode;
    className?: string;
}

export function Card08({ title, children, action, className }: Card08Props) {
    return (
        <motion.div
            initial={{ opacity: 0, scale: 0.95 }}
            animate={{ opacity: 1, scale: 1 }}
            transition={{ duration: 0.4 }}
            className={cn(
                "rounded-xl border border-zinc-200 bg-white shadow-sm dark:border-zinc-800 dark:bg-zinc-900",
                className
            )}
        >
            <div className="flex flex-row items-center justify-between space-y-0 p-6 pb-2">
                <h3 className="font-semibold leading-none tracking-tight text-zinc-900 dark:text-zinc-50">{title}</h3>
                {action && <div>{action}</div>}
            </div>
            <div className="p-6 pt-0">
                {children}
            </div>
        </motion.div>
    );
}
