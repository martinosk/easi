package valueobjects

import (
	"errors"
	"strings"

	domain "easi/backend/internal/shared/eventsourcing"
)

const MaxCategoryLength = 100

var ErrCategoryTooLong = errors.New("category cannot exceed 100 characters")

type Category struct {
	value string
}

func NewCategory(value string) (Category, error) {
	trimmed := strings.TrimSpace(value)
	if len(trimmed) > MaxCategoryLength {
		return Category{}, ErrCategoryTooLong
	}
	return Category{value: trimmed}, nil
}

func EmptyCategory() Category {
	return Category{value: ""}
}

func (c Category) Value() string {
	return c.value
}

func (c Category) IsEmpty() bool {
	return c.value == ""
}

func (c Category) String() string {
	return c.value
}

func (c Category) Equals(other domain.ValueObject) bool {
	if otherCat, ok := other.(Category); ok {
		return c.value == otherCat.value
	}
	return false
}
