import { useState } from "react";
import { motion, AnimatePresence } from "framer-motion";
import { cn } from "../../lib/utils";
import { ChevronDown } from "lucide-react";

interface Accordion01Props {
    items: {
        id: string;
        title: string;
        content: string;
    }[];
    className?: string;
}

export function Accordion01({ items, className }: Accordion01Props) {
    const [openId, setOpenId] = useState<string | null>(null);

    return (
        <div className={cn("w-full space-y-2", className)}>
            {items.map((item) => (
                <div key={item.id} className="rounded-lg border border-zinc-200 bg-white px-4 dark:border-zinc-800 dark:bg-zinc-900">
                    <button
                        onClick={() => setOpenId(openId === item.id ? null : item.id)}
                        className="flex w-full items-center justify-between py-4 text-left text-sm font-medium transition-all hover:text-zinc-900 dark:hover:text-zinc-50"
                    >
                        {item.title}
                        <ChevronDown
                            className={cn(
                                "h-4 w-4 shrink-0 transition-transform duration-200 text-zinc-500",
                                openId === item.id && "rotate-180"
                            )}
                        />
                    </button>
                    <AnimatePresence initial={false}>
                        {openId === item.id && (
                            <motion.div
                                initial={{ height: 0, opacity: 0 }}
                                animate={{ height: "auto", opacity: 1 }}
                                exit={{ height: 0, opacity: 0 }}
                                transition={{ duration: 0.3 }}
                                className="overflow-hidden"
                            >
                                <div className="pb-4 pt-0 text-sm text-zinc-500 dark:text-zinc-400">
                                    {item.content}
                                </div>
                            </motion.div>
                        )}
                    </AnimatePresence>
                </div>
            ))}
        </div>
    );
}
