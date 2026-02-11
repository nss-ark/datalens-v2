import { AlertTriangle, RefreshCw, Home } from 'lucide-react';
import { Button } from './Button';

interface ErrorFallbackProps {
    error: Error;
    resetErrorBoundary?: () => void;
}

export const GlobalErrorFallback = ({ error }: ErrorFallbackProps) => {
    return (
        <div className="min-h-screen flex items-center justify-center bg-gray-50 p-4">
            <div className="max-w-md w-full bg-white rounded-lg shadow-lg p-8 text-center">
                <div className="mx-auto flex items-center justify-center h-16 w-16 rounded-full bg-red-100 mb-6">
                    <AlertTriangle className="h-8 w-8 text-red-600" />
                </div>
                <h2 className="text-2xl font-bold text-gray-900 mb-2">Something went wrong</h2>
                <p className="text-gray-500 mb-6">
                    We encountered an unexpected error. Our team has been notified.
                </p>
                <div className="bg-red-50 p-4 rounded-md text-left mb-6 overflow-auto max-h-40">
                    <pre className="text-xs text-red-800 whitespace-pre-wrap font-mono">
                        {error.message}
                    </pre>
                </div>
                <div className="flex flex-col gap-3">
                    <Button
                        onClick={() => window.location.reload()}
                        variant="primary"
                        className="w-full justify-center"
                        icon={<RefreshCw size={16} />}
                    >
                        Reload Page
                    </Button>
                    <Button
                        onClick={() => window.location.href = '/'}
                        variant="secondary"
                        className="w-full justify-center"
                        icon={<Home size={16} />}
                    >
                        Go to Dashboard
                    </Button>
                </div>
            </div>
        </div>
    );
};

export const SectionErrorFallback = ({ error, resetErrorBoundary }: ErrorFallbackProps) => {
    return (
        <div className="border border-red-200 bg-red-50 rounded-lg p-6 flex flex-col items-center justify-center min-h-[200px] text-center">
            <AlertTriangle className="h-8 w-8 text-red-500 mb-3" />
            <h3 className="text-base font-semibold text-red-900 mb-1">Component Error</h3>
            <p className="text-sm text-red-700 mb-4 max-w-sm">
                This section could not be loaded.
                <br />
                <span className="text-xs opacity-75">{error.message}</span>
            </p>
            {resetErrorBoundary && (
                <Button
                    onClick={resetErrorBoundary}
                    variant="secondary"
                    size="sm"
                    icon={<RefreshCw size={14} />}
                >
                    Try Again
                </Button>
            )}
        </div>
    );
};
