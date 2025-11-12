package operation

import (
	"context"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/romance/entity"
	romancesRepo "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/romance/repository"
	sharedValueObject "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/sharedkernel/valueobject"
)

type GetRomanceOperation struct {
	romancesRepository romancesRepo.RomancesRepository
}

func NewGetRomanceOperation(
	romancesRepository romancesRepo.RomancesRepository,
) *GetRomanceOperation {
	return &GetRomanceOperation{
		romancesRepository: romancesRepository,
	}
}

func (r *GetRomanceOperation) Run(ctx context.Context, voteId sharedValueObject.VoteId) (entity.Romance, error) {
	return r.romancesRepository.GetRomance(ctx, voteId)
}
