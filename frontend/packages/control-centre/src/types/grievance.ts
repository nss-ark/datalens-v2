import type { ID, TenantEntity } from '@datalens/shared';

export type GrievanceStatus = 'OPEN' | 'IN_PROGRESS' | 'RESOLVED' | 'ESCALATED' | 'CLOSED';
export type GrievancePriority = 'LOW' | 'MEDIUM' | 'HIGH' | 'CRITICAL';

export interface Grievance extends TenantEntity {
  data_subject_id: ID;
  subject: string;
  description: string;
  category: string; // e.g., 'CONSENT_WITHDRAWAL', 'DSR_DELAY', 'DATA_BREACH'
  status: GrievanceStatus;
  priority: GrievancePriority;
  submitted_at: string;
  due_date?: string;
  assigned_to?: ID;
  resolution?: string;
  resolved_at?: string;
  escalated_to?: string; // Authority name
  feedback_rating?: number; // 1-5
  feedback_comment?: string;
}

export interface CreateGrievanceRequest {
  subject: string;
  description: string;
  category: string;
  data_subject_id?: ID; // Optional if inferred from context
}

export interface GrievanceFilter {
  status?: GrievanceStatus;
  priority?: GrievancePriority;
  assigned_to?: ID;
}
