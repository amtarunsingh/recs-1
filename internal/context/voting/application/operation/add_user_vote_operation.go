package operation

import (
	"context"
	"errors"
	"fmt"
	"github.com/bmbl-bumble2/recs-votes-storage/config"
	countersRepo "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/counter/repository"
	countersValueObject "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/counter/valueobject"
	romanceDomain "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/romance"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/romance/entity"
	romancesRepo "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/romance/repository"
	romancesValueObject "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/romance/valueobject"
	sharedValueObject "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/sharedkernel/valueobject"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/shared/platform"
	"time"
)

type AddUserVoteOperation struct {
	romancesRepository romancesRepo.RomancesRepository
	countersRepository countersRepo.CountersRepository
	logger             platform.Logger
}

func NewAddUserVoteOperation(
	romancesRepository romancesRepo.RomancesRepository,
	countersRepository countersRepo.CountersRepository,
	logger platform.Logger,
) *AddUserVoteOperation {
	return &AddUserVoteOperation{
		romancesRepository: romancesRepository,
		countersRepository: countersRepository,
		logger:             logger,
	}
}

func (r *AddUserVoteOperation) Run(
	ctx context.Context,
	voteId sharedValueObject.VoteId,
	voteType romancesValueObject.VoteType,
	votedAt time.Time,
) (entity.Vote, error) {
	tries := 0

	getRomanceOperation := NewGetRomanceOperation(r.romancesRepository)
	for {
		romance, err := getRomanceOperation.Run(ctx, voteId)
		if err != nil {
			r.logger.Error(fmt.Sprintf("GetRomance error: %+v", err))
			return entity.Vote{}, err
		}

		if err := isVoteTypeChangeAllowed(romance.ActiveUserVote, voteType); err != nil {
			return entity.Vote{}, err
		}

		if voteType == romance.ActiveUserVote.VoteType {
			return entity.Vote{}, romanceDomain.ErrVoteDuplicate
		}

		currentTime := time.Now()
		counterUpdateGroup, err := countersValueObject.NewCounterUpdateGroup(currentTime)
		if err != nil {
			return entity.Vote{}, err
		}

		newVoteIsPositive := voteType.IsPositive()
		newVoteIsNegative := voteType.IsNegative()
		oldVoteIsNotPositive := !romance.ActiveUserVote.VoteType.IsPositive()
		oldVoteIsNotNegative := !romance.ActiveUserVote.VoteType.IsNegative()

		romance, err = r.romancesRepository.AddActiveUserVoteToRomance(
			ctx,
			romance,
			voteType,
			votedAt,
		)

		if err != nil {
			if errors.Is(err, romanceDomain.ErrVersionConflict) && tries < config.DynamoDbVersionConflictRetriesCount {
				tries += 1
				continue
			}
			r.logger.Error(fmt.Sprintf("AddActiveUserVoteToRomance error: %+v", err))
			return entity.Vote{}, err
		}

		if newVoteIsPositive && oldVoteIsNotPositive {
			r.countersRepository.IncrYesCounters(ctx, voteId, counterUpdateGroup)
		}

		if newVoteIsNegative && oldVoteIsNotNegative {
			r.countersRepository.IncrNoCounters(ctx, voteId, counterUpdateGroup)
		}

		return romance.ActiveUserVote, nil
	}
}
