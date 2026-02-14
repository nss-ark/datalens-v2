export interface PurposeSuggestion {
    id: string;
    dataElement: string;
    currentPurpose?: string;
    suggestedPurpose: string;
    confidenceScore: number;
    dataSource: string;
    table: string;
    column: string;
    reasoning: string;
}

export type PolicyType = 'retention' | 'access' | 'encryption' | 'minimization';

export interface GovernancePolicy {
    id: string;
    name: string;
    type: PolicyType;
    description: string;
    rules: Record<string, unknown>;
    isActive: boolean;
    createdAt: string;
    updatedAt: string;
}

export type ViolationSeverity = 'low' | 'medium' | 'high' | 'critical';

export interface PolicyViolation {
    id: string;
    policyId: string;
    policyName: string;
    dataSource: string;
    dataElement: string;
    severity: ViolationSeverity;
    description: string;
    detectedAt: string;
    status: 'open' | 'resolved' | 'ignored';
}

export interface CreatePolicyRequest {
    name: string;
    type: PolicyType;
    description: string;
    rules: Record<string, unknown>;
}

// Data Lineage
export interface LineageData {
    label: string;
    type: string;
    status?: string;
    details?: Record<string, unknown>;
}

export interface LineageNode {
    id: string;
    type: string; // 'dataSource' | 'process' | 'storage'
    position: { x: number; y: number }; // ReactFlow position
    data: LineageData;
}

export interface LineageEdge {
    id: string;
    source: string;
    target: string;
    animated?: boolean;
    label?: string;
}

export interface LineageGraph {
    nodes: LineageNode[];
    edges: LineageEdge[];
}
