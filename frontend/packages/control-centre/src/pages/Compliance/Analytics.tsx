import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { analyticsService } from '../../services/analytics';
import {
    LineChart,
    Line,
    XAxis,
    YAxis,
    CartesianGrid,
    Tooltip,
    Legend,
    ResponsiveContainer,
    BarChart,
    Bar,
} from 'recharts';
// Card import removed
// Button import removed

// Fallback for Card components if they don't exist yet, I'll check existence first or just build them inline if simple
// Checking previous file list, I didn't see a 'Card' component in common. Retrieve list of components/common first?
// Actually, I'll stick to standard div/classes matching the design system I see in other files.
// Let's assume standard HTML/CSS structure for now to avoid dependency issues, or check common again.

const Analytics = () => {
    const [dateRange, setDateRange] = useState<'7d' | '30d' | '90d'>('30d');

    // Calculate dates based on range
    const getDates = (range: '7d' | '30d' | '90d') => {
        const end = new Date();
        const start = new Date();
        start.setDate(end.getDate() - (range === '7d' ? 7 : range === '30d' ? 30 : 90));
        return {
            from: start.toISOString().split('T')[0],
            to: end.toISOString().split('T')[0],
        };
    };

    const period = getDates(dateRange);

    const { data: conversionData, isLoading: isLoadingConversion, isError: isErrorConversion } = useQuery({
        queryKey: ['analytics-conversion', dateRange],
        queryFn: () => analyticsService.getConversionStats({ ...period, interval: 'day' }),
    });

    const { data: purposeData, isLoading: isLoadingPurpose, isError: isErrorPurpose } = useQuery({
        queryKey: ['analytics-purpose', dateRange],
        queryFn: () => analyticsService.getPurposeStats(period),
    });

    return (
        <div className="p-6 space-y-6">
            <div className="flex justify-between items-center">
                <div>
                    <h1 className="text-2xl font-bold text-gray-900">Consent Analytics</h1>
                    <p className="text-gray-500">Monitor consent acquisition and purpose distribution</p>
                </div>
                <div className="flex gap-2 bg-white p-1 rounded-lg border border-gray-200 shadow-sm">
                    {(['7d', '30d', '90d'] as const).map((range) => (
                        <button
                            key={range}
                            onClick={() => setDateRange(range)}
                            className={`px-3 py-1.5 text-sm font-medium rounded-md transition-colors ${dateRange === range
                                ? 'bg-blue-50 text-blue-700'
                                : 'text-gray-600 hover:bg-gray-50'
                                }`}
                        >
                            Last {range.replace('d', ' Days')}
                        </button>
                    ))}
                </div>
            </div>

            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
                {/* Conversion Trend Chart */}
                {/* Conversion Trend Chart */}
                <div className="bg-white p-6 rounded-xl border border-gray-200 shadow-sm">
                    <div className="mb-6">
                        <h3 className="text-lg font-semibold text-gray-900">Opt-In Rate Trend</h3>
                        <p className="text-sm text-gray-500">Daily conversion performance over time</p>
                    </div>
                    <div className="h-[300px] w-full">
                        {isLoadingConversion ? (
                            <div className="h-full flex items-center justify-center text-gray-400">Loading...</div>
                        ) : isErrorConversion ? (
                            <div className="h-full flex items-center justify-center text-red-500">Failed to load data</div>
                        ) : (
                            <ResponsiveContainer width="100%" height="100%">
                                <LineChart data={conversionData || []}>
                                    <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="#E5E7EB" />
                                    <XAxis
                                        dataKey="date"
                                        tickFormatter={(val) => new Date(val).toLocaleDateString('en-US', { month: 'short', day: 'numeric' })}
                                        stroke="#9CA3AF"
                                        fontSize={12}
                                        tickLine={false}
                                        axisLine={false}
                                    />
                                    <YAxis
                                        stroke="#9CA3AF"
                                        fontSize={12}
                                        tickLine={false}
                                        axisLine={false}
                                        tickFormatter={(val) => `${val}%`}
                                    />
                                    <Tooltip
                                        contentStyle={{ borderRadius: '8px', border: 'none', boxShadow: '0 4px 6px -1px rgb(0 0 0 / 0.1)' }}
                                    />
                                    <Legend />
                                    <Line
                                        type="monotone"
                                        dataKey="opt_in_count"
                                        name="Opt-In"
                                        stroke="#10B981"
                                        strokeWidth={2}
                                        dot={false}
                                        activeDot={{ r: 4 }}
                                    />
                                    <Line
                                        type="monotone"
                                        dataKey="opt_out_count"
                                        name="Opt-Out"
                                        stroke="#EF4444"
                                        strokeWidth={2}
                                        dot={false}
                                        activeDot={{ r: 4 }}
                                    />
                                </LineChart>
                            </ResponsiveContainer>
                        )}
                    </div>
                </div>

                {/* Purpose Breakdown Chart */}
                <div className="bg-white p-6 rounded-xl border border-gray-200 shadow-sm">
                    <div className="mb-6">
                        <h3 className="text-lg font-semibold text-gray-900">Purpose Breakdown</h3>
                        <p className="text-sm text-gray-500">Grant vs Deny rates by processing purpose</p>
                    </div>
                    <div className="h-[300px] w-full">
                        {isLoadingPurpose ? (
                            <div className="h-full flex items-center justify-center text-gray-400">Loading...</div>
                        ) : isErrorPurpose ? (
                            <div className="h-full flex items-center justify-center text-red-500">Failed to load data</div>
                        ) : (
                            <ResponsiveContainer width="100%" height="100%">
                                <BarChart data={purposeData || []} layout="vertical">
                                    <CartesianGrid strokeDasharray="3 3" horizontal={true} vertical={false} stroke="#E5E7EB" />
                                    <XAxis type="number" hide />
                                    <YAxis
                                        dataKey="purpose_name"
                                        type="category"
                                        width={100}
                                        tick={{ fontSize: 12, fill: '#4B5563' }}
                                        axisLine={false}
                                        tickLine={false}
                                    />
                                    <Tooltip cursor={{ fill: '#F3F4F6' }} />
                                    <Legend />
                                    <Bar dataKey="granted_count" name="Granted" fill="#10B981" radius={[0, 4, 4, 0]} stackId="a" />
                                    <Bar dataKey="denied_count" name="Denied" fill="#EF4444" radius={[0, 4, 4, 0]} stackId="a" />
                                </BarChart>
                            </ResponsiveContainer>
                        )}
                    </div>
                </div>
            </div>

            {/* Detailed Stats Table could go here eventually */}
        </div>
    );
};

export default Analytics;
