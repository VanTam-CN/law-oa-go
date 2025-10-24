/**
 * Domain-specific type definitions for Law OA System
 */

export interface UserData {
  id: string;
  username: string;
  email: string;
  password: string;
  role: 'admin' | 'attorney' | 'paralegal' | 'client';
  department?: string;
  status: 'active' | 'inactive' | 'pending';
  createdAt: Date;
  updatedAt: Date;
  profile?: UserProfile;
}

export interface UserProfile {
  firstName?: string;
  lastName?: string;
  phone?: string;
  avatar?: string;
  bio?: string;
  specialization?: string[];
}

export interface ClientData {
  id: string;
  name: string;
  type: 'individual' | 'corporate';
  registrationNumber?: string;
  industry?: string;
  contactPerson?: string;
  email?: string;
  phone?: string;
  address?: Address;
  status: 'active' | 'inactive' | 'prospect';
  createdAt: Date;
  updatedAt: Date;
  cases?: string[]; // Case IDs
}

export interface Address {
  street?: string;
  city?: string;
  state?: string;
  zipCode?: string;
  country?: string;
}

export interface CaseData {
  id: string;
  title: string;
  caseNumber: string;
  type: 'litigation' | 'corporate' | 'family' | 'criminal' | 'other';
  status: 'active' | 'closed' | 'pending' | 'archived';
  priority: 'low' | 'medium' | 'high' | 'urgent';
  assignedTo: string; // User ID
  client: string; // Client ID
  description?: string;
  estimatedValue?: string;
  createdDate: Date;
  updatedDate: Date;
  dueDate?: Date;
  milestones?: CaseMilestone[];
  documents?: string[]; // Document IDs
  conflicts?: ConflictCheck[];
}

export interface CaseMilestone {
  id: string;
  name: string;
  description?: string;
  dueDate: Date;
  completedDate?: Date;
  status: 'pending' | 'completed' | 'overdue';
  priority: 'low' | 'medium' | 'high';
}

export interface DocumentData {
  id: string;
  name: string;
  type: 'contract' | 'evidence' | 'correspondence' | 'court_filing' | 'other';
  caseId: string;
  clientId?: string;
  uploadedBy: string; // User ID
  fileSize: number;
  mimeType: string;
  path: string;
  tags: string[];
  status: 'active' | 'archived' | 'deleted';
  uploadedAt: Date;
  updatedAt: Date;
  version?: number;
  metadata?: DocumentMetadata;
}

export interface DocumentMetadata {
  author?: string;
  creationDate?: Date;
  modificationDate?: Date;
  page_count?: number;
  language?: string;
  encryption?: string;
  checksum?: string;
}

export interface ConflictCheck {
  id: string;
  caseId: string;
  clientIds: string[];
  opposingParty?: string;
  checkDate: Date;
  status: 'pending' | 'passed' | 'failed' | 'review';
  conflicts: Conflict[];
  reviewer?: string; // User ID
  notes?: string;
}

export interface Conflict {
  id: string;
  type: 'client' | 'party' | 'matter' | 'other';
  description: string;
  severity: 'low' | 'medium' | 'high' | 'critical';
  relatedCase?: string; // Case ID
  relatedClient?: string; // Client ID
  resolution?: string;
  resolved?: boolean;
  resolvedDate?: Date;
  resolvedBy?: string; // User ID
}

export interface FinancialTransaction {
  id: string;
  caseId?: string;
  clientId?: string;
  type: 'invoice' | 'payment' | 'expense' | 'refund' | 'retainer';
  amount: number;
  currency: string;
  description: string;
  date: Date;
  status: 'pending' | 'completed' | 'cancelled' | 'disputed';
  method?: 'cash' | 'check' | 'bank_transfer' | 'credit_card' | 'other';
  reference?: string;
  createdBy: string; // User ID
  approvedBy?: string; // User ID
  approvedDate?: Date;
  category?: FinancialCategory;
}

export interface FinancialCategory {
  id: string;
  name: string;
  type: 'income' | 'expense';
  code?: string;
  description?: string;
  taxable?: boolean;
}

export interface SearchQuery {
  id: string;
  query: string;
  type: 'global' | 'cases' | 'clients' | 'documents' | 'users';
  filters?: SearchFilter[];
  sort?: SortOption[];
  userId: string; // User ID
  timestamp: Date;
  results?: SearchResult[];
  totalResults?: number;
  executionTime?: number;
}

export interface SearchFilter {
  field: string;
  operator: 'equals' | 'contains' | 'starts_with' | 'ends_with' | 'greater_than' | 'less_than' | 'in';
  value: string | number | string[];
}

export interface SortOption {
  field: string;
  direction: 'asc' | 'desc';
}

export interface SearchResult {
  type: 'case' | 'client' | 'document' | 'user';
  id: string;
  title: string;
  description?: string;
  relevanceScore: number;
  highlights?: string[];
  metadata?: Record<string, any>;
}

export interface ReportData {
  id: string;
  name: string;
  type: 'case_summary' | 'financial' | 'workload' | 'conflict' | 'custom';
  description?: string;
  generatedBy: string; // User ID
  generatedAt: Date;
  parameters: ReportParameters;
  data: any;
  format: 'pdf' | 'excel' | 'csv' | 'html';
  shareable: boolean;
  expiresAt?: Date;
}

export interface ReportParameters {
  dateRange?: {
    start: Date;
    end: Date;
  };
  filters?: Record<string, any>;
  groupBy?: string[];
  aggregations?: string[];
  sort?: SortOption[];
}

export interface SystemLog {
  id: string;
  timestamp: Date;
  level: 'debug' | 'info' | 'warn' | 'error';
  category: 'auth' | 'api' | 'database' | 'security' | 'performance' | 'business';
  message: string;
  userId?: string;
  sessionId?: string;
  requestId?: string;
  metadata?: Record<string, any>;
  source: {
    file: string;
    line: number;
    function: string;
  };
}

// Test data types
export interface DataSet {
  id: string;
  name: string;
  description?: string;
  users: UserData[];
  clients: ClientData[];
  cases: CaseData[];
  documents: DocumentData[];
  conflicts: ConflictCheck[];
  transactions: FinancialTransaction[];
  queries: SearchQuery[];
  logs: SystemLog[];
  metadata: Record<string, any>;
  createdAt: Date;
  updatedAt: Date;
}

export interface DataTemplate {
  id: string;
  name: string;
  description: string;
  version: string;
  schema: any;
  generators: DataGenerator[];
  relationships: DataRelationship[];
  createdAt: Date;
  updatedAt: Date;
}

export interface DataGenerator {
  field: string;
  type: 'static' | 'random' | 'sequence' | 'faker' | 'custom';
  config: any;
}

export interface DataRelationship {
  from: {
    entity: string;
    field: string;
  };
  to: {
    entity: string;
    field: string;
  };
  type: 'one_to_one' | 'one_to_many' | 'many_to_many';
  cardinality?: number;
}