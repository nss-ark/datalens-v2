import { Sparkles } from 'lucide-react';

export const StartupHero = () => {
    return (
        <div className="relative overflow-hidden bg-white/50 backdrop-blur-sm border-b border-slate-100 mb-8">
            <div className="absolute top-0 w-full h-full bg-[radial-gradient(ellipse_80%_80%_at_50%_-20%,rgba(120,119,198,0.1),rgba(255,255,255,0))]" />

            <div className="relative max-w-5xl mx-auto px-4 sm:px-6 lg:px-8 pt-12 pb-16 text-center">
                <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-blue-50 border border-blue-100 text-blue-600 text-xs font-medium mb-6 animate-fade-in-up">
                    <Sparkles className="w-3 h-3" />
                    <span>0</span>
                </div>

                <h1 className="text-4xl sm:text-5xl font-bold tracking-tight text-slate-900 mb-6 animate-fade-in-up [animation-delay:100ms]">
                    Manage your data privacy <br className="hidden sm:block" />
                    <span className="text-transparent bg-clip-text bg-gradient-to-r from-blue-600 to-indigo-600">
                        with confidence.
                    </span>
                </h1>

                <p className="text-lg text-slate-600 mb-8 max-w-2xl mx-auto animate-fade-in-up [animation-delay:200ms]">
                    Gain full visibility into how your data is used. Manage consents, track requests, and exercise your rights from a single, secure control center.
                </p>

                <div className="flex flex-col sm:flex-row justify-center gap-4 animate-fade-in-up [animation-delay:300ms]">
                    {/* Placeholder for action buttons if needed in future */}
                </div>
            </div>
        </div>
    );
};
