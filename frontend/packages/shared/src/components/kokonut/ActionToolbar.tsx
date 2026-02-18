import { motion } from "framer-motion";
import type { LucideIcon } from "lucide-react";
import { Button } from "../Button";
import { cn } from "../../lib/utils";

interface Action {
    id: string;
    label: string;
    icon: LucideIcon;
    onClick: () => void;
    variant?: "primary" | "secondary" | "outline" | "ghost";
}

interface ActionToolbarProps {
    actions: Action[];
    className?: string;
}

export function ActionToolbar({ actions, className }: ActionToolbarProps) {
    return (
        <div className={cn("fixed bottom-6 left-1/2 z-50 -translate-x-1/2 transform", className)}>
            <motion.div
                layout
                initial={{ width: 40 }}
                animate={{ width: "auto" }}
                className="flex items-center gap-1 rounded-full border border-zinc-200 bg-white/80 p-1.5 shadow-lg backdrop-blur-lg dark:border-zinc-800 dark:bg-zinc-900/80"
            >
                {actions.map((action) => (
                    <Button
                        key={action.id}
                        variant={action.variant || "ghost"}
                        size="icon"
                        onClick={action.onClick}
                        className="rounded-full h-10 w-10 hover:bg-zinc-100 dark:hover:bg-zinc-800"
                        title={action.label}
                    >
                        <action.icon size={20} className="text-zinc-600 dark:text-zinc-300" />
                    </Button>
                ))}
            </motion.div>
        </div>
    );
}
