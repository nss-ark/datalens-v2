import React, { type ButtonHTMLAttributes } from 'react';
import { Loader2 } from 'lucide-react';
import styles from './Button.module.css';
import { cn } from '../../utils/cn';

export interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
    variant?: 'primary' | 'secondary' | 'outline' | 'ghost' | 'danger';
    size?: 'sm' | 'md' | 'lg' | 'icon';
    isLoading?: boolean;
    icon?: React.ReactNode;
    leftIcon?: React.ReactNode;
    rightIcon?: React.ReactNode;
}

export const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
    (
        {
            className,
            variant = 'primary',
            size = 'md',
            isLoading,
            icon,
            leftIcon,
            rightIcon,
            children,
            disabled,
            ...props
        },
        ref
    ) => {
        const resolvedLeftIcon = leftIcon || icon;

        return (
            <button
                ref={ref}
                className={cn(
                    styles.button,
                    styles[variant],
                    styles[size],
                    isLoading && styles.loading,
                    className
                )}
                disabled={disabled || isLoading}
                {...props}
            >
                {isLoading && <Loader2 className={styles.spinner} size={16} />}
                {!isLoading && resolvedLeftIcon && <span className={styles.icon}>{resolvedLeftIcon}</span>}
                {children}
                {!isLoading && rightIcon && <span className={styles.icon}>{rightIcon}</span>}
            </button>
        );
    }
);

Button.displayName = 'Button';

