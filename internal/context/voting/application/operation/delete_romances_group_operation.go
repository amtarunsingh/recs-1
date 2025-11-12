package operation

import (
	"context"
	romancesRepo "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/romance/repository"
	sharedValueObject "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/sharedkernel/valueobject"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/shared/platform"
	"github.com/google/uuid"
)

type DeleteRomancesGroupOperation struct {
	romancesRepository romancesRepo.RomancesRepository
	logger             platform.Logger
}

func NewDeleteRomancesGroupOperation(
	romancesRepository romancesRepo.RomancesRepository,
	logger platform.Logger,
) *DeleteRomancesGroupOperation {
	return &DeleteRomancesGroupOperation{
		romancesRepository: romancesRepository,
		logger:             logger,
	}
}

func (r *DeleteRomancesGroupOperation) Run(
	ctx context.Context,
	userKey sharedValueObject.ActiveUserKey,
	peerIds []uuid.UUID,
) error {
	return r.romancesRepository.DeleteRomancesGroup(ctx, userKey, peerIds)
}
