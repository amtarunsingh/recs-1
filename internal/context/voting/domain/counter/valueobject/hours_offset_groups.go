package valueobject

import (
	"fmt"
	"github.com/bmbl-bumble2/recs-votes-storage/config"
	"sort"
)

type HoursOffsetGroups struct {
	values []uint8
}

func NewHoursOffsetGroups(offsets []uint8) (HoursOffsetGroups, error) {
	err := ValidateHoursOffsets(offsets)
	if err != nil {
		return HoursOffsetGroups{}, err
	}

	cp := make([]uint8, len(offsets))
	copy(cp, offsets)
	sort.Slice(cp, func(i, j int) bool { return cp[i] < cp[j] })
	return HoursOffsetGroups{values: cp}, nil
}

func (h HoursOffsetGroups) Values() []uint8 {
	cp := make([]uint8, len(h.values))
	copy(cp, h.values)
	return cp
}

func ValidateHoursOffsets(offsets []uint8) error {
	if len(offsets) == 0 {
		return fmt.Errorf("hours offset groups cannot be empty")
	}

	seen := map[uint8]struct{}{}

	for _, h := range offsets {
		if h > config.CountersTtlHours || h < 1 {
			return fmt.Errorf("invalid hour offset: %d (must be 1-%d)", h, config.CountersTtlHours)
		}
		if _, ok := seen[h]; ok {
			return fmt.Errorf("duplicate hour offset: %d", h)
		}
		seen[h] = struct{}{}
	}

	return nil
}
