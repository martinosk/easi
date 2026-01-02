import { httpClient } from '../core';
import type {
  StatusOption,
  OwnershipModelOption,
  MaturityLevelsResponse,
  StatusesResponse,
  OwnershipModelsResponse,
  VersionResponse,
  Release,
  ReleasesResponse,
  ReleaseVersion,
  MaturityScale,
} from '../types';

export const metadataApi = {
  async getMaturityLevels(): Promise<string[]> {
    const response = await httpClient.get<MaturityLevelsResponse>(
      '/api/v1/capabilities/metadata/maturity-levels'
    );
    return response.data.data.map((level) => level.value);
  },

  async getStatuses(): Promise<StatusOption[]> {
    const response = await httpClient.get<StatusesResponse>(
      '/api/v1/capabilities/metadata/statuses'
    );
    return response.data.data;
  },

  async getOwnershipModels(): Promise<OwnershipModelOption[]> {
    const response = await httpClient.get<OwnershipModelsResponse>(
      '/api/v1/capabilities/metadata/ownership-models'
    );
    return response.data.data;
  },

  async getVersion(): Promise<string> {
    const response = await httpClient.get<VersionResponse>('/api/v1/version');
    return response.data.version;
  },

  async getLatestRelease(): Promise<Release | null> {
    try {
      const response = await httpClient.get<Release>('/api/v1/releases/latest');
      return response.data;
    } catch {
      return null;
    }
  },

  async getReleaseByVersion(version: ReleaseVersion): Promise<Release | null> {
    try {
      const response = await httpClient.get<Release>(`/api/v1/releases/${version}`);
      return response.data;
    } catch {
      return null;
    }
  },

  async getReleases(): Promise<Release[]> {
    const response = await httpClient.get<ReleasesResponse>('/api/v1/releases');
    return response.data.data || [];
  },

  async getMaturityScale(): Promise<MaturityScale> {
    const response = await httpClient.get<MaturityScale>('/api/v1/meta-model/maturity-scale');
    return response.data;
  },
};

export default metadataApi;
