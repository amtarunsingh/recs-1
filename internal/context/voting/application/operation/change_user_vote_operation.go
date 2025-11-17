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
	"github.com/bmbl-bumble2/recs-votes-storage/internal/shared/platform/metrics"
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
	metrics            *metrics.Metrics
}

func NewChangeUserVoteOperation(
	romancesRepository romancesRepo.RomancesRepository,
	countersRepository countersRepo.CountersRepository,
	logger platform.Logger,
	metrics *metrics.Metrics,
) *ChangeUserVoteOperation {
	return &ChangeUserVoteOperation{
		romancesRepository: romancesRepository,
		countersRepository: countersRepository,
		logger:             logger,
		metrics:            metrics,
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
			r.metrics.RecordVoteError("change", "get_romance_error")
			return entity.Vote{}, err
		}

		oldVoteType := romance.ActiveUserVote.VoteType

		if err := isVoteTypeChangeAllowed(romance.ActiveUserVote, newVoteType); err != nil {
			r.metrics.RecordVoteError("change", "invalid_transition")
			return entity.Vote{}, err
		}

		if newVoteType == romance.ActiveUserVote.VoteType {
			r.metrics.RecordVoteError("change", "duplicate")
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
			r.metrics.RecordVoteError("change", "db_error")
			return entity.Vote{}, err
		}

		// Record successful vote change
		r.metrics.RecordVoteChanged(oldVoteType.String(), newVoteType.String())

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
