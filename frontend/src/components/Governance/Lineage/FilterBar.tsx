import React from 'react';
import { Filter } from 'lucide-react';

interface FilterBarProps {
    onFilterChange: (filters: Record<string, unknown>) => void;
}

const FilterBar: React.FC<FilterBarProps> = ({ onFilterChange }) => {
    return (
        <div className="absolute top-4 left-4 z-10 bg-white p-2 rounded-lg shadow-md border border-gray-200 flex items-center space-x-3">
            <div className="flex items-center text-gray-500 pr-3 border-r border-gray-200">
                <Filter size={16} className="mr-2" />
                <span className="text-sm font-medium">Filters</span>
            </div>

            <select
                className="text-sm border-none bg-transparent focus:ring-0 text-gray-700 font-medium cursor-pointer hover:bg-gray-50 rounded px-2 py-1 transition-colors"
                onChange={(e) => onFilterChange({ type: e.target.value })}
            >
                <option value="ALL">All Node Types</option>
                <option value="DATA_SOURCE">Data Sources</option>
                <option value="PROCESS">Processing</option>
                <option value="THIRD_PARTY">Third Parties</option>
            </select>

            <select
                className="text-sm border-none bg-transparent focus:ring-0 text-gray-700 font-medium cursor-pointer hover:bg-gray-50 rounded px-2 py-1 transition-colors"
                onChange={(e) => onFilterChange({ status: e.target.value })}
            >
                <option value="ALL">All Statuses</option>
                <option value="ACTIVE">Active Flows</option>
                <option value="PROPOSED">Proposed</option>
            </select>
        </div>
    );
};

export default FilterBar;
