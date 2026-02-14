import type { ID, TenantEntity } from '@datalens/shared';

export type NodeType = 'DATA_SOURCE' | 'PROCESS' | 'THIRD_PARTY';
export type FlowStatus = 'ACTIVE' | 'INACTIVE' | 'PROPOSED';

export interface GraphNode {
    id: string;
    label: string;
    type: NodeType;
    data?: Record<string, unknown>;
}

export interface GraphEdge {
    id: string;
    source: string;
    target: string;
    label?: string;
    animated?: boolean;
    flowId?: string;
}

export interface LineageGraph {
    nodes: GraphNode[];
    edges: GraphEdge[];
}

export interface DataFlow extends TenantEntity {
    source_id: ID;
    destination_id: ID;
    data_type: string; // "TABLE", "COLUMN", "FILE"
    data_path: string;
    purpose_id?: ID;
    status: FlowStatus;
    description: string;
}
