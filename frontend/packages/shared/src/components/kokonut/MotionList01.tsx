import { motion } from "framer-motion";
import { cn } from "../../lib/utils";

interface MotionList01Props {
    children: React.ReactNode;
    className?: string;
    delay?: number;
}

export function MotionList01({ children, className, delay = 0 }: MotionList01Props) {
    return (
        <motion.div
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.4, delay, ease: "easeOut" }}
            className={cn("w-full", className)}
        >
            {children}
        </motion.div>
    );
}

export const MotionItem = ({ children, index = 0, className }: { children: React.ReactNode; index?: number; className?: string }) => {
    return (
        <motion.div
            initial={{ opacity: 0, x: -10 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ duration: 0.3, delay: index * 0.05 }}
            className={className}
        >
            {children}
        </motion.div>
    );
};
