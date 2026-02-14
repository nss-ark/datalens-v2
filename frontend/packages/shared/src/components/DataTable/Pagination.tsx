import { ChevronLeft, ChevronRight } from 'lucide-react';
import styles from './Pagination.module.css';

interface PaginationProps {
    page: number;
    pageSize: number;
    total: number;
    onPageChange: (page: number) => void;
    onPageSizeChange?: (size: number) => void;
    pageSizeOptions?: number[];
}

export function Pagination({
    page,
    pageSize,
    total,
    onPageChange,
    onPageSizeChange,
    pageSizeOptions = [10, 20, 50],
}: PaginationProps) {
    const totalPages = Math.max(1, Math.ceil(total / pageSize));
    const from = total === 0 ? 0 : (page - 1) * pageSize + 1;
    const to = Math.min(page * pageSize, total);

    return (
        <div className={styles.pagination}>
            <div className={styles.info}>
                <span>Showing {from}–{to} of {total}</span>
                {onPageSizeChange && (
                    <select
                        className={styles.pageSizeSelect}
                        value={pageSize}
                        onChange={(e) => onPageSizeChange(Number(e.target.value))}
                    >
                        {pageSizeOptions.map((size) => (
                            <option key={size} value={size}>{size} / page</option>
                        ))}
                    </select>
                )}
            </div>

            <div className={styles.controls}>
                <button
                    className={styles.pageBtn}
                    disabled={page <= 1}
                    onClick={() => onPageChange(page - 1)}
                    aria-label="Previous page"
                >
                    <ChevronLeft size={16} />
                </button>
                <span className={styles.pageIndicator}>
                    {page} / {totalPages}
                </span>
                <button
                    className={styles.pageBtn}
                    disabled={page >= totalPages}
                    onClick={() => onPageChange(page + 1)}
                    aria-label="Next page"
                >
                    <ChevronRight size={16} />
                </button>
            </div>
        </div>
    );
}
