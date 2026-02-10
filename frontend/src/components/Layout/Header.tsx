import { Search, Bell, HelpCircle } from 'lucide-react';
import styles from './Header.module.css';

interface HeaderProps {
    title?: string;
}

export const Header = ({ title }: HeaderProps) => {
    return (
        <header className={styles.header}>
            <div className={styles.leftSection}>
                <h1 className={styles.title}>{title || 'Dashboard'}</h1>
            </div>

            <div className={styles.rightSection}>
                <div className={styles.search}>
                    <Search className={styles.searchIcon} size={18} />
                    <input
                        type="text"
                        placeholder="Search anything..."
                        className={styles.searchInput}
                    />
                </div>

                <button className={styles.iconBtn}>
                    <HelpCircle size={20} />
                </button>

                <button className={styles.iconBtn}>
                    <Bell size={20} />
                    <span className={styles.badge}>3</span>
                </button>
            </div>
        </header>
    );
};
