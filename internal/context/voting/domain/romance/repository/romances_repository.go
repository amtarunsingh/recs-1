package repository

import (
	"context"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/romance/entity"
	romancesValueObject "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/romance/valueobject"
	sharedValueObject "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/sharedkernel/valueobject"
	"github.com/google/uuid"
	"time"
)

//go:generate mockgen -destination=../../../../../testlib/mocks/romances_repository_mock.go -package=mocks github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/romance/repository RomancesRepository
type RomancesRepository interface {
	GetRomance(ctx context.Context, voteId sharedValueObject.VoteId) (entity.Romance, error)
	GetAllPeersForActiveUser(ctx context.Context, activeUserKey sharedValueObject.ActiveUserKey) (<-chan uuid.UUID, error)
	DeleteRomance(ctx context.Context, voteId sharedValueObject.VoteId) error
	DeleteRomancesGroup(
		ctx context.Context,
		userKey sharedValueObject.ActiveUserKey,
		peerIds []uuid.UUID,
	) error
	AddActiveUserVoteToRomance(
		ctx context.Context,
		romance entity.Romance,
		voteType romancesValueObject.VoteType,
		votedAt time.Time,
	) (entity.Romance, error)
	ChangeActiveUserVoteTypeInRomance(
		ctx context.Context,
		romance entity.Romance,
		newVoteType romancesValueObject.VoteType,
	) (entity.Romance, error)
	DeleteActiveUserVoteFromRomance(ctx context.Context, romance entity.Romance) error
}
