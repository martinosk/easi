package aggregates

import (
	"log"
	"time"

	"easi/backend/internal/capabilitymapping/domain/events"
	"easi/backend/internal/capabilitymapping/domain/valueobjects"
	domain "easi/backend/internal/shared/eventsourcing"
)

type ApplicationFitScore struct {
	domain.AggregateRoot
	componentID valueobjects.ComponentID
	pillarID    valueobjects.PillarID
	score       valueobjects.FitScore
	rationale   valueobjects.FitRationale
	scoredAt    time.Time
	scoredBy    valueobjects.UserIdentifier
}

type NewFitScoreParams struct {
	ComponentID valueobjects.ComponentID
	PillarID    valueobjects.PillarID
	PillarName  valueobjects.PillarName
	Score       valueobjects.FitScore
	Rationale   valueobjects.FitRationale
	ScoredBy    valueobjects.UserIdentifier
}

func SetApplicationFitScore(params NewFitScoreParams) (*ApplicationFitScore, error) {
	aggregate := &ApplicationFitScore{
		AggregateRoot: domain.NewAggregateRoot(),
	}

	event := events.NewApplicationFitScoreSet(events.ApplicationFitScoreSetParams{
		ID:          aggregate.ID(),
		ComponentID: params.ComponentID.Value(),
		PillarID:    params.PillarID.Value(),
		PillarName:  params.PillarName.Value(),
		Score:       params.Score.Value(),
		Rationale:   params.Rationale.Value(),
		ScoredBy:    params.ScoredBy.Value(),
	})

	aggregate.apply(event)
	aggregate.RaiseEvent(event)

	return aggregate, nil
}

func LoadApplicationFitScoreFromHistory(eventHistory []domain.DomainEvent) (*ApplicationFitScore, error) {
	aggregate := &ApplicationFitScore{
		AggregateRoot: domain.NewAggregateRoot(),
	}

	aggregate.LoadFromHistory(eventHistory, func(event domain.DomainEvent) {
		aggregate.apply(event)
	})

	return aggregate, nil
}

func (a *ApplicationFitScore) Update(score valueobjects.FitScore, rationale valueobjects.FitRationale, updatedBy valueobjects.UserIdentifier) error {
	event := events.NewApplicationFitScoreUpdated(events.ApplicationFitScoreUpdatedParams{
		ID:           a.ID(),
		Score:        score.Value(),
		Rationale:    rationale.Value(),
		OldScore:     a.score.Value(),
		OldRationale: a.rationale.Value(),
		UpdatedBy:    updatedBy.Value(),
	})

	a.apply(event)
	a.RaiseEvent(event)

	return nil
}

func (a *ApplicationFitScore) Remove(removedBy valueobjects.UserIdentifier) error {
	event := events.NewApplicationFitScoreRemoved(
		a.ID(),
		a.componentID.Value(),
		a.pillarID.Value(),
		removedBy.Value(),
	)

	a.apply(event)
	a.RaiseEvent(event)

	return nil
}

func (a *ApplicationFitScore) apply(event domain.DomainEvent) {
	switch e := event.(type) {
	case events.ApplicationFitScoreSet:
		a.AggregateRoot = domain.NewAggregateRootWithID(e.ID)
		if componentID, err := valueobjects.NewComponentIDFromString(e.ComponentID); err != nil {
			log.Printf("ERROR: ApplicationFitScore.apply: invalid component ID in event: %v", err)
		} else {
			a.componentID = componentID
		}
		if pillarID, err := valueobjects.NewPillarIDFromString(e.PillarID); err != nil {
			log.Printf("ERROR: ApplicationFitScore.apply: invalid pillar ID in event: %v", err)
		} else {
			a.pillarID = pillarID
		}
		if score, err := valueobjects.NewFitScore(e.Score); err != nil {
			log.Printf("ERROR: ApplicationFitScore.apply: invalid score in event: %v", err)
		} else {
			a.score = score
		}
		if rationale, err := valueobjects.NewFitRationale(e.Rationale); err != nil {
			log.Printf("ERROR: ApplicationFitScore.apply: invalid rationale in event: %v", err)
		} else {
			a.rationale = rationale
		}
		a.scoredAt = e.ScoredAt
		if scoredBy, err := valueobjects.NewUserIdentifier(e.ScoredBy); err != nil {
			log.Printf("ERROR: ApplicationFitScore.apply: invalid scoredBy in event: %v", err)
		} else {
			a.scoredBy = scoredBy
		}
	case events.ApplicationFitScoreUpdated:
		if score, err := valueobjects.NewFitScore(e.Score); err != nil {
			log.Printf("ERROR: ApplicationFitScore.apply: invalid score in update event: %v", err)
		} else {
			a.score = score
		}
		if rationale, err := valueobjects.NewFitRationale(e.Rationale); err != nil {
			log.Printf("ERROR: ApplicationFitScore.apply: invalid rationale in update event: %v", err)
		} else {
			a.rationale = rationale
		}
	case events.ApplicationFitScoreRemoved:
	}
}

func (a *ApplicationFitScore) ComponentID() valueobjects.ComponentID {
	return a.componentID
}

func (a *ApplicationFitScore) PillarID() valueobjects.PillarID {
	return a.pillarID
}

func (a *ApplicationFitScore) Score() valueobjects.FitScore {
	return a.score
}

func (a *ApplicationFitScore) Rationale() valueobjects.FitRationale {
	return a.rationale
}

func (a *ApplicationFitScore) ScoredAt() time.Time {
	return a.scoredAt
}

func (a *ApplicationFitScore) ScoredBy() valueobjects.UserIdentifier {
	return a.scoredBy
}
