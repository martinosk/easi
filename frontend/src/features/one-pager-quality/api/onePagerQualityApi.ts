import { httpClient } from '../../../api/core/httpClient';
import type { OnePagerQualityOrder, OnePagerQualityResponse, OnePagerQualitySort } from '../types';

export interface GetOnePagerQualityParams {
  sort?: OnePagerQualitySort;
  order?: OnePagerQualityOrder;
  limit?: number;
  cursor?: string;
}

export const onePagerQualityApi = {
  async getList(params: GetOnePagerQualityParams = {}): Promise<OnePagerQualityResponse> {
    const searchParams = new URLSearchParams();
    if (params.sort) searchParams.set('sort', params.sort);
    if (params.order) searchParams.set('order', params.order);
    if (params.limit) searchParams.set('limit', String(params.limit));
    if (params.cursor) searchParams.set('after', params.cursor);

    const query = searchParams.toString();
    const url = query ? `/api/v1/one-pager-quality?${query}` : '/api/v1/one-pager-quality';
    const response = await httpClient.get<OnePagerQualityResponse>(url);
    return response.data;
  },
};
