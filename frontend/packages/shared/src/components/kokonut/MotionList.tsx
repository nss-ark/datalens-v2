import { motion, AnimatePresence } from 'framer-motion';
import { cn } from '../../lib/utils'; // Assuming you have a utils file for merging classnames

interface MotionListProps<T> {
    items: T[];
    renderItem: (item: T, index: number) => React.ReactNode;
    className?: string;
    staggerDelay?: number;
}

export function MotionList<T extends { id: string | number }>({
    items,
    renderItem,
    className,
    staggerDelay = 0.05
}: MotionListProps<T>) {
    return (
        <div className={cn("space-y-2", className)}>
            <AnimatePresence mode='popLayout'>
                {items.map((item, index) => (
                    <motion.div
                        key={item.id}
                        initial={{ opacity: 0, y: 20, filter: 'blur(4px)' }}
                        animate={{ opacity: 1, y: 0, filter: 'blur(0px)' }}
                        exit={{ opacity: 0, scale: 0.95, filter: 'blur(4px)' }}
                        transition={{
                            duration: 0.3,
                            delay: index * staggerDelay,
                            type: 'spring',
                            stiffness: 260,
                            damping: 20
                        }}
                    >
                        {renderItem(item, index)}
                    </motion.div>
                ))}
            </AnimatePresence>
        </div>
    );
}
