import { usePortalAuthStore } from '@/stores/portalAuthStore';
import { IdentityCard } from '@/components/IdentityCard';
import { StatsCard } from '@/components/StatsCard';
import { ActionCard } from '@/components/ActionCard';
import { ShieldCheck, FileText, User, Bell, Activity, Lock, ArrowRight } from 'lucide-react';
import { portalService } from '@/services/portalService';
import { useQuery } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';

const PortalDashboard = () => {
    const profile = usePortalAuthStore(state => state.profile);
    const navigate = useNavigate();

    const { data: consentSummary } = useQuery({
        queryKey: ['consents-summary'],
        queryFn: portalService.getConsentSummary
    });

    const activeConsents = consentSummary?.filter(c => c.status === 'GRANTED').length || 0;
    const firstName = profile?.email?.split('@')[0] || 'User';

    return (
        <div className="flex flex-col gap-8 animate-fade-in pb-10">
            {/* ── Hero Bento Section ── */}
            <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 w-full">
                {/* Welcome Card — spans 2 cols */}
                <div className="lg:col-span-2 relative bg-gradient-to-br from-slate-900 via-slate-800 to-slate-900 rounded-3xl p-8 md:p-10 overflow-hidden group shadow-xl">
                    {/* Animated gradient orbs */}
                    <div className="absolute top-0 right-0 w-80 h-80 bg-gradient-to-br from-blue-500/20 via-indigo-500/10 to-transparent rounded-full blur-3xl -mr-20 -mt-20 pointer-events-none group-hover:from-blue-500/30 transition-all duration-700" />
                    <div className="absolute bottom-0 left-0 w-64 h-64 bg-gradient-to-tr from-purple-500/15 via-blue-500/10 to-transparent rounded-full blur-3xl -ml-16 -mb-16 pointer-events-none" />

                    <div className="relative z-10 flex flex-col h-full justify-between">
                        <div>
                            <div className="inline-flex items-center gap-2 bg-white/10 backdrop-blur-md text-blue-200 text-xs font-semibold px-3 py-1.5 rounded-full mb-6 border border-white/10 shadow-sm">
                                <div className="w-1.5 h-1.5 bg-emerald-400 rounded-full animate-pulse shadow-[0_0_8px_rgba(52,211,153,0.6)]" />
                                Privacy Dashboard
                            </div>
                            <h1 className="text-4xl md:text-5xl font-extrabold text-white tracking-tight mb-4 leading-tight">
                                Hello, <span className="bg-clip-text text-transparent bg-gradient-to-r from-blue-400 to-indigo-400">{firstName}</span>
                            </h1>
                            <p className="text-slate-400 max-w-lg text-lg leading-relaxed font-medium">
                                Manage your data rights, review consents, and verify your identity — all in one place.
                            </p>
                        </div>

                        <div className="mt-8">
                            <button
                                onClick={() => navigate('/requests/new')}
                                className="inline-flex items-center gap-2 bg-white text-slate-900 px-6 py-3 rounded-2xl text-sm font-bold hover:bg-blue-50 transition-all duration-200 shadow-lg shadow-black/20 active:scale-[0.98] ring-1 ring-white/50"
                            >
                                Submit a Request
                                <ArrowRight className="w-4 h-4" />
                            </button>
                        </div>
                    </div>
                </div>

                {/* Privacy Score Card — KokonutUI Apple Activity style */}
                <div className="relative bg-gradient-to-br from-emerald-500 to-emerald-600 rounded-3xl p-8 text-white overflow-hidden group shadow-xl flex flex-col justify-between min-h-[300px]">
                    <div className="absolute -top-12 -right-12 w-48 h-48 bg-white/10 rounded-full blur-3xl pointer-events-none" />
                    <div className="absolute bottom-4 right-4 opacity-10 group-hover:opacity-20 transition-opacity duration-300 transform group-hover:scale-110 group-hover:-rotate-12">
                        <ShieldCheck className="w-32 h-32" />
                    </div>

                    <div className="relative z-10">
                        <div className="inline-flex items-center gap-1.5 bg-emerald-400/20 backdrop-blur-sm px-3 py-1 rounded-full text-emerald-50 text-xs font-medium mb-2 border border-emerald-400/30">
                            <Activity className="w-3 h-3" />
                            Live Status
                        </div>
                        <h2 className="text-6xl font-extrabold tracking-tighter mb-1 mt-2">Good</h2>
                        <p className="text-emerald-100 font-medium text-lg">Your Privacy Score</p>
                    </div>

                    <div className="relative z-10 bg-black/10 backdrop-blur-sm rounded-xl p-4 border border-white/10">
                        <div className="flex items-center gap-3">
                            <div className="w-10 h-10 rounded-full bg-emerald-400 flex items-center justify-center text-emerald-900 shadow-lg shadow-emerald-900/20">
                                <ShieldCheck className="w-6 h-6" />
                            </div>
                            <div>
                                <div className="text-sm font-bold text-white">All systems secure</div>
                                <div className="text-xs text-emerald-100/80">Last check: Just now</div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>

            {/* ── Stats + Identity Bento Row ── */}
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6 w-full stagger-children">
                <StatsCard
                    title="Active Consents"
                    value={activeConsents.toString()}
                    icon={Lock}
                    color="blue"
                    trend={{ value: '1 New', isPositive: true }}
                />
                <StatsCard
                    title="Open Requests"
                    value="0"
                    icon={FileText}
                    color="orange"
                />
                <StatsCard
                    title="Recent Activity"
                    value="2"
                    icon={Activity}
                    color="purple"
                />
                <div className="sm:col-span-2 lg:col-span-1 h-full min-h-[140px]">
                    <div className="bg-white rounded-2xl border border-slate-200/80 p-6 h-full flex flex-col justify-center hover:shadow-lg hover:border-slate-300/80 transition-all duration-300 relative overflow-hidden group">
                        <div className="absolute top-0 right-0 w-20 h-20 bg-slate-50 rounded-bl-full -mr-10 -mt-10 opacity-50 group-hover:scale-150 transition-transform duration-500" />

                        <p className="text-sm font-medium text-slate-500 mb-1.5 relative z-10">Verification</p>
                        <p className="text-2xl font-bold text-slate-900 relative z-10 mb-3">
                            {profile?.verification_status === 'VERIFIED' ? 'Verified' : 'Unverified'}
                        </p>

                        <button
                            onClick={() => navigate('/profile')}
                            className="mt-auto text-sm text-blue-600 font-semibold hover:text-blue-700 transition-colors flex items-center gap-1.5 group/btn"
                        >
                            View Profile <ArrowRight className="w-4 h-4 transition-transform group-hover/btn:translate-x-1" />
                        </button>
                    </div>
                </div>
            </div>

            {/* ── Identity + Quick Actions Grid ── */}
            <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 w-full">
                {/* Identity Card */}
                <div className="lg:col-span-2 h-full">
                    <IdentityCard />
                </div>

                {/* Mini Actions Column */}
                <div className="flex flex-col gap-6 h-full">
                    <div className="bg-gradient-to-br from-blue-500 to-indigo-600 rounded-2xl p-6 text-white relative overflow-hidden group shadow-lg flex-1 flex flex-col justify-center min-h-[180px]">
                        <div className="absolute -bottom-8 -right-8 w-32 h-32 bg-white/10 rounded-full blur-2xl pointer-events-none group-hover:scale-125 transition-transform duration-500" />
                        <div className="relative z-10">
                            <div className="flex justify-between items-start mb-4">
                                <div className="p-2 bg-white/10 rounded-lg backdrop-blur-sm">
                                    <Lock className="w-5 h-5 text-blue-100" />
                                </div>
                                <span className="text-xs font-bold bg-white/20 px-2 py-1 rounded text-blue-50">MANAGE</span>
                            </div>
                            <div className="flex-1 flex flex-col justify-center py-2">
                                <p className="text-blue-100 text-sm font-medium mb-1">Total Consents</p>
                                <h3 className="text-4xl font-extrabold tracking-tight">{consentSummary?.length || 0}</h3>
                            </div>
                            <button
                                onClick={() => navigate('/consent')}
                                className="mt-4 w-full bg-white/10 hover:bg-white/20 border border-white/10 text-white rounded-xl py-2 text-sm font-semibold transition-colors flex items-center justify-center gap-2"
                            >
                                Review All <ArrowRight className="w-3.5 h-3.5" />
                            </button>
                        </div>
                    </div>

                    <div className="bg-white rounded-2xl border border-slate-200/80 p-6 hover:shadow-lg hover:border-slate-300/80 transition-all duration-300 flex-1 flex flex-col justify-center group min-h-[180px]">
                        <div className="flex justify-between items-start mb-4">
                            <div className="p-2 bg-slate-50 rounded-lg group-hover:bg-blue-50 transition-colors">
                                <FileText className="w-5 h-5 text-slate-400 group-hover:text-blue-500 transition-colors" />
                            </div>
                        </div>
                        <p className="text-sm font-medium text-slate-500 mb-1">Data Requests</p>
                        <h3 className="text-4xl font-extrabold text-slate-900 mb-4">0</h3>
                        <button
                            onClick={() => navigate('/requests')}
                            className="mt-auto w-full bg-slate-50 hover:bg-slate-100 text-slate-600 hover:text-slate-900 rounded-xl py-2 text-sm font-semibold transition-colors border border-slate-200"
                        >
                            View History
                        </button>
                    </div>
                </div>
            </div>

            {/* ── Quick Actions ── */}
            <div className="pt-4">
                <div className="flex items-center justify-between mb-6">
                    <div>
                        <h2 className="text-2xl font-bold text-slate-900 tracking-tight">Quick Actions</h2>
                        <p className="text-slate-500 mt-1 font-medium">Exercise your privacy rights</p>
                    </div>
                </div>
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 stagger-children w-full">
                    <ActionCard
                        title="Submit Request"
                        description="Exercise your rights to access, correct, or erase your data."
                        icon={FileText}
                        to="/requests/new"
                        color="blue"
                    />
                    <ActionCard
                        title="Manage Consents"
                        description="Review and control which applications can access your data."
                        icon={Lock}
                        to="/consent"
                        color="indigo"
                    />
                    <ActionCard
                        title="My Profile"
                        description="Update your personal details and verification status."
                        icon={User}
                        to="/profile"
                        color="purple"
                    />
                    <ActionCard
                        title="Safety Alerts"
                        description="View breach notifications and security alerts."
                        icon={Bell}
                        to="/notifications/breach"
                        color="emerald"
                    />
                </div>
            </div>
        </div>
    );
};

export default PortalDashboard;
