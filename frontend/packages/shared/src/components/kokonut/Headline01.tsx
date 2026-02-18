import { motion } from "framer-motion";
import { cn } from "../../lib/utils";

interface Headline01Props {
    title: string;
    subtitle?: string;
    className?: string;
}

export function Headline01({ title, subtitle, className }: Headline01Props) {
    return (
        <div className={cn("space-y-2", className)}>
            <motion.h1
                initial={{ opacity: 0, y: -10 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.5, ease: "easeOut" }}
                className="text-3xl font-bold tracking-tight bg-gradient-to-r from-zinc-900 to-zinc-600 bg-clip-text text-transparent dark:from-white dark:to-zinc-400 sm:text-4xl"
            >
                {title}
            </motion.h1>
            {subtitle && (
                <motion.p
                    initial={{ opacity: 0 }}
                    animate={{ opacity: 1 }}
                    transition={{ duration: 0.5, delay: 0.2 }}
                    className="text-lg text-zinc-500 dark:text-zinc-400"
                >
                    {subtitle}
                </motion.p>
            )}
        </div>
    );
}
