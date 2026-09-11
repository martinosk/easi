import { followLink, type ResourceWithLinks } from '../../utils/hateoas';
import { httpClient } from '../core/httpClient';
import type {
  AddAttributeOptionRequest,
  DefineSubjectAttributeRequest,
  RenameSubjectAttributeRequest,
  SchemaVersionRequest,
  SetAttributeBoundsRequest,
  SubjectAttribute,
  SubjectAttributeOption,
  SubjectAttributeSchema,
} from '../types';

async function sendSchemaCommand(
  method: 'post' | 'put',
  resource: ResourceWithLinks,
  rel: string,
  request: unknown,
): Promise<SubjectAttributeSchema> {
  const response = await httpClient[method]<SubjectAttributeSchema>(followLink(resource, rel), request);
  return response.data;
}

export const subjectAttributeSchemaApi = {
  async getSchema(subjectType: string): Promise<SubjectAttributeSchema> {
    const response = await httpClient.get<SubjectAttributeSchema>(
      `/api/v1/meta-model/subject-types/${subjectType}/attributes`,
    );
    return response.data;
  },

  defineAttribute: (schema: SubjectAttributeSchema, request: DefineSubjectAttributeRequest) =>
    sendSchemaCommand('post', schema, 'x-define', request),

  renameAttribute: (attribute: SubjectAttribute, request: RenameSubjectAttributeRequest) =>
    sendSchemaCommand('put', attribute, 'x-rename', request),

  retireAttribute: (attribute: SubjectAttribute, request: SchemaVersionRequest) =>
    sendSchemaCommand('post', attribute, 'x-retire', request),

  reactivateAttribute: (attribute: SubjectAttribute, request: SchemaVersionRequest) =>
    sendSchemaCommand('post', attribute, 'x-reactivate', request),

  addOption: (attribute: SubjectAttribute, request: AddAttributeOptionRequest) =>
    sendSchemaCommand('post', attribute, 'x-add-option', request),

  retireOption: (option: SubjectAttributeOption, request: SchemaVersionRequest) =>
    sendSchemaCommand('post', option, 'x-retire', request),

  setBounds: (attribute: SubjectAttribute, request: SetAttributeBoundsRequest) =>
    sendSchemaCommand('put', attribute, 'x-set-bounds', request),
};

export default subjectAttributeSchemaApi;
