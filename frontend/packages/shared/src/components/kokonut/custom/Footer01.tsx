
import { Shield, Github, Twitter, Linkedin } from 'lucide-react';

export const Footer01 = () => {
    const currentYear = new Date().getFullYear();

    return (
        <footer className="bg-white border-t border-slate-100 py-12">
            <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
                <div className="grid grid-cols-1 md:grid-cols-4 gap-8 mb-8">
                    <div className="col-span-1 md:col-span-1">
                        <div className="flex items-center gap-2 mb-4">
                            <div className="h-8 w-8 bg-blue-600 rounded-lg flex items-center justify-center">
                                <Shield className="h-5 w-5 text-white" />
                            </div>
                            <span className="text-lg font-bold tracking-tight text-slate-900">DataLens</span>
                        </div>
                        <p className="text-sm text-slate-500 leading-relaxed">
                            Empowering individuals to take control of their digital privacy. Secure, transparent, and compliant.
                        </p>
                    </div>

                    <div>
                        <h3 className="text-sm font-semibold text-slate-900 uppercase tracking-wider mb-4">Platform</h3>
                        <ul className="space-y-3">
                            <li><a href="#" className="text-sm text-slate-500 hover:text-blue-600 transition-colors">Dashboard</a></li>
                            <li><a href="#" className="text-sm text-slate-500 hover:text-blue-600 transition-colors">Consent History</a></li>
                            <li><a href="#" className="text-sm text-slate-500 hover:text-blue-600 transition-colors">Requests</a></li>
                            <li><a href="#" className="text-sm text-slate-500 hover:text-blue-600 transition-colors">Profile</a></li>
                        </ul>
                    </div>

                    <div>
                        <h3 className="text-sm font-semibold text-slate-900 uppercase tracking-wider mb-4">Legal</h3>
                        <ul className="space-y-3">
                            <li><a href="#" className="text-sm text-slate-500 hover:text-blue-600 transition-colors">Privacy Policy</a></li>
                            <li><a href="#" className="text-sm text-slate-500 hover:text-blue-600 transition-colors">Terms of Service</a></li>
                            <li><a href="#" className="text-sm text-slate-500 hover:text-blue-600 transition-colors">Cookie Policy</a></li>
                            <li><a href="#" className="text-sm text-slate-500 hover:text-blue-600 transition-colors">Security</a></li>
                        </ul>
                    </div>

                    <div>
                        <h3 className="text-sm font-semibold text-slate-900 uppercase tracking-wider mb-4">Connect</h3>
                        <div className="flex space-x-4">
                            <a href="#" className="text-slate-400 hover:text-blue-600 transition-colors">
                                <Twitter className="h-5 w-5" />
                            </a>
                            <a href="#" className="text-slate-400 hover:text-blue-600 transition-colors">
                                <Github className="h-5 w-5" />
                            </a>
                            <a href="#" className="text-slate-400 hover:text-blue-600 transition-colors">
                                <Linkedin className="h-5 w-5" />
                            </a>
                        </div>
                    </div>
                </div>

                <div className="border-t border-slate-100 pt-8 flex flex-col md:flex-row justify-between items-center gap-4">
                    <p className="text-sm text-slate-500">
                        &copy; {currentYear} DataLens. All rights reserved.
                    </p>
                    <div className="flex gap-6">
                        <span className="flex items-center gap-2 text-sm text-slate-500">
                            <div className="w-2 h-2 rounded-full bg-green-500 animate-pulse" />
                            System Operational
                        </span>
                    </div>
                </div>
            </div>
        </footer>
    );
};
