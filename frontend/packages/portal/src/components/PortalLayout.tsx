import React, { useState, useEffect, useRef } from 'react';
import { Shield, LayoutDashboard, Bell, User, LogOut, Menu, X, ChevronDown, FileText } from 'lucide-react';
import { Link, useLocation, useNavigate } from 'react-router-dom';
import { usePortalAuthStore } from '@/stores/portalAuthStore';

interface PortalLayoutProps {
    children: React.ReactNode;
}

/* ── Footer styles (inline for guaranteed rendering) ── */
const footerStyles = {
    wrapper: {
        backgroundColor: '#ffffff',
        borderTop: '1px solid #e5e7eb',
        padding: '48px 0',
        marginTop: 'auto',
    },
    inner: {
        maxWidth: '1280px',
        margin: '0 auto',
        padding: '0 16px',
    },
    grid: {
        display: 'grid',
        gridTemplateColumns: 'repeat(4, 1fr)',
        gap: '32px',
        marginBottom: '32px',
    },
    heading: {
        fontWeight: 700,
        fontSize: '12px',
        color: '#111827',
        textTransform: 'uppercase' as const,
        letterSpacing: '0.05em',
        marginBottom: '16px',
    },
    link: {
        fontSize: '13px',
        color: '#6b7280',
        textDecoration: 'none',
        display: 'block',
        marginBottom: '8px',
        transition: 'color 0.2s',
        cursor: 'pointer',
    },
    bottom: {
        borderTop: '1px solid #e5e7eb',
        paddingTop: '32px',
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center',
        flexWrap: 'wrap' as const,
        gap: '16px',
    },
    copyright: {
        fontSize: '13px',
        color: '#6b7280',
    },
    status: {
        display: 'flex',
        alignItems: 'center',
        gap: '8px',
        fontSize: '13px',
        color: '#6b7280',
    },
    statusDot: {
        width: '8px',
        height: '8px',
        borderRadius: '50%',
        backgroundColor: '#10b981',
    },
};

export const PortalLayout: React.FC<PortalLayoutProps> = ({ children }) => {
    const location = useLocation();
    const navigate = useNavigate();
    const logout = usePortalAuthStore(state => state.logout);
    const profile = usePortalAuthStore(state => state.profile);
    const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);
    const [isProfileOpen, setIsProfileOpen] = useState(false);
    const [scrolled, setScrolled] = useState(false);
    const profileRef = useRef<HTMLDivElement>(null);

    const isActive = (path: string) => location.pathname === path;

    const handleLogout = () => {
        logout();
        navigate('/login');
    };

    // Navbar shadow on scroll
    useEffect(() => {
        const onScroll = () => setScrolled(window.scrollY > 8);
        window.addEventListener('scroll', onScroll, { passive: true });
        return () => window.removeEventListener('scroll', onScroll);
    }, []);

    // Close profile dropdown on outside click
    useEffect(() => {
        const handler = (e: MouseEvent) => {
            if (profileRef.current && !profileRef.current.contains(e.target as Node)) {
                setIsProfileOpen(false);
            }
        };
        document.addEventListener('mousedown', handler);
        return () => document.removeEventListener('mousedown', handler);
    }, []);

    const navItems = [
        { path: '/dashboard', label: 'Dashboard', icon: LayoutDashboard },
        { path: '/requests', label: 'My Requests', icon: FileText },
        { path: '/notifications/breach', label: 'Notifications', icon: Bell },
        { path: '/profile', label: 'Profile', icon: User },
    ];

    return (
        <div style={{ minHeight: '100vh', display: 'flex', flexDirection: 'column', backgroundColor: '#F9FAFB', fontFamily: "var(--font-sans, 'Inter', sans-serif)", color: '#111827' }}>
            {/* ── Navigation Bar ── */}
            <nav style={{
                backgroundColor: scrolled ? 'rgba(255,255,255,0.95)' : 'rgba(255,255,255,0.9)',
                backdropFilter: 'blur(16px)',
                borderBottom: '1px solid #e5e7eb',
                position: 'sticky',
                top: 0,
                zIndex: 50,
                transition: 'box-shadow 0.3s, background-color 0.3s',
                boxShadow: scrolled ? '0 1px 3px rgba(0,0,0,0.06)' : 'none',
            }}>
                <div style={{ maxWidth: '1280px', margin: '0 auto', padding: '0 16px' }}>
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', height: '64px' }}>
                        {/* Branding & Desktop Nav */}
                        <div style={{ display: 'flex', alignItems: 'center', gap: '32px' }}>
                            <button
                                onClick={() => navigate('/dashboard')}
                                aria-label="Go to dashboard"
                                style={{ display: 'flex', alignItems: 'center', gap: '8px', background: 'none', border: 'none', cursor: 'pointer', padding: 0, fontFamily: 'inherit' }}
                            >
                                <div style={{
                                    background: 'linear-gradient(135deg, #2563eb 0%, #4f46e5 100%)',
                                    padding: '6px',
                                    borderRadius: '8px',
                                    color: '#fff',
                                    display: 'flex',
                                    boxShadow: '0 4px 6px rgba(59,130,246,0.2)',
                                }}>
                                    <Shield size={20} />
                                </div>
                                <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'flex-start', lineHeight: 1 }}>
                                    <span style={{ fontSize: '17px', fontWeight: 700, color: '#111827', letterSpacing: '-0.02em' }}>DataLens</span>
                                    <span style={{ fontSize: '10px', fontWeight: 600, color: '#3b82f6', textTransform: 'uppercase', letterSpacing: '0.05em' }}>Portal</span>
                                </div>
                            </button>

                            <div className="hidden md:flex" style={{ alignItems: 'center', gap: '4px' }}>
                                {navItems.map((item) => (
                                    <Link
                                        key={item.path}
                                        to={item.path}
                                        style={{
                                            display: 'inline-flex',
                                            alignItems: 'center',
                                            gap: '8px',
                                            padding: '8px 12px',
                                            borderRadius: '8px',
                                            fontSize: '13px',
                                            fontWeight: 500,
                                            textDecoration: 'none',
                                            transition: 'all 0.2s',
                                            color: isActive(item.path) ? '#1d4ed8' : '#6b7280',
                                            backgroundColor: isActive(item.path) ? 'rgba(59,130,246,0.08)' : 'transparent',
                                        }}
                                    >
                                        <item.icon size={16} style={{ color: isActive(item.path) ? '#3b82f6' : '#9ca3af' }} />
                                        {item.label}
                                    </Link>
                                ))}
                            </div>
                        </div>

                        {/* User Menu (Desktop) */}
                        <div className="hidden md:flex" style={{ alignItems: 'center', gap: '16px' }}>
                            <Link
                                to="/notifications/breach"
                                style={{
                                    padding: '8px',
                                    borderRadius: '50%',
                                    color: '#6b7280',
                                    display: 'flex',
                                    alignItems: 'center',
                                    justifyContent: 'center',
                                    transition: 'background-color 0.2s, color 0.2s',
                                    position: 'relative'
                                }}
                                className="hover:bg-slate-100 hover:text-slate-900"
                            >
                                <Bell size={20} />
                                <span style={{
                                    position: 'absolute',
                                    top: '6px',
                                    right: '6px',
                                    width: '8px',
                                    height: '8px',
                                    backgroundColor: '#ef4444',
                                    borderRadius: '50%',
                                    border: '1.5px solid #fff'
                                }} />
                            </Link>
                            <div style={{ height: '24px', width: '1px', backgroundColor: '#e5e7eb', margin: '0 4px' }} />
                            <div style={{ position: 'relative' }} ref={profileRef}>
                                <button
                                    onClick={() => setIsProfileOpen(!isProfileOpen)}
                                    style={{
                                        display: 'flex', alignItems: 'center', gap: '8px',
                                        padding: '4px 8px 4px 4px', borderRadius: '999px',
                                        border: 'none', backgroundColor: 'transparent',
                                        cursor: 'pointer', fontFamily: 'inherit',
                                    }}
                                    aria-label="Open user menu"
                                >
                                    <div style={{
                                        width: '32px', height: '32px', borderRadius: '50%',
                                        backgroundColor: '#dbeafe',
                                        display: 'flex', alignItems: 'center', justifyContent: 'center',
                                        color: '#3b82f6', fontWeight: 700, fontSize: '13px',
                                    }}>
                                        {profile?.email?.[0].toUpperCase() || 'U'}
                                    </div>
                                    <div className="hidden sm:block" style={{ textAlign: 'left' }}>
                                        <p style={{ fontSize: '13px', fontWeight: 600, color: '#111827', margin: 0, lineHeight: 1.3 }}>
                                            {profile?.email?.split('@')[0] || 'User'}
                                        </p>
                                        <p style={{ fontSize: '11px', color: '#6b7280', margin: 0 }}>Data Principal</p>
                                    </div>
                                    <ChevronDown size={16} style={{ color: '#6b7280', transition: 'transform 0.3s', transform: isProfileOpen ? 'rotate(180deg)' : 'none' }} />
                                </button>

                                {isProfileOpen && (
                                    <div className="animate-slide-down" style={{
                                        position: 'absolute', right: 0, marginTop: '12px',
                                        width: '256px', backgroundColor: '#fff',
                                        borderRadius: '16px', boxShadow: '0 20px 25px -5px rgba(0,0,0,0.1)',
                                        border: '1px solid #f3f4f6', padding: '8px', zIndex: 50,
                                    }}>
                                        <div style={{ padding: '12px 16px', backgroundColor: '#f9fafb', borderRadius: '12px', marginBottom: '4px' }}>
                                            <p style={{ fontSize: '13px', fontWeight: 600, color: '#111827', margin: 0 }}>{profile?.email}</p>
                                            <div style={{ display: 'flex', alignItems: 'center', gap: '6px', marginTop: '4px' }}>
                                                <div style={{ width: '6px', height: '6px', borderRadius: '50%', backgroundColor: '#10b981' }} />
                                                <p style={{ fontSize: '11px', color: '#6b7280', margin: 0 }}>Verified Account</p>
                                            </div>
                                        </div>
                                        <Link
                                            to="/profile"
                                            onClick={() => setIsProfileOpen(false)}
                                            style={{ display: 'flex', alignItems: 'center', gap: '10px', padding: '10px 12px', fontSize: '13px', color: '#374151', textDecoration: 'none', borderRadius: '10px', transition: 'background-color 0.2s' }}
                                        >
                                            <User size={16} style={{ color: '#9ca3af' }} /> Your Profile
                                        </Link>
                                        <Link
                                            to="/history"
                                            onClick={() => setIsProfileOpen(false)}
                                            style={{ display: 'flex', alignItems: 'center', gap: '10px', padding: '10px 12px', fontSize: '13px', color: '#374151', textDecoration: 'none', borderRadius: '10px', transition: 'background-color 0.2s' }}
                                        >
                                            <Shield size={16} style={{ color: '#9ca3af' }} /> Consent History
                                        </Link>
                                        <div style={{ borderTop: '1px solid #f3f4f6', margin: '4px 8px' }} />
                                        <button
                                            onClick={handleLogout}
                                            style={{ display: 'flex', alignItems: 'center', gap: '10px', padding: '10px 12px', fontSize: '13px', color: '#dc2626', width: '100%', border: 'none', borderRadius: '10px', backgroundColor: 'transparent', cursor: 'pointer', fontFamily: 'inherit', transition: 'background-color 0.2s' }}
                                        >
                                            <LogOut size={16} style={{ color: '#ef4444' }} /> Sign out
                                        </button>
                                    </div>
                                )}
                            </div>
                        </div>

                        {/* Mobile menu button */}
                        <div className="flex items-center md:hidden">
                            <button
                                onClick={() => setIsMobileMenuOpen(!isMobileMenuOpen)}
                                style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', padding: '8px', borderRadius: '8px', border: 'none', backgroundColor: 'transparent', cursor: 'pointer', color: '#6b7280' }}
                                aria-label="Toggle menu"
                            >
                                {isMobileMenuOpen ? <X size={24} /> : <Menu size={24} />}
                            </button>
                        </div>
                    </div>
                </div>

                {/* Mobile Menu */}
                {isMobileMenuOpen && (
                    <div className="md:hidden animate-slide-down" style={{
                        backgroundColor: 'rgba(255,255,255,0.98)',
                        backdropFilter: 'blur(16px)',
                        borderBottom: '1px solid #e5e7eb',
                        position: 'absolute', width: '100%', zIndex: 40,
                        boxShadow: '0 20px 25px -5px rgba(0,0,0,0.1)',
                    }}>
                        <div style={{ padding: '16px', display: 'flex', flexDirection: 'column', gap: '8px' }}>
                            {navItems.map((item) => (
                                <Link
                                    key={item.path}
                                    to={item.path}
                                    onClick={() => setIsMobileMenuOpen(false)}
                                    style={{
                                        display: 'flex', alignItems: 'center', gap: '12px',
                                        padding: '14px 16px', borderRadius: '12px',
                                        fontSize: '13px', fontWeight: 500, textDecoration: 'none',
                                        color: isActive(item.path) ? '#1d4ed8' : '#374151',
                                        backgroundColor: isActive(item.path) ? 'rgba(59,130,246,0.08)' : 'transparent',
                                    }}
                                >
                                    <item.icon size={20} style={{ color: isActive(item.path) ? '#3b82f6' : '#9ca3af' }} />
                                    {item.label}
                                </Link>
                            ))}
                        </div>
                        <div style={{ padding: '16px', backgroundColor: '#f9fafb', borderTop: '1px solid #e5e7eb' }}>
                            <div style={{ display: 'flex', alignItems: 'center', gap: '12px', marginBottom: '16px', padding: '12px', backgroundColor: '#fff', borderRadius: '12px', border: '1px solid #f3f4f6' }}>
                                <div style={{
                                    width: '40px', height: '40px', borderRadius: '50%',
                                    background: 'linear-gradient(135deg, #3b82f6, #4f46e5)',
                                    display: 'flex', alignItems: 'center', justifyContent: 'center',
                                    color: '#fff', fontWeight: 700,
                                }}>
                                    {profile?.email?.[0].toUpperCase() || 'U'}
                                </div>
                                <div>
                                    <div style={{ fontSize: '13px', fontWeight: 600, color: '#111827' }}>{profile?.email}</div>
                                    <div style={{ fontSize: '11px', color: '#6b7280', display: 'flex', alignItems: 'center', gap: '4px' }}>
                                        <span style={{ width: '6px', height: '6px', borderRadius: '50%', backgroundColor: '#10b981', display: 'inline-block' }} />
                                        Data Principal
                                    </div>
                                </div>
                            </div>
                            <button
                                onClick={handleLogout}
                                style={{
                                    width: '100%', display: 'flex', alignItems: 'center', justifyContent: 'center',
                                    gap: '8px', padding: '12px', fontSize: '13px', fontWeight: 500,
                                    color: '#dc2626', backgroundColor: '#fef2f2', border: 'none',
                                    borderRadius: '12px', cursor: 'pointer', fontFamily: 'inherit',
                                }}
                            >
                                <LogOut size={16} /> Sign out
                            </button>
                        </div>
                    </div>
                )}
            </nav>

            {/* ── Main Content ── */}
            <main style={{ flexGrow: 1, width: '100%', position: 'relative' }}>
                <div style={{ maxWidth: '1280px', margin: '0 auto', padding: '32px 16px 80px', position: 'relative', zIndex: 1 }}>
                    <div className="animate-fade-in">
                        {children}
                    </div>
                </div>
            </main>

            {/* ── Footer ── */}
            <footer style={footerStyles.wrapper}>
                <div style={footerStyles.inner}>
                    <div style={footerStyles.grid} className="portal-footer-grid">
                        {/* Brand column */}
                        <div>
                            <div style={{ display: 'flex', alignItems: 'center', gap: '8px', marginBottom: '16px' }}>
                                <div style={{ backgroundColor: '#3b82f6', color: '#fff', padding: '4px', borderRadius: '6px', display: 'flex' }}>
                                    <Shield size={14} />
                                </div>
                                <span style={{ fontWeight: 700, fontSize: '17px', color: '#111827' }}>DataLens</span>
                            </div>
                            <p style={{ fontSize: '13px', color: '#6b7280', lineHeight: 1.7 }}>
                                Empowering individuals to take control of their digital privacy. Secure, transparent, and compliant.
                            </p>
                        </div>

                        {/* Platform links */}
                        <div>
                            <h4 style={footerStyles.heading}>Platform</h4>
                            {['Dashboard', 'Consent History', 'Requests', 'Profile'].map(label => (
                                <a key={label} href="#" style={footerStyles.link}>{label}</a>
                            ))}
                        </div>

                        {/* Legal links */}
                        <div>
                            <h4 style={footerStyles.heading}>Legal</h4>
                            {['Privacy Policy', 'Terms of Service', 'Cookie Policy', 'Security'].map(label => (
                                <a key={label} href="#" style={footerStyles.link}>{label}</a>
                            ))}
                        </div>

                        {/* Connect */}
                        <div>
                            <h4 style={footerStyles.heading}>Connect</h4>
                            <div style={{ display: 'flex', gap: '16px' }}>
                                {/* Twitter/X */}
                                <a href="#" style={{ color: '#6b7280', transition: 'color 0.2s' }}>
                                    <svg width="24" height="24" fill="currentColor" viewBox="0 0 24 24">
                                        <path d="M8.29 20.251c7.547 0 11.675-6.253 11.675-11.675 0-.178 0-.355-.012-.53A8.348 8.348 0 0022 5.92a8.19 8.19 0 01-2.357.646 4.118 4.118 0 001.804-2.27 8.224 8.224 0 01-2.605.996 4.107 4.107 0 00-6.993 3.743 11.65 11.65 0 01-8.457-4.287 4.106 4.106 0 001.27 5.477A4.072 4.072 0 012.8 9.713v.052a4.105 4.105 0 003.292 4.022 4.095 4.095 0 01-1.853.07 4.108 4.108 0 003.834 2.85A8.233 8.233 0 012 18.407a11.616 11.616 0 006.29 1.84" />
                                    </svg>
                                </a>
                                {/* GitHub */}
                                <a href="#" style={{ color: '#6b7280', transition: 'color 0.2s' }}>
                                    <svg width="24" height="24" fill="currentColor" viewBox="0 0 24 24">
                                        <path fillRule="evenodd" clipRule="evenodd" d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.531 1.032 1.531 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0112 6.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z" />
                                    </svg>
                                </a>
                                {/* LinkedIn */}
                                <a href="#" style={{ color: '#6b7280', transition: 'color 0.2s' }}>
                                    <svg width="24" height="24" fill="currentColor" viewBox="0 0 24 24">
                                        <path fillRule="evenodd" clipRule="evenodd" d="M19 0h-14c-2.761 0-5 2.239-5 5v14c0 2.761 2.239 5 5 5h14c2.762 0 5-2.239 5-5v-14c0-2.761-2.238-5-5-5zm-11 19h-3v-11h3v11zm-1.5-12.268c-.966 0-1.75-.79-1.75-1.764s.784-1.764 1.75-1.764 1.75.79 1.75 1.764-.783 1.764-1.75 1.764zm13.5 12.268h-3v-5.604c0-3.368-4-3.113-4 0v5.604h-3v-11h3v1.765c1.396-2.586 7-2.777 7 2.476v6.759z" />
                                    </svg>
                                </a>
                            </div>
                        </div>
                    </div>

                    {/* Bottom bar */}
                    <div style={footerStyles.bottom}>
                        <p style={footerStyles.copyright}>© {new Date().getFullYear()} DataLens. All rights reserved.</p>
                        <div style={footerStyles.status}>
                            <span style={footerStyles.statusDot} />
                            <span>System Operational</span>
                        </div>
                    </div>
                </div>
            </footer>
        </div>
    );
};
