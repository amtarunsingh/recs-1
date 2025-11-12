package operation

import (
	"context"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/counter/entity"
	countersRepo "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/counter/repository"
	countersValueObject "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/counter/valueobject"
	sharedValueObject "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/sharedkernel/valueobject"
)

type GetHourlyCountersOperation struct {
	countersRepository countersRepo.CountersRepository
}

func NewGetHourlyCountersOperation(
	countersRepository countersRepo.CountersRepository,
) *GetHourlyCountersOperation {
	return &GetHourlyCountersOperation{
		countersRepository: countersRepository,
	}
}

func (r *GetHourlyCountersOperation) Run(
	ctx context.Context,
	activeUserKey sharedValueObject.ActiveUserKey,
	hoursOffsetGroups countersValueObject.HoursOffsetGroups,
) (map[uint8]*entity.CountersGroup, error) {

	countersGroups, err := r.countersRepository.GetHourlyCounters(
		ctx,
		activeUserKey,
		hoursOffsetGroups,
	)
	if err != nil {
		return map[uint8]*entity.CountersGroup{}, err
	}

	return countersGroups, nil
}
