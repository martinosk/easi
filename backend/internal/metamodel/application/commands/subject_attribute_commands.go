package commands

import "easi/backend/internal/metamodel/publishedlanguage"

type ImportSubjectAttribute = publishedlanguage.ImportSubjectAttribute

type CreateSubjectAttributeSchema struct {
	TenantID    string
	SubjectType string
	CreatedBy   string
}

func (c CreateSubjectAttributeSchema) CommandName() string { return "CreateSubjectAttributeSchema" }

type DefineSubjectAttribute struct {
	SchemaID      string
	Name          string
	AttributeType string
	HelpText      string
	OptionLabels  []string
	Min           *float64
	Max           *float64
	ModifiedBy    string
}

func (c DefineSubjectAttribute) CommandName() string     { return "DefineSubjectAttribute" }
func (c DefineSubjectAttribute) SubjectSchemaID() string { return c.SchemaID }
func (c DefineSubjectAttribute) ModifiedByEmail() string { return c.ModifiedBy }

type RenameSubjectAttribute struct {
	SchemaID      string
	AttributeID   string
	Name          string
	HelpText      string
	RequestedType string
	ModifiedBy    string
}

func (c RenameSubjectAttribute) CommandName() string     { return "RenameSubjectAttribute" }
func (c RenameSubjectAttribute) SubjectSchemaID() string { return c.SchemaID }
func (c RenameSubjectAttribute) ModifiedByEmail() string { return c.ModifiedBy }

type RetireSubjectAttribute struct {
	SchemaID    string
	AttributeID string
	ModifiedBy  string
}

func (c RetireSubjectAttribute) CommandName() string     { return "RetireSubjectAttribute" }
func (c RetireSubjectAttribute) SubjectSchemaID() string { return c.SchemaID }
func (c RetireSubjectAttribute) ModifiedByEmail() string { return c.ModifiedBy }

type ReactivateSubjectAttribute struct {
	SchemaID    string
	AttributeID string
	ModifiedBy  string
}

func (c ReactivateSubjectAttribute) CommandName() string     { return "ReactivateSubjectAttribute" }
func (c ReactivateSubjectAttribute) SubjectSchemaID() string { return c.SchemaID }
func (c ReactivateSubjectAttribute) ModifiedByEmail() string { return c.ModifiedBy }

type AddSubjectAttributeOption struct {
	SchemaID    string
	AttributeID string
	Label       string
	ModifiedBy  string
}

func (c AddSubjectAttributeOption) CommandName() string     { return "AddSubjectAttributeOption" }
func (c AddSubjectAttributeOption) SubjectSchemaID() string { return c.SchemaID }
func (c AddSubjectAttributeOption) ModifiedByEmail() string { return c.ModifiedBy }

type RetireSubjectAttributeOption struct {
	SchemaID    string
	AttributeID string
	OptionID    string
	ModifiedBy  string
}

func (c RetireSubjectAttributeOption) CommandName() string     { return "RetireSubjectAttributeOption" }
func (c RetireSubjectAttributeOption) SubjectSchemaID() string { return c.SchemaID }
func (c RetireSubjectAttributeOption) ModifiedByEmail() string { return c.ModifiedBy }

type SetSubjectAttributeBounds struct {
	SchemaID    string
	AttributeID string
	Min         *float64
	Max         *float64
	ModifiedBy  string
}

func (c SetSubjectAttributeBounds) CommandName() string     { return "SetSubjectAttributeBounds" }
func (c SetSubjectAttributeBounds) SubjectSchemaID() string { return c.SchemaID }
func (c SetSubjectAttributeBounds) ModifiedByEmail() string { return c.ModifiedBy }
