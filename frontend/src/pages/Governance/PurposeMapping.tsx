import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Briefcase, CheckCircle } from 'lucide-react';
import { governanceService } from '../../services/governance';
import { SuggestionCard } from '../../components/Governance/SuggestionCard';
import { toast } from '../../stores/toastStore';
import { Button } from '../../components/ui/button';
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from '../../components/ui/select';
import type { PurposeSuggestion } from '../../types/governance';

const PurposeMapping = () => {
    const queryClient = useQueryClient();
    const [filter, setFilter] = useState<'all' | 'high_confidence'>('all');

    // Fetch suggestions
    const { data: suggestions = [], isLoading } = useQuery({
        queryKey: ['purposeSuggestions'],
        queryFn: governanceService.getPurposeSuggestions,
    });

    // Accept mutation
    const acceptMutation = useMutation({
        mutationFn: governanceService.acceptSuggestion,
        onSuccess: () => {
            toast.success('Purpose accepted', 'The purpose mapping has been updated.');
            queryClient.invalidateQueries({ queryKey: ['purposeSuggestions'] });
        },
        onError: () => {
            toast.error('Error', 'Failed to accept suggestion.');
        }
    });

    // Reject mutation
    const rejectMutation = useMutation({
        mutationFn: governanceService.rejectSuggestion,
        onSuccess: () => {
            toast.info('Suggestion rejected', 'The suggestion has been dismissed.');
            queryClient.invalidateQueries({ queryKey: ['purposeSuggestions'] });
        },
        onError: () => {
            toast.error('Error', 'Failed to reject suggestion.');
        }
    });

    const filteredSuggestions = suggestions.filter((s: PurposeSuggestion) => {
        if (filter === 'high_confidence') return s.confidenceScore >= 0.8;
        return true;
    });

    const handleAcceptAllHighConfidence = async () => {
        const highConfidence = suggestions.filter((s: PurposeSuggestion) => s.confidenceScore >= 0.8);
        if (highConfidence.length === 0) {
            toast.info('No high confidence suggestions', 'There are no suggestions with >80% confidence.');
            return;
        }

        try {
            await Promise.all(highConfidence.map((s: PurposeSuggestion) => governanceService.acceptSuggestion(s.id)));
            toast.success('Batch Accept Complete', `Accepted ${highConfidence.length} suggestions.`);
            queryClient.invalidateQueries({ queryKey: ['purposeSuggestions'] });
        } catch (error) {
            toast.error('Batch Error', 'Failed to accept some suggestions.');
            console.error(error);
        }
    };

    return (
        <div className="p-6 max-w-7xl mx-auto">
            <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4 mb-8">
                <div>
                    <h1 className="text-2xl font-bold text-gray-900 flex items-center gap-2">
                        <Briefcase className="text-blue-600" />
                        Purpose Mapping
                    </h1>
                    <p className="text-gray-500 mt-1">
                        Review AI suggestions for unmapped data elements.
                    </p>
                </div>
                <div className="flex flex-wrap items-center gap-3">
                    <Select
                        value={filter}
                        onValueChange={(val) => setFilter(val as 'all' | 'high_confidence')}
                    >
                        <SelectTrigger className="w-[180px]">
                            <SelectValue placeholder="Filter suggestions" />
                        </SelectTrigger>
                        <SelectContent>
                            <SelectItem value="all">All Suggestions</SelectItem>
                            <SelectItem value="high_confidence">High Confidence Only</SelectItem>
                        </SelectContent>
                    </Select>

                    <Button
                        onClick={handleAcceptAllHighConfidence}
                        variant="default"
                        className="bg-green-600 hover:bg-green-700 text-white"
                    >
                        <CheckCircle size={16} className="mr-2" />
                        Accept High Confidence
                    </Button>
                </div>
            </div>

            {isLoading ? (
                <div className="flex justify-center py-12">
                    <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
                </div>
            ) : filteredSuggestions.length === 0 ? (
                <div className="text-center py-12 bg-white rounded-lg shadow-sm border border-gray-200">
                    <Briefcase className="mx-auto h-12 w-12 text-gray-300" />
                    <h3 className="mt-2 text-sm font-medium text-gray-900">No suggestions found</h3>
                    <p className="mt-1 text-sm text-gray-500">
                        Good job! All data elements are currently mapped.
                    </p>
                </div>
            ) : (
                <div className="grid grid-cols-1 gap-6">
                    {filteredSuggestions.map((suggestion: PurposeSuggestion) => (
                        <SuggestionCard
                            key={suggestion.id}
                            suggestion={suggestion}
                            onAccept={(id) => acceptMutation.mutate(id)}
                            onReject={(id) => rejectMutation.mutate(id)}
                        />
                    ))}
                </div>
            )}
        </div>
    );
};

export default PurposeMapping;
