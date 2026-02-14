import { Component, type ErrorInfo, type ReactNode } from 'react';

interface Props {
    children: ReactNode;
    fallback?: ReactNode;
    FallbackComponent?: React.ComponentType<{ error: Error; resetErrorBoundary: () => void }>;
    onError?: (error: Error, errorInfo: ErrorInfo) => void;
}

interface State {
    hasError: boolean;
    error: Error | null;
}

export class ErrorBoundary extends Component<Props, State> {
    public state: State = {
        hasError: false,
        error: null,
    };

    public static getDerivedStateFromError(error: Error): State {
        return { hasError: true, error };
    }

    public componentDidCatch(error: Error, errorInfo: ErrorInfo) {
        console.error('Uncaught error:', error, errorInfo);
        this.props.onError?.(error, errorInfo);
    }

    public resetErrorBoundary = () => {
        this.setState({ hasError: false, error: null });
    };

    public render() {
        if (this.state.hasError) {
            if (this.props.FallbackComponent && this.state.error) {
                const FallbackComponent = this.props.FallbackComponent;
                return (
                    <FallbackComponent
                        error={this.state.error}
                        resetErrorBoundary={this.resetErrorBoundary}
                    />
                );
            }
            return this.props.fallback || <h1>Sorry.. there was an error</h1>;
        }

        return this.props.children;
    }
}
