import { ResponsiveContainer, BarChart, Bar, XAxis, YAxis, Tooltip, Cell } from 'recharts';
import { Loader2, PieChart } from 'lucide-react';

interface PIIChartProps {
    data: Record<string, number>;
    loading?: boolean;
}

const COLORS = ['#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', '#ec4899'];

export const PIIChart = ({ data, loading }: PIIChartProps) => {
    if (loading) {
        return (
            <div className="h-[300px] w-full flex items-center justify-center bg-gray-50 rounded-lg border border-dashed border-gray-300">
                <Loader2 className="animate-spin text-gray-400" size={32} />
            </div>
        );
    }

    const chartData = Object.entries(data)
        .map(([name, value]) => ({ name, value }))
        .sort((a, b) => b.value - a.value)
        .slice(0, 8); // Top 8 categories

    if (chartData.length === 0) {
        return (
            <div className="h-[300px] w-full flex flex-col items-center justify-center bg-gray-50 rounded-lg border border-dashed border-gray-300 text-gray-500 gap-2">
                <div className="p-3 bg-white rounded-full shadow-sm">
                    <PieChart className="text-gray-400" size={24} />
                </div>
                <p className="font-medium">No PII data found yet</p>
                <p className="text-xs text-gray-400">Scan results will appear here</p>
            </div>
        );
    }

    return (
        <div className="h-[300px] w-full">
            <ResponsiveContainer width="100%" height="100%">
                <BarChart data={chartData} layout="vertical" margin={{ top: 5, right: 30, left: 40, bottom: 5 }}>
                    <XAxis type="number" hide />
                    <YAxis
                        type="category"
                        dataKey="name"
                        tick={{ fontSize: 12, fill: '#6b7280' }}
                        width={100}
                        tickLine={false}
                        axisLine={false}
                    />
                    <Tooltip
                        cursor={{ fill: '#f3f4f6' }}
                        contentStyle={{ borderRadius: '8px', border: 'none', boxShadow: '0 4px 6px -1px rgb(0 0 0 / 0.1)' }}
                    />
                    <Bar dataKey="value" radius={[0, 4, 4, 0]} barSize={20}>
                        {chartData.map((_, index) => (
                            <Cell key={`cell - ${index} `} fill={COLORS[index % COLORS.length]} />
                        ))}
                    </Bar>
                </BarChart>
            </ResponsiveContainer>
        </div>
    );
};
