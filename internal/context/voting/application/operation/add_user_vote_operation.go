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
	"github.com/bmbl-bumble2/recs-votes-storage/internal/shared/platform/metrics"
	"time"
)

type AddUserVoteOperation struct {
	romancesRepository romancesRepo.RomancesRepository
	countersRepository countersRepo.CountersRepository
	logger             platform.Logger
	metrics            *metrics.Metrics
}

func NewAddUserVoteOperation(
	romancesRepository romancesRepo.RomancesRepository,
	countersRepository countersRepo.CountersRepository,
	logger platform.Logger,
	metrics *metrics.Metrics,
) *AddUserVoteOperation {
	return &AddUserVoteOperation{
		romancesRepository: romancesRepository,
		countersRepository: countersRepository,
		logger:             logger,
		metrics:            metrics,
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
			r.metrics.RecordVoteError("add", "get_romance_error")
			return entity.Vote{}, err
		}

		if err := isVoteTypeChangeAllowed(romance.ActiveUserVote, voteType); err != nil {
			r.metrics.RecordVoteError("add", "invalid_transition")
			return entity.Vote{}, err
		}

		if voteType == romance.ActiveUserVote.VoteType {
			r.metrics.RecordVoteError("add", "duplicate")
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
			r.metrics.RecordVoteError("add", "db_error")
			return entity.Vote{}, err
		}

		if newVoteIsPositive && oldVoteIsNotPositive {
			r.countersRepository.IncrYesCounters(ctx, voteId, counterUpdateGroup)
		}

		if newVoteIsNegative && oldVoteIsNotNegative {
			r.countersRepository.IncrNoCounters(ctx, voteId, counterUpdateGroup)
		}

		// Record successful vote addition
		r.metrics.RecordVoteAdded(voteType.String())

		return romance.ActiveUserVote, nil
	}
}
