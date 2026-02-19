import { useState, type FormEvent } from 'react';
import { useNavigate } from 'react-router-dom';
import { Gavel, AlertCircle } from 'lucide-react';
import { Button, useAuthStore } from '@datalens/shared';
import { adminService } from '@/services/adminService';

const Login = () => {
    const [email, setEmail] = useState('');
    const [password, setPassword] = useState('');
    const [isPending, setIsPending] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const navigate = useNavigate();
    const login = useAuthStore((state) => state.login);

    const handleSubmit = async (e: FormEvent) => {
        e.preventDefault();
        setIsPending(true);
        setError(null);

        try {
            // 1. Authenticate to get tokens
            const { access_token, refresh_token } = await adminService.login(email, password);

            // Set token in store immediately so subsequent requests have auth header
            // We pass null user initially
            login(null as any, access_token, 'system', refresh_token);

            // 2. Fetch current user details
            const adminUser = await adminService.getCurrentUser();
            const user = {
                ...adminUser,
                last_login_at: adminUser.last_login_at ?? undefined
            };

            // 3. Update store with full user details
            login(user, access_token, 'system', refresh_token);

            // 4. Redirect to dashboard
            navigate('/');
        } catch (err: any) {
            console.error('Login failed:', err);
            const msg = err.response?.data?.message || err.message || 'Login failed. Please check your credentials.';
            setError(msg);
            // Clear tokens if failed mid-way
            useAuthStore.getState().logout();
        } finally {
            setIsPending(false);
        }
    };

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
                        SuperAdmin Portal
                    </h1>
                    <p style={{ color: 'var(--text-secondary)', fontSize: '0.875rem' }}>
                        Sign in to manage the DataLens platform
                    </p>
                </div>

                {/* Error */}
                {error && (
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
                        {error}
                    </div>
                )}

                {/* Form */}
                <form onSubmit={handleSubmit} style={{ display: 'flex', flexDirection: 'column', gap: '1.125rem' }}>
                    <div>
                        <label style={labelStyle}>Email address</label>
                        <input
                            type="email"
                            value={email}
                            onChange={(e) => setEmail(e.target.value)}
                            placeholder="admin@datalens.ai"
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
                        Sign in to Console
                    </Button>
                </form>
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
