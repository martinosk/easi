import { httpClient } from '../../../api/core/httpClient';
import { followLink, getLink, type ResourceWithLinks } from '../../../utils/hateoas';
import type {
  AddSelectionOptionRequest,
  BuiltInField,
  ChangeRequirementRequest,
  CustomField,
  DefineCustomFieldRequest,
  FieldValue,
  OnePagerConfiguration,
  OnePagerFacts,
  OnePagerImpactPreview,
  OnePagerSubjectType,
  OnePagerView,
  RecordFieldValueRequest,
  RenameCustomFieldRequest,
  ReorderFieldsRequest,
  SelectionOption,
  VersionRequest,
} from '../types';

async function sendCommand(
  method: 'post' | 'put',
  resource: ResourceWithLinks,
  rel: string,
  request: unknown,
): Promise<OnePagerConfiguration> {
  const response = await httpClient[method]<OnePagerConfiguration>(followLink(resource, rel), request);
  return response.data;
}

export const onePagersApi = {
  async getConfiguration(subjectType: OnePagerSubjectType): Promise<OnePagerConfiguration> {
    const response = await httpClient.get<OnePagerConfiguration>(`/api/v1/one-pagers/configurations/${subjectType}`);
    return response.data;
  },

  defineCustomField: (configuration: OnePagerConfiguration, request: DefineCustomFieldRequest) =>
    sendCommand('post', configuration, 'x-define-custom-field', request),

  reorderFields: (configuration: OnePagerConfiguration, request: ReorderFieldsRequest) =>
    sendCommand('put', configuration, 'x-reorder', request),

  includeBuiltInField: (field: BuiltInField, request: VersionRequest) =>
    sendCommand('post', field, 'x-include', request),

  excludeBuiltInField: (field: BuiltInField, request: VersionRequest) =>
    sendCommand('post', field, 'x-exclude', request),

  renameCustomField: (field: CustomField, request: RenameCustomFieldRequest) =>
    sendCommand('put', field, 'x-rename', request),

  changeFieldRequirement: (field: CustomField, request: ChangeRequirementRequest) =>
    sendCommand('put', field, 'x-set-requirement', request),

  retireCustomField: (field: CustomField, request: VersionRequest) => sendCommand('post', field, 'x-retire', request),

  reactivateCustomField: (field: CustomField, request: VersionRequest) =>
    sendCommand('post', field, 'x-reactivate', request),

  addSelectionOption: (field: CustomField, request: AddSelectionOptionRequest) =>
    sendCommand('post', field, 'x-add-option', request),

  retireSelectionOption: (option: SelectionOption, request: VersionRequest) =>
    sendCommand('post', option, 'x-retire', request),

  async getImpactPreview(configuration: OnePagerConfiguration, fieldId?: string): Promise<OnePagerImpactPreview> {
    const base = getLink(configuration, 'x-impact-preview');
    if (!base) throw new Error("Link 'x-impact-preview' not found on resource");
    const url = fieldId ? `${base}?fieldId=${encodeURIComponent(fieldId)}` : base;
    const response = await httpClient.get<OnePagerImpactPreview>(url);
    return response.data;
  },

  async getFacts(subjectType: OnePagerSubjectType, subjectId: string): Promise<OnePagerFacts> {
    const response = await httpClient.get<OnePagerFacts>(`/api/v1/one-pagers/${subjectType}/${subjectId}/facts`);
    return response.data;
  },

  async getOnePager(subjectType: OnePagerSubjectType, subjectId: string): Promise<OnePagerView> {
    const response = await httpClient.get<OnePagerView>(`/api/v1/one-pagers/${subjectType}/${subjectId}`);
    return response.data;
  },

  async recordFieldValue(
    facts: OnePagerFacts,
    fieldId: string,
    request: RecordFieldValueRequest,
  ): Promise<OnePagerFacts> {
    const response = await httpClient.put<OnePagerFacts>(`${followLink(facts, 'x-record')}/${fieldId}`, request);
    return response.data;
  },

  async clearFieldValue(fieldValue: FieldValue): Promise<OnePagerFacts> {
    const response = await httpClient.delete<OnePagerFacts>(followLink(fieldValue, 'x-clear'));
    return response.data;
  },
};
