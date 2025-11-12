package operation

import (
	"context"
	romancesRepo "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/romance/repository"
	sharedValueObject "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/sharedkernel/valueobject"
)

type DeleteRomanceOperation struct {
	romancesRepository romancesRepo.RomancesRepository
}

func NewDeleteRomanceOperation(
	romancesRepository romancesRepo.RomancesRepository,
) *DeleteRomanceOperation {
	return &DeleteRomanceOperation{
		romancesRepository: romancesRepository,
	}
}

func (r *DeleteRomanceOperation) Run(ctx context.Context, voteId sharedValueObject.VoteId) error {
	return r.romancesRepository.DeleteRomance(ctx, voteId)
}
