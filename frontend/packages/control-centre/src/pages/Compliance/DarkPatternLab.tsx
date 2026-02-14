import { useState } from 'react';
import { useMutation } from '@tanstack/react-query';
import { AlertCircle, CheckCircle, Info, Play, RefreshCw, ShieldAlert, Code, Type, Link } from 'lucide-react';
import { Button } from '@datalens/shared';
import { StatusBadge } from '@datalens/shared';
import { darkPatternService } from '../../services/darkPatternService';
import type { DetectedPattern } from '../../types/darkPattern';
import { toast } from '@datalens/shared';

export default function DarkPatternLab() {
    const [content, setContent] = useState('');
    const [contentType, setContentType] = useState<'TEXT' | 'CODE' | 'HTML'>('TEXT');
    // const [activeTab, setActiveTab] = useState<'text' | 'code' | 'url'>('text'); // For UI tabs

    const { mutate: analyze, isPending, data: result, reset } = useMutation({
        mutationFn: darkPatternService.analyzeContent,
        onError: (error) => {
            console.error('Analysis failed:', error);
            toast.error('Analysis Failed', 'Could not analyze content. Please try again.');
        },
    });

    const handleAnalyze = () => {
        if (!content.trim()) {
            toast.error('Empty Content', 'Please enter some text or code to analyze.');
            return;
        }
        analyze({ content, type: contentType });
    };

    const getScoreColor = (score: number) => {
        if (score >= 90) return 'text-green-600 bg-green-50 border-green-200';
        if (score >= 70) return 'text-yellow-600 bg-yellow-50 border-yellow-200';
        return 'text-red-600 bg-red-50 border-red-200';
    };

    return (
        <div className="p-6 max-w-7xl mx-auto">
            <div className="mb-8">
                <h1 className="text-2xl font-bold text-gray-900 flex items-center gap-2">
                    <ShieldAlert className="w-8 h-8 text-indigo-600" />
                    Dark Pattern Lab
                </h1>
                <p className="text-gray-500 mt-1">
                    Test your UI text and code against India's "Guidelines for Prevention and Regulation of Dark Patterns, 2023".
                </p>
            </div>

            <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
                {/* Input Section */}
                <div className="flex flex-col gap-4">
                    <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-4">
                        <div className="flex flex-wrap border-b border-gray-100 mb-4 gap-4">
                            <button
                                onClick={() => setContentType('TEXT')}
                                className={`flex items-center gap-2 pb-2 text-sm font-medium transition-colors border-b-2 -mb-[1px] ${contentType === 'TEXT'
                                    ? 'text-indigo-600 border-indigo-600'
                                    : 'text-gray-500 border-transparent hover:text-gray-700 hover:border-gray-200'
                                    }`}
                            >
                                <Type className="w-4 h-4" />
                                Text Analysis
                            </button>
                            <button
                                onClick={() => setContentType('CODE')}
                                className={`flex items-center gap-2 pb-2 text-sm font-medium transition-colors border-b-2 -mb-[1px] ${contentType === 'CODE'
                                    ? 'text-indigo-600 border-indigo-600'
                                    : 'text-gray-500 border-transparent hover:text-gray-700 hover:border-gray-200'
                                    }`}
                            >
                                <Code className="w-4 h-4" />
                                Code / React
                            </button>
                            <button
                                disabled
                                className="flex items-center gap-2 pb-2 text-sm font-medium text-gray-300 cursor-not-allowed border-b-2 border-transparent -mb-[1px]"
                                title="URL scanning coming soon"
                            >
                                <Link className="w-4 h-4" />
                                URL (Coming Soon)
                            </button>
                        </div>

                        <textarea
                            value={content}
                            onChange={(e) => setContent(e.target.value)}
                            placeholder={
                                contentType === 'TEXT'
                                    ? "Paste marketing copy here (e.g., 'Hurry! Only 2 items left!')..."
                                    : "Paste React component code or HTML snippet here..."
                            }
                            className="w-full h-96 p-4 text-sm text-gray-700 bg-gray-50 border border-gray-200 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent resize-none font-mono"
                        />

                        <div className="mt-4 flex justify-between items-center">
                            <Button
                                variant="secondary"
                                onClick={() => {
                                    setContent('');
                                    reset();
                                }}
                                disabled={isPending || !content}
                                icon={<RefreshCw className="w-4 h-4" />}
                            >
                                Clear
                            </Button>
                            <Button
                                onClick={handleAnalyze}
                                disabled={isPending || !content}
                                isLoading={isPending}
                                icon={<Play className="w-4 h-4" />}
                            >
                                Analyze {contentType === 'TEXT' ? 'Text' : 'Code'}
                            </Button>
                        </div>
                    </div>

                    {/* Quick Examples */}
                    <div className="bg-indigo-50 rounded-lg p-4 border border-indigo-100">
                        <h4 className="text-sm font-semibold text-indigo-900 mb-2">Try these examples:</h4>
                        <div className="flex flex-wrap gap-2">
                            <button
                                onClick={() => {
                                    setContent("Confirm Shaming: 'No thanks, I prefer paying full price because I hate saving money.'");
                                    setContentType('TEXT');
                                }}
                                className="text-xs bg-white text-indigo-700 px-3 py-1 rounded-full border border-indigo-200 hover:bg-indigo-100 transition"
                            >
                                Confirm Shaming
                            </button>
                            <button
                                onClick={() => {
                                    setContent("False Urgency: 'Hurry! 15 other people are viewing this item right now!'");
                                    setContentType('TEXT');
                                }}
                                className="text-xs bg-white text-indigo-700 px-3 py-1 rounded-full border border-indigo-200 hover:bg-indigo-100 transition"
                            >
                                False Urgency
                            </button>
                            <button
                                onClick={() => {
                                    setContent("Basket Sneaking: 'We added insurance to your cart for your protection.'");
                                    setContentType('TEXT');
                                }}
                                className="text-xs bg-white text-indigo-700 px-3 py-1 rounded-full border border-indigo-200 hover:bg-indigo-100 transition"
                            >
                                Basket Sneaking
                            </button>
                        </div>
                    </div>
                </div>

                {/* Results Section */}
                <div className="flex flex-col gap-6">
                    {!result && !isPending && (
                        <div className="h-full flex flex-col items-center justify-center text-gray-400 bg-gray-50 rounded-xl border border-dashed border-gray-300 p-12">
                            <ShieldAlert className="w-16 h-16 mb-4 text-gray-300" />
                            <p className="text-lg font-medium">Ready to Analyze</p>
                            <p className="text-sm">Paste content and click Analyze to detect dark patterns.</p>
                        </div>
                    )}

                    {isPending && (
                        <div className="h-full flex flex-col items-center justify-center p-12">
                            <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-600 mb-4"></div>
                            <p className="text-gray-600">Analyzing patterns...</p>
                        </div>
                    )}

                    {result && (
                        <>
                            {/* Score Card */}
                            <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6 flex items-center justify-between">
                                <div>
                                    <h3 className="text-lg font-semibold text-gray-900">Compliance Score</h3>
                                    <p className="text-sm text-gray-500">Based on India 2023 Guidelines</p>
                                </div>
                                <div className={`flex items-center justify-center w-20 h-20 rounded-full border-4 text-2xl font-bold ${getScoreColor(result.overall_score)}`}>
                                    {result.overall_score}
                                </div>
                            </div>

                            {/* Detected Patterns */}
                            <div className="space-y-4">
                                <h3 className="text-lg font-semibold text-gray-900 flex items-center gap-2">
                                    Detected Issues
                                    <StatusBadge label={result.detected_patterns.length > 0 ? 'Review Needed' : 'PASS'} />
                                </h3>

                                {result.detected_patterns.length === 0 ? (
                                    <div className="bg-green-50 border border-green-200 rounded-lg p-6 flex flex-col items-center text-center">
                                        <CheckCircle className="w-12 h-12 text-green-500 mb-2" />
                                        <h4 className="font-semibold text-green-900">Clean! No Dark Patterns Detected.</h4>
                                        <p className="text-green-700 text-sm">Great job keeping your UI ethical and compliant.</p>
                                    </div>
                                ) : (
                                    result.detected_patterns.map((pattern: DetectedPattern, idx: number) => (
                                        <div key={idx} className="bg-white rounded-lg border border-red-200 shadow-sm overflow-hidden">
                                            <div className="bg-red-50 px-4 py-3 border-b border-red-100 flex justify-between items-center">
                                                <div className="flex items-center gap-2">
                                                    <AlertCircle className="w-5 h-5 text-red-600" />
                                                    <span className="font-semibold text-red-900">{pattern.pattern_name}</span>
                                                </div>
                                                <span className="text-xs font-mono bg-white px-2 py-1 rounded text-red-600 border border-red-200">
                                                    {pattern.regulation_clause}
                                                </span>
                                            </div>
                                            <div className="p-4">
                                                <p className="text-sm text-gray-700 mb-3">{pattern.description}</p>

                                                {pattern.location && (
                                                    <div className="mb-3">
                                                        <span className="text-xs font-semibold text-gray-500 uppercase tracking-wider">Location:</span>
                                                        <p className="text-sm font-mono text-gray-600 bg-gray-50 p-2 rounded mt-1 border border-gray-200">
                                                            {pattern.location}
                                                        </p>
                                                    </div>
                                                )}

                                                <div className="bg-indigo-50 rounded p-3 border border-indigo-100">
                                                    <div className="flex items-start gap-2">
                                                        <Info className="w-4 h-4 text-indigo-600 mt-0.5" />
                                                        <div>
                                                            <span className="text-xs font-bold text-indigo-700 uppercase">Suggestion</span>
                                                            <p className="text-sm text-indigo-900 mt-1">{pattern.fix_suggestion}</p>
                                                        </div>
                                                    </div>
                                                </div>
                                            </div>
                                        </div>
                                    ))
                                )}
                            </div>
                        </>
                    )}
                </div>
            </div>
        </div>
    );
}
