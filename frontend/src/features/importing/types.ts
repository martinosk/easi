declare const __brand: unique symbol;
type Brand<B> = { [__brand]: B };
type Branded<T, B> = T & Brand<B>;

export type ImportSessionId = Branded<string, 'ImportSessionId'>;

export type ImportStatus = 'pending' | 'importing' | 'completed' | 'failed';
export type SourceFormat = 'archimate-openexchange';

export interface ImportPreview {
  supported: {
    capabilities: number;
    components: number;
    parentChildRelationships: number;
    realizations: number;
  };
  unsupported: {
    elements: Record<string, number>;
    relationships: Record<string, number>;
  };
}

export interface ImportProgress {
  phase: string;
  totalItems: number;
  completedItems: number;
}

export interface ImportError {
  sourceElement: string;
  sourceName: string;
  error: string;
  action: string;
}

export interface ImportResult {
  capabilitiesCreated: number;
  componentsCreated: number;
  realizationsCreated: number;
  domainAssignments: number;
  errors: ImportError[];
}

export interface ImportSession {
  id: ImportSessionId;
  status: ImportStatus;
  sourceFormat: SourceFormat;
  businessDomainId?: string;
  preview?: ImportPreview;
  progress?: ImportProgress;
  result?: ImportResult;
  createdAt: string;
  completedAt?: string;
  _links: {
    self: string;
    confirm?: string;
    delete?: string;
  };
}

export interface CreateImportSessionRequest {
  file: File;
  sourceFormat: SourceFormat;
  businessDomainId?: string;
}
