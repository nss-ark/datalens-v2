import React, { useCallback, useEffect } from 'react';
import ReactFlow, {
    Background,
    Controls,
    useNodesState,
    useEdgesState,
    Panel,
} from 'reactflow';
import type { Edge } from 'reactflow';
import type { NodeTypes } from 'reactflow';
import type { Node } from 'reactflow';
import dagre from 'dagre';
import 'reactflow/dist/style.css';

import CustomNode from './CustomNode';
import type { GraphNode, GraphEdge } from '../../../types/lineage';
import { Position } from 'reactflow';

// Define custom node types
const nodeTypes: NodeTypes = {
    custom: CustomNode,
};

interface LineageGraphProps {
    initialNodes: GraphNode[];
    initialEdges: GraphEdge[];
    onNodeClick: (node: GraphNode) => void;
}

const nodeWidth = 200;
const nodeHeight = 80;

const getLayoutedElements = (nodes: Node[], edges: Edge[]) => {
    const dagreGraph = new dagre.graphlib.Graph();
    dagreGraph.setDefaultEdgeLabel(() => ({}));

    dagreGraph.setGraph({ rankdir: 'LR' });

    nodes.forEach((node) => {
        dagreGraph.setNode(node.id, { width: nodeWidth, height: nodeHeight });
    });

    edges.forEach((edge) => {
        dagreGraph.setEdge(edge.source, edge.target);
    });

    dagre.layout(dagreGraph);

    const layoutedNodes = nodes.map((node) => {
        const nodeWithPosition = dagreGraph.node(node.id);
        node.sourcePosition = Position.Right;
        node.targetPosition = Position.Left;

        // We are shifting the dagre node position (anchor=center center) to the top left
        // so it matches the React Flow node anchor point (top left).
        node.position = {
            x: nodeWithPosition.x - nodeWidth / 2,
            y: nodeWithPosition.y - nodeHeight / 2,
        };

        return node;
    });

    return { nodes: layoutedNodes, edges };
};

const LineageGraph: React.FC<LineageGraphProps> = ({ initialNodes, initialEdges, onNodeClick }) => {
    const [nodes, setNodes, onNodesChange] = useNodesState([]);
    const [edges, setEdges, onEdgesChange] = useEdgesState([]);

    useEffect(() => {
        // Transform backend nodes/edges to ReactFlow format
        const rfNodes: Node[] = initialNodes.map((n) => ({
            id: n.id,
            type: 'custom',
            data: { ...n.data, label: n.label, type: n.type },
            position: { x: 0, y: 0 }, // Position will be calculated by dagre
        }));

        const rfEdges: Edge[] = initialEdges.map((e) => ({
            id: e.id,
            source: e.source,
            target: e.target,
            type: 'smoothstep',
            animated: true,
            label: e.label,
        }));

        const { nodes: layoutedNodes, edges: layoutedEdges } = getLayoutedElements(
            rfNodes,
            rfEdges
        );

        setNodes(layoutedNodes);
        setEdges(layoutedEdges);
    }, [initialNodes, initialEdges, setNodes, setEdges]);

    const handleNodeClick = useCallback(
        (_: React.MouseEvent, node: Node) => {
            // Find original node data
            const originalNode = initialNodes.find((n) => n.id === node.id);
            if (originalNode) {
                onNodeClick(originalNode);
            }
        },
        [initialNodes, onNodeClick]
    );

    return (
        <div className="h-full w-full bg-slate-50 border rounded-lg overflow-hidden">
            <ReactFlow
                nodes={nodes}
                edges={edges}
                onNodesChange={onNodesChange}
                onEdgesChange={onEdgesChange}
                onNodeClick={handleNodeClick}
                nodeTypes={nodeTypes}
                fitView
            >
                <Background color="#ccc" gap={20} />
                <Controls />
                <Panel position="bottom-right" className="bg-white p-2 rounded shadow text-xs text-gray-500">
                    Auto-generated from Data Flows
                </Panel>
            </ReactFlow>
        </div>
    );
};

export default LineageGraph;
