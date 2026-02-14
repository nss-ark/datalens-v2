import { useState, type FormEvent } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { Gavel, AlertCircle, CheckCircle } from 'lucide-react';
import { Button } from '@datalens/shared';
import { useRegister } from '@datalens/shared';

const Register = () => {
    const [form, setForm] = useState({
        tenant_name: '',
        domain: '',
        industry: '',
        country: '',
        email: '',
        name: '',
        password: '',
    });
    const [success, setSuccess] = useState(false);
    const navigate = useNavigate();
    const { mutate: registerMutate, isPending, error } = useRegister();

    const handleChange = (field: string) => (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
        setForm((prev) => ({ ...prev, [field]: e.target.value }));
    };

    const handleSubmit = (e: FormEvent) => {
        e.preventDefault();
        registerMutate(form, {
            onSuccess: () => {
                setSuccess(true);
                setTimeout(() => navigate('/login'), 2000);
            },
        });
    };

    const apiError = error as { response?: { data?: { message?: string } } } | null;
    const errorMessage = apiError?.response?.data?.message || (error ? 'Registration failed.' : null);

    return (
        <div style={{
            minHeight: '100vh',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            backgroundColor: 'var(--slate-50)',
            padding: '2rem',
        }}>
            <div style={{
                width: '100%',
                maxWidth: '480px',
                padding: '2.5rem',
                backgroundColor: 'white',
                borderRadius: 'var(--radius-lg)',
                border: '1px solid var(--border-color)',
                boxShadow: 'var(--shadow-lg)',
            }}>
                {/* Header */}
                <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', marginBottom: '2rem' }}>
                    <div style={{
                        width: '52px', height: '52px', borderRadius: '14px',
                        backgroundColor: 'var(--primary-100)', color: 'var(--primary-600)',
                        display: 'flex', alignItems: 'center', justifyContent: 'center', marginBottom: '1.25rem',
                    }}>
                        <Gavel size={28} />
                    </div>
                    <h1 style={{ fontSize: '1.5rem', fontWeight: 700, color: 'var(--text-primary)', marginBottom: '0.25rem' }}>
                        Create your account
                    </h1>
                    <p style={{ color: 'var(--text-secondary)', fontSize: '0.875rem' }}>
                        Set up your organization on DataLens
                    </p>
                </div>

                {success && (
                    <div style={{
                        display: 'flex', alignItems: 'center', gap: '0.5rem', padding: '0.75rem 1rem',
                        marginBottom: '1.25rem', borderRadius: 'var(--radius-md)',
                        backgroundColor: 'var(--success-50)', color: 'var(--success-700)',
                        fontSize: '0.875rem', border: '1px solid var(--success-200)',
                    }}>
                        <CheckCircle size={16} />
                        Account created successfully! Redirecting to login...
                    </div>
                )}

                {errorMessage && (
                    <div style={{
                        display: 'flex', alignItems: 'center', gap: '0.5rem', padding: '0.75rem 1rem',
                        marginBottom: '1.25rem', borderRadius: 'var(--radius-md)',
                        backgroundColor: 'var(--danger-50)', color: 'var(--danger-700)',
                        fontSize: '0.875rem', border: '1px solid var(--danger-200)',
                    }}>
                        <AlertCircle size={16} />
                        {errorMessage}
                    </div>
                )}

                <form onSubmit={handleSubmit} style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
                    <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem' }}>
                        <div>
                            <label style={labelStyle}>Organization Name</label>
                            <input type="text" value={form.tenant_name} onChange={handleChange('tenant_name')} required style={inputStyle} placeholder="Acme Corp" />
                        </div>
                        <div>
                            <label style={labelStyle}>Domain</label>
                            <input type="text" value={form.domain} onChange={handleChange('domain')} required style={inputStyle} placeholder="acme-corp" />
                        </div>
                    </div>

                    <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem' }}>
                        <div>
                            <label style={labelStyle}>Industry</label>
                            <input type="text" value={form.industry} onChange={handleChange('industry')} style={inputStyle} placeholder="Technology" />
                        </div>
                        <div>
                            <label style={labelStyle}>Country</label>
                            <input type="text" value={form.country} onChange={handleChange('country')} style={inputStyle} placeholder="India" />
                        </div>
                    </div>

                    <div>
                        <label style={labelStyle}>Full Name</label>
                        <input type="text" value={form.name} onChange={handleChange('name')} required style={inputStyle} placeholder="Jane Doe" />
                    </div>

                    <div>
                        <label style={labelStyle}>Email</label>
                        <input type="email" value={form.email} onChange={handleChange('email')} required style={inputStyle} placeholder="jane@acme.com" />
                    </div>

                    <div>
                        <label style={labelStyle}>Password</label>
                        <input type="password" value={form.password} onChange={handleChange('password')} required minLength={8} style={inputStyle} placeholder="Min 8 characters" />
                    </div>

                    <Button type="submit" isLoading={isPending} style={{ width: '100%', marginTop: '0.5rem' }}>
                        Create Account
                    </Button>
                </form>

                <p style={{ textAlign: 'center', marginTop: '1.5rem', fontSize: '0.875rem', color: 'var(--text-secondary)' }}>
                    Already have an account?{' '}
                    <Link to="/login" style={{ color: 'var(--primary-600)', fontWeight: 500, textDecoration: 'none' }}>
                        Sign in
                    </Link>
                </p>
            </div>
        </div>
    );
};

const labelStyle: React.CSSProperties = {
    display: 'block', fontSize: '0.875rem', fontWeight: 500,
    color: 'var(--text-primary)', marginBottom: '0.375rem',
};

const inputStyle: React.CSSProperties = {
    width: '100%', height: '42px', padding: '0 0.875rem',
    borderRadius: 'var(--radius-md)', border: '1px solid var(--border-color)',
    fontSize: '0.875rem', outline: 'none', boxSizing: 'border-box',
};

export default Register;
