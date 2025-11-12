package query

import (
	"fmt"
	counterValueObject "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/counter/valueobject"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/interface/api/rest/v1/contract"
	huma "github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
)

type LifetimeCountersGet struct {
	CountryId    uint16    `path:"country_id" doc:"Current active user country ID"`
	ActiveUserId uuid.UUID `path:"active_user_id" format:"uuid" doc:"Active User Id"`
}

type HourlyCountersGet struct {
	CountryId            uint16                       `path:"country_id" doc:"Current active user country ID"`
	ActiveUserId         uuid.UUID                    `path:"active_user_id" format:"uuid" doc:"Active User Id"`
	HoursOffsetGroupsRaw contract.NonNullIntArrayType `query:"hours_offset_groups" required:"true" minItems:"1" example:"[12,24]" maxItems:"10" doc:"Specifies the hours for which counters need to be returned"`
	HoursOffsetGroups    []uint8                      `json:"-"`
}

func (in *HourlyCountersGet) Resolve(ctx huma.Context, prefix *huma.PathBuffer) []error {
	offsets := make([]uint8, len(in.HoursOffsetGroupsRaw))
	location := prefix.With("query.hours_offset_groups")

	for i, p := range in.HoursOffsetGroupsRaw {
		if int(uint8(p)) != p {
			return []error{&huma.ErrorDetail{
				Location: location,
				Message:  fmt.Sprintf("The value %d is greater than the maximum value of uint8", p),
				Value:    in.HoursOffsetGroupsRaw,
			}}
		}
		offsets[i] = uint8(p)
	}

	err := counterValueObject.ValidateHoursOffsets(offsets)
	if err != nil {
		return []error{&huma.ErrorDetail{
			Location: location,
			Message:  err.Error(),
			Value:    in.HoursOffsetGroupsRaw,
		}}
	}
	in.HoursOffsetGroups = offsets
	return nil
}
