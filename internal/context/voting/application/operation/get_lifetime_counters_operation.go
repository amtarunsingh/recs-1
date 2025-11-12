package operation

import (
	"context"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/counter/entity"
	countersRepo "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/counter/repository"
	sharedValueObject "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/sharedkernel/valueobject"
)

type GetLifetimeCountersOperation struct {
	countersRepository countersRepo.CountersRepository
}

func NewGetLifetimeCountersOperation(
	countersRepository countersRepo.CountersRepository,
) *GetLifetimeCountersOperation {
	return &GetLifetimeCountersOperation{
		countersRepository: countersRepository,
	}
}

func (r *GetLifetimeCountersOperation) Run(
	ctx context.Context,
	activeUserKey sharedValueObject.ActiveUserKey,
) (entity.CountersGroup, error) {

	counterGroup, err := r.countersRepository.GetLifetimeCounter(ctx, activeUserKey)
	if err != nil {
		return entity.CountersGroup{}, err
	}

	return counterGroup, nil
}
