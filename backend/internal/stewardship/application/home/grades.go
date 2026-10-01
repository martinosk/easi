package home

import (
	"errors"
	"fmt"
	"sort"
)

var ErrUnknownGrade = errors.New("unknown TIME grade")

const (
	GradeInvest    = "invest"
	GradeTolerate  = "tolerate"
	GradeMigrate   = "migrate"
	GradeEliminate = "eliminate"
)

type GradeCounts struct {
	Invest      int
	Tolerate    int
	Migrate     int
	Eliminate   int
	NotAssessed int
}

func (c *GradeCounts) Add(recordedGrade string, n int) error {
	slot := c.slotOf(recordedGrade)
	if slot == nil {
		return fmt.Errorf("%w: %q", ErrUnknownGrade, recordedGrade)
	}
	*slot += n
	return nil
}

func (c *GradeCounts) slotOf(recordedGrade string) *int {
	switch recordedGrade {
	case "Invest":
		return &c.Invest
	case "Tolerate":
		return &c.Tolerate
	case "Migrate":
		return &c.Migrate
	case "Eliminate":
		return &c.Eliminate
	case "":
		return &c.NotAssessed
	}
	return nil
}

func (c GradeCounts) total() int {
	return c.Invest + c.Tolerate + c.Migrate + c.Eliminate + c.NotAssessed
}

func (c GradeCounts) dominant() string {
	bySeverity := []struct {
		grade string
		count int
	}{
		{GradeEliminate, c.Eliminate},
		{GradeMigrate, c.Migrate},
		{GradeTolerate, c.Tolerate},
		{GradeInvest, c.Invest},
	}
	dominant, most := "", 0
	for _, candidate := range bySeverity {
		if candidate.count > most {
			dominant, most = candidate.grade, candidate.count
		}
	}
	return dominant
}

func timeSharesOf(c GradeCounts) *TimeShares {
	total := c.total()
	if total == 0 {
		return nil
	}
	counts := []int{c.Invest, c.Tolerate, c.Migrate, c.Eliminate, c.NotAssessed}
	shares := largestRemainder(counts, total)
	return &TimeShares{Invest: shares[0], Tolerate: shares[1], Migrate: shares[2], Eliminate: shares[3], NotAssessed: shares[4]}
}

func largestRemainder(counts []int, total int) []int {
	shares := make([]int, len(counts))
	order := make([]int, len(counts))
	assigned := 0
	for i, count := range counts {
		shares[i] = count * 100 / total
		assigned += shares[i]
		order[i] = i
	}
	remainder := func(i int) int { return counts[i] * 100 % total }
	sort.SliceStable(order, func(a, b int) bool { return remainder(order[a]) > remainder(order[b]) })
	for _, i := range order[:100-assigned] {
		shares[i]++
	}
	return shares
}
