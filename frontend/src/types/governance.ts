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
