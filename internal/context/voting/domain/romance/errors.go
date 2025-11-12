package romance

import (
	"errors"
	"fmt"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/romance/valueobject"
)

var (
	ErrVoteNotFound    = errors.New("vote not found")
	ErrWrongVote       = errors.New("wrong vote")
	ErrVoteDuplicate   = errors.New("vote duplicate")
	ErrVersionConflict = errors.New("version conflict")
)

func NewChangingVoteTypeError(oldVote valueobject.VoteType, newVote valueobject.VoteType) error {
	return fmt.Errorf("%w: vote type change from `%s` to `%s` is not allowed", ErrWrongVote, oldVote, newVote)
}
