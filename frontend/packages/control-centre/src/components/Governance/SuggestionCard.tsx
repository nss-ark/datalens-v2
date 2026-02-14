import React from 'react';
import { Check, X, AlertCircle } from 'lucide-react';
import type { PurposeSuggestion } from '../../types/governance';

interface SuggestionCardProps {
    suggestion: PurposeSuggestion;
    onAccept: (id: string) => void;
    onReject: (id: string) => void;
}

export const SuggestionCard: React.FC<SuggestionCardProps> = ({
    suggestion,
    onAccept,
    onReject
}) => {
    const getConfidenceColor = (score: number) => {
        if (score >= 0.8) return 'text-green-500';
        if (score >= 0.5) return 'text-yellow-500';
        return 'text-red-500';
    };

    return (
        <div className="bg-white p-4 rounded-lg shadow-sm border border-gray-200 mb-4">
            <div className="flex justify-between items-start">
                <div>
                    <h4 className="text-md font-semibold text-gray-800">
                        {suggestion.dataSource} / {suggestion.table} / {suggestion.column}
                    </h4>
                    <p className="text-sm text-gray-500 mt-1">
                        Detected Data Element: <span className="font-medium text-gray-700">{suggestion.dataElement}</span>
                    </p>
                </div>
                <div className={`flex items-center ${getConfidenceColor(suggestion.confidenceScore)}`}>
                    <AlertCircle size={16} className="mr-1" />
                    <span className="text-sm font-bold">{Math.round(suggestion.confidenceScore * 100)}% Confidence</span>
                </div>
            </div>

            <div className="mt-4 p-3 bg-blue-50 rounded-md border border-blue-100">
                <p className="text-sm text-blue-800">
                    <span className="font-semibold">Suggested Purpose:</span> {suggestion.suggestedPurpose}
                </p>
                <p className="text-xs text-blue-600 mt-1">
                    Reasoning: {suggestion.reasoning}
                </p>
            </div>

            <div className="mt-4 flex justify-end space-x-2">
                <button
                    onClick={() => onReject(suggestion.id)}
                    className="flex items-center px-3 py-1.5 text-sm font-medium text-red-600 bg-red-50 hover:bg-red-100 rounded-md transition-colors"
                >
                    <X size={16} className="mr-1" />
                    Reject
                </button>
                <button
                    onClick={() => onAccept(suggestion.id)}
                    className="flex items-center px-3 py-1.5 text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 rounded-md transition-colors"
                >
                    <Check size={16} className="mr-1" />
                    Accept
                </button>
            </div>
        </div>
    );
};
