import React, { useCallback, useEffect, useState } from 'react';
import ReactFlow, {
    Controls,
    Background,
    useNodesState,
    useEdgesState,
    MarkerType,
    ConnectionLineType,
    Panel,
    Position,
} from 'reactflow';
import type { Node, Edge } from 'reactflow';
import 'reactflow/dist/style.css';
import * as dagre from 'dagre';
import { useQuery } from '@tanstack/react-query';
import { governanceService } from '../../services/governance';
import type { LineageNode } from '../../types/governance';
import { StatusBadge } from '../../components/common/StatusBadge';
import { Database, Server, FileText, Globe, Layers, Cog } from 'lucide-react';

// --- Layout Helper ---
const dagreGraph = new dagre.graphlib.Graph();
dagreGraph.setDefaultEdgeLabel(() => ({}));

const nodeWidth = 172;
const nodeHeight = 36;

const getLayoutedElements = (nodes: Node[], edges: Edge[], direction = 'LR') => {
    const isHorizontal = direction === 'LR';
    dagreGraph.setGraph({ rankdir: direction });

    nodes.forEach((node) => {
        dagreGraph.setNode(node.id, { width: nodeWidth, height: nodeHeight });
    });

    edges.forEach((edge) => {
        dagreGraph.setEdge(edge.source, edge.target);
    });

    dagre.layout(dagreGraph);

    nodes.forEach((node) => {
        const nodeWithPosition = dagreGraph.node(node.id);
        node.targetPosition = isHorizontal ? Position.Left : Position.Top;
        node.sourcePosition = isHorizontal ? Position.Right : Position.Bottom;

        // We are shifting the dagre node position (anchor=center center) to the top left
        // so it matches the React Flow node anchor point (top left).
        node.position = {
            x: nodeWithPosition.x - nodeWidth / 2,
            y: nodeWithPosition.y - nodeHeight / 2,
        };

        return node;
    });

    return { nodes, edges };
};

// --- Custom Node Icons ---
const getIconForType = (type: string) => {
    switch (type.toLowerCase()) {
        case 'datasource': return <Database size={16} />;
        case 'process': return <Cog size={16} />; // specific icon import needed
        case 'api': return <Globe size={16} />;
        case 'file': return <FileText size={16} />;
        case 'storage': return <Server size={16} />;
        default: return <Layers size={16} />;
    }
};

const DataLineage: React.FC = () => {
    const [nodes, setNodes, onNodesChange] = useNodesState([]);
    const [edges, setEdges, onEdgesChange] = useEdgesState([]);
    const [selectedNode, setSelectedNode] = useState<LineageNode | null>(null);

    const { data, isLoading, isError } = useQuery({
        queryKey: ['lineage'],
        queryFn: governanceService.getLineage,
    });

    useEffect(() => {
        if (data) {
            const rfNodes: Node[] = data.nodes.map((node) => ({
                id: node.id,
                type: 'default', // Using default for now, could be custom
                data: {
                    label: (
                        <div className="flex items-center gap-2">
                            {getIconForType(node.type)}
                            <span className="font-medium text-sm">{node.data.label}</span>
                        </div>
                    ),
                    originalData: node // Store full node data for details panel
                },
                position: node.position || { x: 0, y: 0 },
                style: {
                    border: '1px solid #e2e8f0',
                    borderRadius: '8px',
                    padding: '8px 12px',
                    backgroundColor: 'white',
                    minWidth: '150px'
                }
            }));

            const rfEdges: Edge[] = data.edges.map((edge) => ({
                id: edge.id,
                source: edge.source,
                target: edge.target,
                type: 'smoothstep',
                animated: true,
                markerEnd: {
                    type: MarkerType.ArrowClosed,
                },
                style: { stroke: '#64748b' }
            }));

            const { nodes: layoutedNodes, edges: layoutedEdges } = getLayoutedElements(
                rfNodes,
                rfEdges
            );

            setNodes(layoutedNodes);
            setEdges(layoutedEdges);
        }
    }, [data, setNodes, setEdges]);

    const onNodeClick = useCallback((_: React.MouseEvent, node: Node) => {
        // Extract original lineage node data
        if (node.data && node.data.originalData) {
            setSelectedNode(node.data.originalData);
        }
    }, []);

    const onPaneClick = useCallback(() => {
        setSelectedNode(null);
    }, []);

    if (isLoading) {
        return (
            <div className="h-[calc(100vh-100px)] flex items-center justify-center">
                <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
            </div>
        );
    }

    if (isError) {
        return (
            <div className="p-6">
                <div className="bg-red-50 text-red-600 p-4 rounded-md">
                    Failed to load data lineage. Please try again later.
                </div>
            </div>
        );
    }

    return (
        <div className="h-[calc(100vh-64px)] flex flex-col relative">
            <div className="p-4 border-b bg-white flex justify-between items-center z-10">
                <div>
                    <h1 className="text-xl font-semibold text-slate-800">Data Lineage</h1>
                    <p className="text-sm text-slate-500">Visualize data flow across your infrastructure</p>
                </div>
                <div className="flex gap-2">
                    <span className="text-xs px-2 py-1 bg-blue-50 text-blue-700 rounded border border-blue-100 flex items-center gap-1">
                        <div className="w-2 h-2 rounded-full bg-blue-500 animate-pulse"></div>
                        Live Updates
                    </span>
                </div>
            </div>

            <div className="flex-1 relative flex">
                <div className="flex-1 h-full bg-slate-50">
                    <ReactFlow
                        nodes={nodes}
                        edges={edges}
                        onNodesChange={onNodesChange}
                        onEdgesChange={onEdgesChange}
                        onNodeClick={onNodeClick}
                        onPaneClick={onPaneClick}
                        connectionLineType={ConnectionLineType.SmoothStep}
                        fitView
                    >
                        <Controls />
                        <Background color="#cbd5e1" gap={16} />
                        <Panel position="top-right">
                            <div className="bg-white p-2 rounded shadow-sm border text-xs text-slate-500">
                                {nodes.length} Nodes • {edges.length} Edges
                            </div>
                        </Panel>
                    </ReactFlow>
                </div>

                {/* Details Panel */}
                <div className={`
                    absolute right-0 top-0 h-full w-80 bg-white shadow-xl border-l transform transition-transform duration-300 ease-in-out z-20 overflow-y-auto
                    ${selectedNode ? 'translate-x-0' : 'translate-x-full'}
                 `}>
                    {selectedNode && (
                        <div className="p-6">
                            <div className="flex items-center gap-3 mb-6">
                                <div className="p-2 bg-blue-50 rounded-lg text-blue-600">
                                    {getIconForType(selectedNode.type)}
                                </div>
                                <div>
                                    <h3 className="font-semibold text-lg leading-tight">{selectedNode.data.label}</h3>
                                    <span className="text-xs text-slate-500 uppercase tracking-wider font-semibold">{selectedNode.type}</span>
                                </div>
                            </div>

                            <div className="space-y-6">
                                <div>
                                    <h4 className="text-sm font-medium text-slate-900 mb-2">Details</h4>
                                    <div className="bg-slate-50 rounded p-3 space-y-2 text-sm">
                                        {selectedNode.data.details && Object.entries(selectedNode.data.details).map(([key, value]) => (
                                            <div key={key} className="flex justify-between">
                                                <span className="text-slate-500 capitalize">{key.replace(/_/g, ' ')}:</span>
                                                <span className="font-medium truncate ml-2 max-w-[150px]" title={String(value)}>{String(value)}</span>
                                            </div>
                                        ))}
                                        {!selectedNode.data.details && (
                                            <div className="text-slate-400 italic">No additional details available</div>
                                        )}
                                    </div>
                                </div>

                                {selectedNode.data.status && (
                                    <div>
                                        <h4 className="text-sm font-medium text-slate-900 mb-2">Status</h4>
                                        <StatusBadge label={selectedNode.data.status} />
                                    </div>
                                )}

                                <div>
                                    <h4 className="text-sm font-medium text-slate-900 mb-2">PII Types</h4>
                                    <div className="flex flex-wrap gap-1">
                                        {/* Mock PII data if not in details, typically this would come from the API */}
                                        {['Email', 'IP Address', 'Name'].map((tag) => (
                                            <span key={tag} className="px-2 py-1 bg-yellow-50 text-yellow-700 border border-yellow-200 rounded text-xs">
                                                {tag}
                                            </span>
                                        ))}
                                    </div>
                                </div>
                            </div>

                            <div className="mt-8 pt-6 border-t">
                                <button
                                    onClick={() => setSelectedNode(null)}
                                    className="w-full py-2 px-4 border border-slate-300 rounded text-slate-600 hover:bg-slate-50 text-sm font-medium transition-colors"
                                >
                                    Close Panel
                                </button>
                            </div>
                        </div>
                    )}
                </div>
            </div>
        </div>
    );
};

export default DataLineage;
