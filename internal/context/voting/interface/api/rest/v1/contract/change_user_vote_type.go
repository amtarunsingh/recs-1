package contract

import (
	"fmt"
	romancesValueObject "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/romance/valueobject"
	huma "github.com/danielgtaylor/huma/v2"
)

type ChangeUserVoteType romancesValueObject.VoteType

var changeUserVoteTypeToString = map[romancesValueObject.VoteType]string{
	romancesValueObject.VoteTypeYes:        "yes",
	romancesValueObject.VoteTypeCrush:      "crush",
	romancesValueObject.VoteTypeCompliment: "compliment",
}

var changeUserVoteTypeFromString = map[string]romancesValueObject.VoteType{
	"yes":        romancesValueObject.VoteTypeYes,
	"crush":      romancesValueObject.VoteTypeCrush,
	"compliment": romancesValueObject.VoteTypeCompliment,
}

func (v *ChangeUserVoteType) MarshalText() ([]byte, error) {
	return []byte(changeUserVoteTypeToString[romancesValueObject.VoteType(*v)]), nil
}
func (v *ChangeUserVoteType) UnmarshalText(b []byte) error {
	vv, ok := changeUserVoteTypeFromString[string(b)]
	if !ok {
		return fmt.Errorf("invalid vote type: %q", b)
	}
	*v = ChangeUserVoteType(vv)
	return nil
}

func (v *ChangeUserVoteType) Schema(r huma.Registry) *huma.Schema {
	enums := make([]any, 0, len(changeUserVoteTypeToString))
	for voteType := range changeUserVoteTypeToString {
		enums = append(enums, changeUserVoteTypeToString[voteType])
	}

	return &huma.Schema{
		Type: huma.TypeString,
		Enum: enums,
	}
}
