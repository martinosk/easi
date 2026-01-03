import { httpClient } from '../../../api/core';
import type { AuditHistoryResponse } from '../../../api/types';

export interface GetAuditHistoryParams {
  aggregateId: string;
  limit?: number;
  cursor?: string;
}

export const auditApi = {
  async getHistory({ aggregateId, limit = 50, cursor }: GetAuditHistoryParams): Promise<AuditHistoryResponse> {
    const params = new URLSearchParams();
    params.set('limit', String(limit));
    if (cursor) {
      params.set('cursor', cursor);
    }

    const response = await httpClient.get<AuditHistoryResponse>(
      `/api/v1/audit/${aggregateId}?${params.toString()}`
    );
    return response.data;
  },
};
