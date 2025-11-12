package operation

import (
	"context"
	"errors"
	"fmt"
	"github.com/bmbl-bumble2/recs-votes-storage/config"
	countersRepo "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/counter/repository"
	romanceDomain "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/romance"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/romance/entity"
	romancesRepo "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/romance/repository"
	romancesValueObject "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/romance/valueobject"
	sharedValueObject "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/sharedkernel/valueobject"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/shared/platform"
)

var allowedVoteTransitions = map[romancesValueObject.VoteType]map[romancesValueObject.VoteType]struct{}{
	romancesValueObject.VoteTypeEmpty: {
		romancesValueObject.VoteTypeNo:         {},
		romancesValueObject.VoteTypeYes:        {},
		romancesValueObject.VoteTypeCrush:      {},
		romancesValueObject.VoteTypeCompliment: {},
	},
	romancesValueObject.VoteTypeNo: {
		romancesValueObject.VoteTypeYes:        {},
		romancesValueObject.VoteTypeCrush:      {},
		romancesValueObject.VoteTypeCompliment: {},
	},
	romancesValueObject.VoteTypeYes: {
		romancesValueObject.VoteTypeCrush:      {},
		romancesValueObject.VoteTypeCompliment: {},
	},
	romancesValueObject.VoteTypeCrush:      {},
	romancesValueObject.VoteTypeCompliment: {},
}

type ChangeUserVoteOperation struct {
	romancesRepository romancesRepo.RomancesRepository
	countersRepository countersRepo.CountersRepository
	logger             platform.Logger
}

func NewChangeUserVoteOperation(
	romancesRepository romancesRepo.RomancesRepository,
	countersRepository countersRepo.CountersRepository,
	logger platform.Logger,
) *ChangeUserVoteOperation {
	return &ChangeUserVoteOperation{
		romancesRepository: romancesRepository,
		countersRepository: countersRepository,
		logger:             logger,
	}
}

func (r *ChangeUserVoteOperation) Run(
	ctx context.Context,
	voteId sharedValueObject.VoteId,
	newVoteType romancesValueObject.VoteType,
) (entity.Vote, error) {
	tries := 0

	getRomanceOperation := NewGetRomanceOperation(r.romancesRepository)
	for {
		romance, err := getRomanceOperation.Run(ctx, voteId)
		if err != nil {
			r.logger.Error(fmt.Sprintf("GetRomance error: %+v", err))
			return entity.Vote{}, err
		}

		if err := isVoteTypeChangeAllowed(romance.ActiveUserVote, newVoteType); err != nil {
			return entity.Vote{}, err
		}

		if newVoteType == romance.ActiveUserVote.VoteType {
			return entity.Vote{}, romanceDomain.ErrVoteDuplicate
		}

		romance, err = r.romancesRepository.ChangeActiveUserVoteTypeInRomance(
			ctx,
			romance,
			newVoteType,
		)

		if err != nil {
			if errors.Is(err, romanceDomain.ErrVersionConflict) && tries < config.DynamoDbVersionConflictRetriesCount {
				tries += 1
				continue
			}
			r.logger.Error(fmt.Sprintf("ChangeActiveUserVoteTypeInRomance error: %+v", err))
			return entity.Vote{}, err
		}

		return romance.ActiveUserVote, nil
	}
}

func isVoteTypeChangeAllowed(oldVote entity.Vote, newVoteType romancesValueObject.VoteType) error {
	allowedTargets, ok := allowedVoteTransitions[oldVote.VoteType]
	if !ok {
		return romanceDomain.NewChangingVoteTypeError(oldVote.VoteType, newVoteType)
	}

	_, allowed := allowedTargets[newVoteType]
	if !allowed {
		return romanceDomain.NewChangingVoteTypeError(oldVote.VoteType, newVoteType)
	}

	return nil
}
