import { useState, type FormEvent } from 'react';
import { Link } from 'react-router-dom';
import { Gavel, AlertCircle } from 'lucide-react';
import { Button } from '@datalens/shared';
import { useLogin } from '@datalens/shared';

const Login = () => {
    const [domain, setDomain] = useState('');
    const [email, setEmail] = useState('');
    const [password, setPassword] = useState('');
    const { mutate: loginMutate, isPending, error } = useLogin();

    const handleSubmit = (e: FormEvent) => {
        e.preventDefault();
        loginMutate({ domain, email, password });
    };

    const apiError = error as { response?: { data?: { message?: string } } } | null;
    const errorMessage = apiError?.response?.data?.message || (error ? 'Login failed. Please check your credentials.' : null);

    return (
        <div style={{
            minHeight: '100vh',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            backgroundColor: 'var(--slate-50)',
        }}>
            <div style={{
                width: '100%',
                maxWidth: '420px',
                padding: '2.5rem',
                backgroundColor: 'white',
                borderRadius: 'var(--radius-lg)',
                border: '1px solid var(--border-color)',
                boxShadow: 'var(--shadow-lg)',
            }}>
                {/* Header */}
                <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', marginBottom: '2rem' }}>
                    <div style={{
                        width: '52px',
                        height: '52px',
                        borderRadius: '14px',
                        backgroundColor: 'var(--primary-100)',
                        color: 'var(--primary-600)',
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        marginBottom: '1.25rem',
                    }}>
                        <Gavel size={28} />
                    </div>
                    <h1 style={{ fontSize: '1.5rem', fontWeight: 700, color: 'var(--text-primary)', marginBottom: '0.25rem' }}>
                        Welcome back
                    </h1>
                    <p style={{ color: 'var(--text-secondary)', fontSize: '0.875rem' }}>
                        Sign in to your DataLens account
                    </p>
                </div>

                {/* Error */}
                {errorMessage && (
                    <div style={{
                        display: 'flex',
                        alignItems: 'center',
                        gap: '0.5rem',
                        padding: '0.75rem 1rem',
                        marginBottom: '1.25rem',
                        borderRadius: 'var(--radius-md)',
                        backgroundColor: 'var(--danger-50)',
                        color: 'var(--danger-700)',
                        fontSize: '0.875rem',
                        border: '1px solid var(--danger-200)',
                    }}>
                        <AlertCircle size={16} />
                        {errorMessage}
                    </div>
                )}

                {/* Form */}
                <form onSubmit={handleSubmit} style={{ display: 'flex', flexDirection: 'column', gap: '1.125rem' }}>
                    <div>
                        <label style={labelStyle}>Organization Domain (Optional)</label>
                        <input
                            type="text"
                            value={domain}
                            onChange={(e) => setDomain(e.target.value)}
                            placeholder="e.g. acme-corp"
                            style={inputStyle}
                        />
                        <small style={{ fontSize: '0.75rem', color: 'var(--text-tertiary)', marginTop: '4px', display: 'block' }}>
                            Leave blank to sign in with email only
                        </small>
                    </div>

                    <div>
                        <label style={labelStyle}>Email address</label>
                        <input
                            type="email"
                            value={email}
                            onChange={(e) => setEmail(e.target.value)}
                            placeholder="you@company.com"
                            required
                            style={inputStyle}
                        />
                    </div>

                    <div>
                        <label style={labelStyle}>Password</label>
                        <input
                            type="password"
                            value={password}
                            onChange={(e) => setPassword(e.target.value)}
                            placeholder="••••••••"
                            required
                            minLength={8}
                            style={inputStyle}
                        />
                    </div>

                    <Button type="submit" isLoading={isPending} style={{ width: '100%', marginTop: '0.5rem' }}>
                        Sign in
                    </Button>
                </form>

                <p style={{ textAlign: 'center', marginTop: '1.5rem', fontSize: '0.875rem', color: 'var(--text-secondary)' }}>
                    Don't have an account?{' '}
                    <Link to="/register" style={{ color: 'var(--primary-600)', fontWeight: 500, textDecoration: 'none' }}>
                        Register
                    </Link>
                </p>
            </div>
        </div>
    );
};

const labelStyle: React.CSSProperties = {
    display: 'block',
    fontSize: '0.875rem',
    fontWeight: 500,
    color: 'var(--text-primary)',
    marginBottom: '0.375rem',
};

const inputStyle: React.CSSProperties = {
    width: '100%',
    height: '42px',
    padding: '0 0.875rem',
    borderRadius: 'var(--radius-md)',
    border: '1px solid var(--border-color)',
    fontSize: '0.875rem',
    outline: 'none',
    transition: 'border-color 0.15s',
    boxSizing: 'border-box',
};

export default Login;
