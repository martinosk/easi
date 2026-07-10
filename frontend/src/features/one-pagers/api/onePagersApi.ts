import { httpClient } from '../../../api/core/httpClient';
import { followLink, type ResourceWithLinks } from '../../../utils/hateoas';
import type {
  AddSelectionOptionRequest,
  BuiltInField,
  ChangeRequirementRequest,
  CustomField,
  DefineCustomFieldRequest,
  OnePagerConfiguration,
  OnePagerSubjectType,
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
};
