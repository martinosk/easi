import { httpClient } from '../../../api/core/httpClient';
import type { StrategicFitAnalysis } from '../../../api/types';

export const strategicFitApi = {
  async getStrategicFitAnalysis(pillarId: string): Promise<StrategicFitAnalysis> {
    const response = await httpClient.get<StrategicFitAnalysis>(`/api/v1/strategic-fit-analysis/${pillarId}`);
    return response.data;
  },
};
