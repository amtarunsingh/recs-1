package entity

import (
	"github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/romance/valueobject"
	sharedValueObject "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/sharedkernel/valueobject"
	"time"
)

type Vote struct {
	Id        sharedValueObject.VoteId
	VoteType  valueobject.VoteType
	VotedAt   *time.Time
	CreatedAt *time.Time
	UpdatedAt *time.Time
}
