import React, { memo } from 'react';
import { Handle, Position, type NodeProps } from 'reactflow';
import { Database, Server, Globe } from 'lucide-react';

const icons: Record<string, React.ElementType> = {
    DATA_SOURCE: Database,
    PROCESS: Server,
    THIRD_PARTY: Globe,
};

const colors: Record<string, string> = {
    DATA_SOURCE: 'border-blue-500 bg-blue-50',
    PROCESS: 'border-purple-500 bg-purple-50',
    THIRD_PARTY: 'border-orange-500 bg-orange-50',
};

const CustomNode = ({ data }: NodeProps) => {
    const Icon = icons[data.type] || Database;
    const colorClass = colors[data.type] || 'border-gray-500 bg-gray-50';

    return (
        <div className={`px-4 py-2 shadow-md rounded-md border-2 ${colorClass} min-w-[150px]`}>
            <Handle type="target" position={Position.Top} className="w-16 !bg-gray-400" />
            <div className="flex items-center">
                <div className="rounded-full bg-white p-1 border mr-2">
                    <Icon className="w-4 h-4 text-gray-700" />
                </div>
                <div className="text-sm font-bold text-gray-900">{data.label}</div>
            </div>
            <div className="text-xs text-gray-500 mt-1">{data.subLabel}</div>
            <Handle type="source" position={Position.Bottom} className="w-16 !bg-gray-400" />
        </div>
    );
};

export default memo(CustomNode);
