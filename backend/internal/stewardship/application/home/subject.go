package home

type SubjectType string

const (
	SubjectCapability  SubjectType = "capability"
	SubjectApplication SubjectType = "application"
)

type Subject struct {
	Type SubjectType
	ID   string
}

type Subjects []Subject

func (s Subjects) IDsOf(subjectType SubjectType) []string {
	ids := newIDSet()
	for _, subject := range s {
		if subject.Type == subjectType {
			ids.add(subject.ID)
		}
	}
	return ids.ids
}
