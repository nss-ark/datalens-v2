import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { lineageService } from '../../services/lineage';
import LineageGraph from '../../components/Governance/Lineage/LineageGraph';
import NodeDetailsPanel from '../../components/Governance/Lineage/NodeDetailsPanel';
import FilterBar from '../../components/Governance/Lineage/FilterBar';
import type { GraphNode } from '../../types/lineage';

const DataLineage = () => {
    const [selectedNode, setSelectedNode] = useState<GraphNode | null>(null);
    const [filters, setFilters] = useState({});

    const { data, isLoading, isError, error } = useQuery({
        queryKey: ['lineage-graph'],
        queryFn: lineageService.getGraph,
    });

    const handleNodeClick = (node: GraphNode) => {
        setSelectedNode(node);
    };

    const handleFilterChange = (newFilters: Record<string, unknown>) => {
        const updated = { ...filters, ...newFilters };
        setFilters(updated);
        console.log('Filters applied:', updated);
        // Note: Real filtering would happen either here (client-side) or via API refetch
        // For now, we'll implement client-side filtering logic if needed, but the graph keeps simple.
    };

    if (isLoading) {
        return (
            <div className="flex h-screen items-center justify-center bg-gray-50">
                <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
            </div>
        );
    }

    if (isError) {
        return <div className="p-4 text-red-500">Error loading lineage: {error ? (error as Error).message : 'Unknown error'}</div>;
    }

    // Prepare data
    const nodes = data?.nodes || [];
    const edges = data?.edges || [];

    return (
        <div className="h-[calc(100vh-64px)] relative flex flex-col">
            <div className="border-b bg-white px-6 py-4">
                <h1 className="text-2xl font-bold text-gray-900">Data Lineage</h1>
                <p className="text-sm text-gray-500">Visualize data flow across your organization</p>
            </div>

            <div className="flex-1 relative">
                <FilterBar onFilterChange={handleFilterChange} />

                {nodes.length > 0 ? (
                    <LineageGraph
                        initialNodes={nodes}
                        initialEdges={edges}
                        onNodeClick={handleNodeClick}
                    />
                ) : (
                    <div className="flex h-full items-center justify-center text-gray-400">
                        No lineage data available
                    </div>
                )}

                <NodeDetailsPanel
                    node={selectedNode}
                    onClose={() => setSelectedNode(null)}
                />
            </div>
        </div>
    );
};

export default DataLineage;
