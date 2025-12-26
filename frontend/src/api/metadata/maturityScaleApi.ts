import { httpClient } from '../core/httpClient';
import type {
  MaturityScaleConfiguration,
  UpdateMaturityScaleRequest,
} from '../types';

export const maturityScaleApi = {
  async getConfiguration(): Promise<MaturityScaleConfiguration> {
    const response = await httpClient.get<MaturityScaleConfiguration>(
      '/api/v1/meta-model/maturity-scale'
    );
    return response.data;
  },

  async updateConfiguration(request: UpdateMaturityScaleRequest): Promise<MaturityScaleConfiguration> {
    const response = await httpClient.put<MaturityScaleConfiguration>(
      '/api/v1/meta-model/maturity-scale',
      request
    );
    return response.data;
  },

  async resetToDefaults(): Promise<MaturityScaleConfiguration> {
    const response = await httpClient.post<MaturityScaleConfiguration>(
      '/api/v1/meta-model/maturity-scale/reset'
    );
    return response.data;
  },
};

export default maturityScaleApi;
