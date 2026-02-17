import React, { useState, useEffect, useRef } from 'react';
import { Shield, LayoutDashboard, Bell, User, LogOut, Menu, X, ChevronDown, FileText } from 'lucide-react';
import { Link, useLocation, useNavigate } from 'react-router-dom';
import { usePortalAuthStore } from '@/stores/portalAuthStore';

interface PortalLayoutProps {
    children: React.ReactNode;
}

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
        <div className="min-h-screen bg-slate-50/80 flex flex-col font-sans text-slate-900 selection:bg-blue-100 selection:text-blue-900">
            {/* Navigation Bar */}
            <nav className={`bg-white/90 backdrop-blur-xl border-b border-slate-200/60 sticky top-0 z-50 transition-all duration-300 ${scrolled ? 'shadow-sm bg-white/95' : ''}`}>
                <div className="w-full max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
                    <div className="flex justify-between h-18">
                        {/* Branding & Desktop Nav */}
                        <div className="flex items-center gap-8 lg:gap-12">
                            <button
                                className="flex-shrink-0 flex items-center gap-3 group"
                                onClick={() => navigate('/dashboard')}
                                aria-label="Go to dashboard"
                            >
                                <div className="bg-gradient-to-br from-blue-600 to-indigo-600 p-2 text-white rounded-xl shadow-lg shadow-blue-500/20 group-hover:shadow-blue-500/30 transition-all duration-300 group-hover:scale-105">
                                    <Shield className="w-5 h-5" />
                                </div>
                                <div className="flex flex-col items-start leading-none">
                                    <span className="text-lg font-bold text-slate-900 tracking-tight">DataLens</span>
                                    <span className="text-[10px] font-semibold text-blue-600 uppercase tracking-wider">Portal</span>
                                </div>
                            </button>

                            <div className="hidden md:flex items-center gap-1">
                                {navItems.map((item) => (
                                    <Link
                                        key={item.path}
                                        to={item.path}
                                        className={`inline-flex items-center gap-2 px-3.5 py-2 rounded-lg text-sm font-medium transition-all duration-200 border border-transparent ${isActive(item.path)
                                            ? 'bg-blue-50/50 text-blue-700 border-blue-100/50 shadow-sm'
                                            : 'text-slate-500 hover:text-slate-900 hover:bg-slate-50/80'
                                            }`}
                                    >
                                        <item.icon className={`w-4 h-4 ${isActive(item.path) ? 'text-blue-600' : 'text-slate-400 group-hover:text-slate-600'}`} />
                                        {item.label}
                                    </Link>
                                ))}
                            </div>
                        </div>

                        {/* User Menu (Desktop) */}
                        <div className="hidden md:flex items-center gap-4">
                            <div className="h-8 w-[1px] bg-slate-200/80 mx-2" />
                            <div className="relative" ref={profileRef}>
                                <button
                                    onClick={() => setIsProfileOpen(!isProfileOpen)}
                                    className="flex items-center gap-3 pl-1 pr-2 py-1 rounded-full hover:bg-slate-50 transition-colors border border-transparent hover:border-slate-100"
                                    aria-label="Open user menu"
                                >
                                    <div className="h-9 w-9 rounded-full bg-gradient-to-br from-blue-100 to-indigo-50 border-2 border-white shadow-sm flex items-center justify-center text-blue-700 font-bold text-sm">
                                        {profile?.email?.[0].toUpperCase() || 'U'}
                                    </div>
                                    <div className="flex flex-col items-start hidden lg:flex">
                                        <span className="text-slate-700 font-semibold text-sm leading-tight max-w-[120px] truncate">
                                            {profile?.email?.split('@')[0] || 'User'}
                                        </span>
                                        <span className="text-[10px] text-slate-400 font-medium">Data Principal</span>
                                    </div>
                                    <ChevronDown className={`w-4 h-4 text-slate-400 transition-transform duration-300 ${isProfileOpen ? 'rotate-180 text-blue-600' : ''}`} />
                                </button>

                                {isProfileOpen && (
                                    <div className="absolute right-0 mt-3 w-64 bg-white rounded-2xl shadow-xl border border-slate-100 py-2 animate-slide-down origin-top-right z-50 ring-1 ring-black/5">
                                        <div className="px-5 py-4 bg-slate-50/50 border-b border-slate-100 mb-1">
                                            <p className="text-sm font-semibold text-slate-900 truncate">{profile?.email}</p>
                                            <div className="flex items-center gap-1.5 mt-1.5">
                                                <div className="w-1.5 h-1.5 rounded-full bg-emerald-500" />
                                                <p className="text-xs text-slate-500 font-medium">Verified Account</p>
                                            </div>
                                        </div>
                                        <div className="p-1.5">
                                            <Link
                                                to="/profile"
                                                className="flex items-center gap-3 px-3.5 py-2.5 text-sm text-slate-600 hover:bg-blue-50 hover:text-blue-700 rounded-xl transition-colors group"
                                                onClick={() => setIsProfileOpen(false)}
                                            >
                                                <div className="p-1.5 bg-slate-50 rounded-lg group-hover:bg-blue-100/50 transition-colors">
                                                    <User className="w-4 h-4 text-slate-400 group-hover:text-blue-600" />
                                                </div>
                                                Your Profile
                                            </Link>
                                            <Link
                                                to="/history"
                                                className="flex items-center gap-3 px-3.5 py-2.5 text-sm text-slate-600 hover:bg-blue-50 hover:text-blue-700 rounded-xl transition-colors group"
                                                onClick={() => setIsProfileOpen(false)}
                                            >
                                                <div className="p-1.5 bg-slate-50 rounded-lg group-hover:bg-blue-100/50 transition-colors">
                                                    <Shield className="w-4 h-4 text-slate-400 group-hover:text-blue-600" />
                                                </div>
                                                Consent History
                                            </Link>
                                        </div>
                                        <div className="border-t border-slate-100 mx-3 my-1.5" />
                                        <div className="p-1.5">
                                            <button
                                                onClick={handleLogout}
                                                className="w-full flex items-center gap-3 px-3.5 py-2.5 text-sm text-red-600 hover:bg-red-50 rounded-xl transition-colors group"
                                            >
                                                <div className="p-1.5 bg-red-50 rounded-lg group-hover:bg-red-100/50 transition-colors">
                                                    <LogOut className="w-4 h-4 text-red-500" />
                                                </div>
                                                Sign out
                                            </button>
                                        </div>
                                    </div>
                                )}
                            </div>
                        </div>

                        {/* Mobile menu button */}
                        <div className="flex items-center md:hidden">
                            <button
                                onClick={() => setIsMobileMenuOpen(!isMobileMenuOpen)}
                                className="inline-flex items-center justify-center p-2 rounded-xl text-slate-500 hover:text-slate-900 hover:bg-slate-100 transition-all border border-transparent hover:border-slate-200"
                                aria-label="Toggle menu"
                            >
                                {isMobileMenuOpen ? (
                                    <X className="h-6 w-6" />
                                ) : (
                                    <Menu className="h-6 w-6" />
                                )}
                            </button>
                        </div>
                    </div>
                </div>

                {/* Mobile Menu */}
                {isMobileMenuOpen && (
                    <div className="md:hidden bg-white/95 backdrop-blur-xl border-b border-slate-200 animate-slide-down shadow-xl absolute w-full z-40">
                        <div className="p-4 space-y-2">
                            {navItems.map((item) => (
                                <Link
                                    key={item.path}
                                    to={item.path}
                                    onClick={() => setIsMobileMenuOpen(false)}
                                    className={`flex items-center gap-3 px-4 py-3.5 rounded-xl text-sm font-medium transition-all ${isActive(item.path)
                                        ? 'bg-blue-50 text-blue-700 shadow-sm ring-1 ring-blue-100'
                                        : 'text-slate-600 hover:bg-slate-50 hover:text-slate-900'
                                        }`}
                                >
                                    <item.icon className={`w-5 h-5 ${isActive(item.path) ? 'text-blue-600' : 'text-slate-400'}`} />
                                    {item.label}
                                </Link>
                            ))}
                        </div>
                        <div className="p-4 bg-slate-50/50 border-t border-slate-200">
                            <div className="flex items-center gap-3 mb-4 p-3 bg-white rounded-xl border border-slate-100 shadow-sm">
                                <div className="h-10 w-10 rounded-full bg-gradient-to-br from-blue-500 to-indigo-600 flex items-center justify-center text-white font-bold shadow-md">
                                    {profile?.email?.[0].toUpperCase() || 'U'}
                                </div>
                                <div className="overflow-hidden">
                                    <div className="text-sm font-semibold text-slate-900 truncate">{profile?.email}</div>
                                    <div className="text-xs text-slate-500 flex items-center gap-1">
                                        <span className="w-1.5 h-1.5 rounded-full bg-emerald-500"></span>
                                        Data Principal
                                    </div>
                                </div>
                            </div>
                            <button
                                onClick={handleLogout}
                                className="w-full flex items-center justify-center gap-2 text-sm font-medium text-red-600 hover:text-red-700 bg-red-50 hover:bg-red-100/80 py-3 rounded-xl transition-colors"
                            >
                                <LogOut className="w-4 h-4" />
                                Sign out
                            </button>
                        </div>
                    </div>
                )}
            </nav>

            {/* Main Content */}
            <main className="flex-grow w-full relative">
                {/* Subtle background texture */}
                <div className="absolute inset-0 bg-[url('https://grainy-gradients.vercel.app/noise.svg')] opacity-20 pointer-events-none mix-blend-soft-light fixed"></div>

                <div className="w-full max-w-7xl mx-auto pt-8 pb-20 px-4 sm:px-6 lg:px-8 relative z-10">
                    <div className="animate-fade-in">
                        {children}
                    </div>
                </div>
            </main>

            {/* Footer */}
            <footer className="bg-white border-t border-slate-200/80 mt-auto relative z-20">
                <div className="w-full max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-10 flex flex-col md:flex-row justify-between items-center gap-6">
                    <div className="flex items-center gap-3 opacity-80 hover:opacity-100 transition-opacity">
                        <div className="bg-slate-100 p-1.5 rounded-lg">
                            <Shield className="w-4 h-4 text-slate-400" />
                        </div>
                        <div className="text-sm text-slate-400">
                            © {new Date().getFullYear()} <span className="text-slate-600 font-semibold">DataLens</span>
                            <span className="mx-2 text-slate-300">|</span>
                            Secure Privacy Infrastructure
                        </div>
                    </div>
                    <div className="flex gap-8 text-sm font-medium text-slate-500">
                        <a href="#" className="hover:text-blue-600 transition-colors">Privacy</a>
                        <a href="#" className="hover:text-blue-600 transition-colors">Terms</a>
                        <a href="#" className="hover:text-blue-600 transition-colors">Support</a>
                    </div>
                </div>
            </footer>
        </div>
    );
};
