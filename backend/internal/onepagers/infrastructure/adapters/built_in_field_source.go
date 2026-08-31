package adapters

import (
	"context"
	"fmt"
	"slices"

	"easi/backend/internal/infrastructure/database"
	"easi/backend/internal/onepagers/application/ports"
	"easi/backend/internal/onepagers/application/readmodels"
	"easi/backend/internal/onepagers/domain/catalog"
	"easi/backend/internal/onepagers/domain/valueobjects"
)

type SubjectAttributeStore interface {
	AttributeRows(ctx context.Context, subjectType string, subjectIDs []string) ([]readmodels.SubjectAttributeRow, error)
	CountSubjects(ctx context.Context, subjectType string) (int, error)
	Exists(ctx context.Context, subject readmodels.SubjectKey) (bool, error)
}

type SubjectRelationReader interface {
	References(ctx context.Context, query readmodels.RelationQuery) (map[string][]readmodels.RelationReference, error)
	CountSubjectsWithEntry(ctx context.Context, subjectType, entryID string) (int, error)
}

type entrySet []string

func (e entrySet) contains(entryID string) bool {
	return slices.Contains(e, entryID)
}

type builtInFieldSource struct {
	subjectType     string
	attributes      []builtInAttribute
	relationEntries entrySet
	subjects        SubjectAttributeStore
	relations       SubjectRelationReader
}

func NewOnePagerBuiltInFieldSources(db *database.TenantAwareDB) map[string]ports.BuiltInFieldSource {
	return NewBuiltInFieldSources(
		readmodels.NewOnePagerSubjectIndexReadModel(db),
		readmodels.NewSubjectRelationCacheReadModel(db),
	)
}

func NewBuiltInFieldSources(subjects SubjectAttributeStore, relations SubjectRelationReader) map[string]ports.BuiltInFieldSource {
	sources := make(map[string]ports.BuiltInFieldSource, len(valueobjects.AllSubjectTypes()))
	for _, subjectType := range valueobjects.AllSubjectTypes() {
		sources[subjectType.Value()] = builtInFieldSource{
			subjectType:     subjectType.Value(),
			attributes:      builtInAttributesBySubjectType[subjectType.Value()],
			relationEntries: relationEntryIDs(subjectType),
			subjects:        subjects,
			relations:       relations,
		}
	}
	return sources
}

func relationEntryIDs(subjectType valueobjects.SubjectType) entrySet {
	entryIDs := entrySet{}
	for _, entry := range catalog.EntriesFor(subjectType) {
		if entry.Relation {
			entryIDs = append(entryIDs, entry.ID)
		}
	}
	return entryIDs
}

func (s builtInFieldSource) FetchSubject(ctx context.Context, subjectID string, includedEntryIDs []string) (*ports.SubjectSnapshot, error) {
	rows, err := s.subjects.AttributeRows(ctx, s.subjectType, []string{subjectID})
	if err != nil {
		return nil, s.wrap("fetch", err)
	}
	if len(rows) == 0 {
		return nil, nil
	}

	references, err := s.referencesFor(ctx, []string{subjectID}, includedEntryIDs)
	if err != nil {
		return nil, err
	}
	return s.snapshot(rows[0], references[subjectID], includedEntryIDs)
}

func (s builtInFieldSource) CountSubjects(ctx context.Context) (int, error) {
	count, err := s.subjects.CountSubjects(ctx, s.subjectType)
	if err != nil {
		return 0, s.wrap("count", err)
	}
	return count, nil
}

func (s builtInFieldSource) FilledBuiltInFields(ctx context.Context, subjectIDs, entryIDs []string) (map[string]map[string]bool, error) {
	filled := map[string]map[string]bool{}
	if len(subjectIDs) == 0 || len(entryIDs) == 0 {
		return filled, nil
	}

	rows, err := s.subjects.AttributeRows(ctx, s.subjectType, subjectIDs)
	if err != nil {
		return nil, s.wrap("fetch", err)
	}
	references, err := s.referencesFor(ctx, subjectIDs, entryIDs)
	if err != nil {
		return nil, err
	}

	for _, row := range rows {
		snapshot, err := s.snapshot(row, references[row.SubjectID], entryIDs)
		if err != nil {
			return nil, err
		}
		filled[row.SubjectID] = filledEntries(snapshot, entryIDs)
	}
	return filled, nil
}

func (s builtInFieldSource) CountSubjectsWithBuiltInValue(ctx context.Context, entryID string) (int, error) {
	if s.relationEntries.contains(entryID) {
		return s.countSubjectsWithRelation(ctx, entryID)
	}
	return s.countSubjectsWithAttribute(ctx, entryID)
}

func (s builtInFieldSource) countSubjectsWithRelation(ctx context.Context, entryID string) (int, error) {
	count, err := s.relations.CountSubjectsWithEntry(ctx, s.subjectType, entryID)
	if err != nil {
		return 0, s.wrap("count related", err)
	}
	return count, nil
}

func (s builtInFieldSource) countSubjectsWithAttribute(ctx context.Context, entryID string) (int, error) {
	rows, err := s.subjects.AttributeRows(ctx, s.subjectType, nil)
	if err != nil {
		return 0, s.wrap("fetch", err)
	}
	count := 0
	for _, row := range rows {
		snapshot, err := s.snapshot(row, nil, nil)
		if err != nil {
			return 0, err
		}
		if ports.ValueFilled(snapshot.Fields[entryID]) {
			count++
		}
	}
	return count, nil
}

func (s builtInFieldSource) snapshot(row readmodels.SubjectAttributeRow, references []readmodels.RelationReference, includedEntryIDs entrySet) (*ports.SubjectSnapshot, error) {
	fields := map[string]ports.BuiltInFieldValue{nameEntryID: ports.TextValue{Text: row.Name}}
	for _, attribute := range s.attributes {
		raw, _ := row.Attributes.Raw(attribute.attribute)
		value, err := attribute.decode(raw)
		if err != nil {
			return nil, fmt.Errorf("read %s %s built-in field %s: %w", s.subjectType, row.SubjectID, attribute.entryID, err)
		}
		fields[attribute.entryID] = value
	}
	for _, entryID := range s.includedRelations(includedEntryIDs) {
		fields[entryID] = referenceList(references, entryID)
	}
	return &ports.SubjectSnapshot{Name: row.Name, Fields: fields}, nil
}

func (s builtInFieldSource) includedRelations(includedEntryIDs entrySet) entrySet {
	included := make(entrySet, 0, len(s.relationEntries))
	for _, entryID := range s.relationEntries {
		if includedEntryIDs.contains(entryID) {
			included = append(included, entryID)
		}
	}
	return included
}

func (s builtInFieldSource) referencesFor(ctx context.Context, subjectIDs []string, includedEntryIDs entrySet) (map[string][]readmodels.RelationReference, error) {
	entryIDs := s.includedRelations(includedEntryIDs)
	if len(entryIDs) == 0 {
		return map[string][]readmodels.RelationReference{}, nil
	}
	references, err := s.relations.References(ctx, readmodels.RelationQuery{SubjectType: s.subjectType, SubjectIDs: subjectIDs, EntryIDs: entryIDs})
	if err != nil {
		return nil, s.wrap("resolve relations of", err)
	}
	return references, nil
}

func referenceList(references []readmodels.RelationReference, entryID string) ports.ReferenceListValue {
	matched := make([]ports.Reference, 0, len(references))
	for _, reference := range references {
		if reference.EntryID != entryID {
			continue
		}
		matched = append(matched, ports.Reference{ID: reference.RelatedID, Label: reference.Label, SubjectType: reference.RelatedType})
	}
	return ports.ReferenceListValue{References: matched}
}

func filledEntries(snapshot *ports.SubjectSnapshot, entryIDs entrySet) map[string]bool {
	filled := make(map[string]bool, len(entryIDs))
	for _, entryID := range entryIDs {
		filled[entryID] = ports.ValueFilled(snapshot.Fields[entryID])
	}
	return filled
}

func (s builtInFieldSource) wrap(action string, err error) error {
	return fmt.Errorf("%s %s subjects for one-pager: %w", action, s.subjectType, err)
}
