import { httpClient } from '../../../api/core/httpClient';
import type { HATEOASLink } from '../../../api/types';
import type { ConcernStewardship, DomainStewardships } from '../types';

export const stewardshipApi = {
  async getDomainStewardships(href: string): Promise<DomainStewardships> {
    const response = await httpClient.get<DomainStewardships>(href);
    return response.data;
  },

  async assign(link: HATEOASLink, stewardId: string): Promise<ConcernStewardship> {
    const response = await httpClient.put<ConcernStewardship>(link.href, { stewardId });
    return response.data;
  },

  async release(link: HATEOASLink): Promise<void> {
    await httpClient.delete(link.href);
  },
};
