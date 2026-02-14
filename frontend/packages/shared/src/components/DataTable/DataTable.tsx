import { useState, type ReactNode } from 'react';
import { ArrowUpDown, ArrowUp, ArrowDown, Inbox } from 'lucide-react';
import { cn } from '../../lib/utils';
import styles from './DataTable.module.css';

// Column definition
export interface Column<T> {
    key: string;
    header: string;
    sortable?: boolean;
    width?: string;
    render?: (row: T) => ReactNode;
}

export type SortDirection = 'asc' | 'desc';

export interface SortState {
    key: string;
    direction: SortDirection;
}

interface DataTableProps<T> {
    columns: Column<T>[];
    data: T[];
    isLoading?: boolean;
    onRowClick?: (row: T) => void;
    emptyTitle?: string;
    emptyDescription?: string;
    keyExtractor: (row: T) => string;
    onSort?: (sort: SortState) => void;
    defaultSort?: SortState;
    loadingRows?: number;
}

export function DataTable<T>({
    columns,
    data = [],
    isLoading = false,
    onRowClick,
    emptyTitle = 'No data found',
    emptyDescription = 'There are no records to display.',
    keyExtractor,
    onSort,
    defaultSort,
    loadingRows = 5,
}: DataTableProps<T>) {
    const [sort, setSort] = useState<SortState | undefined>(defaultSort);

    const handleSort = (key: string) => {
        const newSort: SortState = {
            key,
            direction: sort?.key === key && sort.direction === 'asc' ? 'desc' : 'asc',
        };
        setSort(newSort);
        onSort?.(newSort);
    };

    // Client-side sort if no onSort handler
    const safeData = Array.isArray(data) ? data : [];
    const sortedData = !onSort && sort
        ? [...safeData].sort((a, b) => {
            const aVal = (a as Record<string, unknown>)[sort.key];
            const bVal = (b as Record<string, unknown>)[sort.key];
            if (aVal == null) return 1;
            if (bVal == null) return -1;
            const cmp = String(aVal).localeCompare(String(bVal));
            return sort.direction === 'asc' ? cmp : -cmp;
        })
        : safeData;

    const renderSortIcon = (key: string) => {
        if (sort?.key !== key) return <ArrowUpDown size={14} />;
        return sort.direction === 'asc' ? <ArrowUp size={14} /> : <ArrowDown size={14} />;
    };

    return (
        <div className={styles.wrapper}>
            <table className={styles.table}>
                <thead>
                    <tr className={styles.headerRow}>
                        {columns.map((col) => (
                            <th
                                key={col.key}
                                className={cn(
                                    styles.headerCell,
                                    col.sortable && styles.sortable,
                                    sort?.key === col.key && styles.sortActive
                                )}
                                style={col.width ? { width: col.width } : undefined}
                                onClick={() => col.sortable && handleSort(col.key)}
                            >
                                {col.header}
                                {col.sortable && (
                                    <span className={styles.sortIcon}>
                                        {renderSortIcon(col.key)}
                                    </span>
                                )}
                            </th>
                        ))}
                    </tr>
                </thead>
                <tbody>
                    {isLoading ? (
                        Array.from({ length: loadingRows }).map((_, i) => (
                            <tr key={`skeleton-${i}`} className={styles.loadingRow}>
                                {columns.map((col) => (
                                    <td key={col.key} className={styles.cell}>
                                        <div className={styles.skeleton} style={{ width: `${60 + Math.random() * 30}%` }} />
                                    </td>
                                ))}
                            </tr>
                        ))
                    ) : sortedData.length === 0 ? (
                        <tr>
                            <td colSpan={columns.length} className={styles.cell}>
                                <div className={styles.emptyState}>
                                    <Inbox size={48} className={styles.emptyIcon} />
                                    <div className={styles.emptyTitle}>{emptyTitle}</div>
                                    <div className={styles.emptyDescription}>{emptyDescription}</div>
                                </div>
                            </td>
                        </tr>
                    ) : (
                        sortedData.map((row) => (
                            <tr
                                key={keyExtractor(row)}
                                className={cn(styles.row, onRowClick && styles.clickableRow)}
                                onClick={() => onRowClick?.(row)}
                            >
                                {columns.map((col) => (
                                    <td key={col.key} className={styles.cell}>
                                        {col.render
                                            ? col.render(row)
                                            : String((row as Record<string, unknown>)[col.key] ?? '')}
                                    </td>
                                ))}
                            </tr>
                        ))
                    )}
                </tbody>
            </table>
        </div>
    );
}
