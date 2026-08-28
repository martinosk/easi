package entities

import "errors"

var (
	ErrMilestoneNotFound        = errors.New("milestone not found on journey")
	ErrMilestoneOrderIncomplete = errors.New("milestone order must list every milestone exactly once")
	ErrMilestoneOrderDuplicate  = errors.New("milestone order must not repeat a milestone")
)

type Milestones struct {
	items []Milestone
}

func NoMilestones() Milestones {
	return Milestones{items: []Milestone{}}
}

func (ms Milestones) List() []Milestone {
	out := make([]Milestone, len(ms.items))
	copy(out, ms.items)
	return out
}

func (ms Milestones) Count() int { return len(ms.items) }

func (ms Milestones) Has(id string) bool {
	return ms.indexOf(id) >= 0
}

func (ms Milestones) Record(milestone Milestone) Milestones {
	items := ms.List()
	if i := ms.indexOf(milestone.ID()); i >= 0 {
		items[i] = milestone
		return Milestones{items: items}
	}
	return Milestones{items: append(items, milestone)}
}

func (ms Milestones) Remove(id string) Milestones {
	items := make([]Milestone, 0, len(ms.items))
	for _, m := range ms.items {
		if m.ID() != id {
			items = append(items, m)
		}
	}
	return Milestones{items: items}
}

func (ms Milestones) ValidateSequence(ids []string) error {
	if len(ids) != len(ms.items) {
		return ErrMilestoneOrderIncomplete
	}
	seen := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		if _, dup := seen[id]; dup {
			return ErrMilestoneOrderDuplicate
		}
		if !ms.Has(id) {
			return ErrMilestoneNotFound
		}
		seen[id] = struct{}{}
	}
	return nil
}

func (ms Milestones) InSequence(ids []string) bool {
	if len(ids) != len(ms.items) {
		return false
	}
	for i, m := range ms.items {
		if m.ID() != ids[i] {
			return false
		}
	}
	return true
}

func (ms Milestones) Reorder(ids []string) (Milestones, error) {
	if err := ms.ValidateSequence(ids); err != nil {
		return Milestones{}, err
	}
	items := make([]Milestone, len(ids))
	for i, id := range ids {
		items[i] = ms.items[ms.indexOf(id)]
	}
	return Milestones{items: items}, nil
}

func (ms Milestones) indexOf(id string) int {
	for i, m := range ms.items {
		if m.ID() == id {
			return i
		}
	}
	return -1
}
