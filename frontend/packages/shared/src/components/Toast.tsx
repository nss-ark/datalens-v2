import { X, CheckCircle, AlertCircle, AlertTriangle, Info } from 'lucide-react';
import { useToastStore, type ToastVariant } from '../stores/toastStore';
import { cn } from '../lib/utils';
import styles from './Toast.module.css';

const ICONS: Record<ToastVariant, typeof CheckCircle> = {
    success: CheckCircle,
    error: AlertCircle,
    warning: AlertTriangle,
    info: Info,
};

export function ToastContainer() {
    const { toasts, removeToast } = useToastStore();

    if (toasts.length === 0) return null;

    return (
        <div className={styles.container}>
            {toasts.map((t) => {
                const Icon = ICONS[t.variant];
                return (
                    <div key={t.id} className={cn(styles.toast, styles[t.variant])}>
                        <Icon size={18} className={styles.icon} />
                        <div className={styles.content}>
                            <div className={styles.title}>{t.title}</div>
                            {t.message && <div className={styles.message}>{t.message}</div>}
                        </div>
                        <button className={styles.closeBtn} onClick={() => removeToast(t.id)}>
                            <X size={14} />
                        </button>
                    </div>
                );
            })}
        </div>
    );
}
