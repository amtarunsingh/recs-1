package contract

import (
	"fmt"
	romancesValueObject "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/romance/valueobject"
	huma "github.com/danielgtaylor/huma/v2"
)

type AddUserVoteType romancesValueObject.VoteType

var addUserVoteTypeToString = map[romancesValueObject.VoteType]string{
	romancesValueObject.VoteTypeYes:        "yes",
	romancesValueObject.VoteTypeNo:         "no",
	romancesValueObject.VoteTypeCrush:      "crush",
	romancesValueObject.VoteTypeCompliment: "compliment",
}

var addUserVoteTypeFromString = map[string]romancesValueObject.VoteType{
	"yes":        romancesValueObject.VoteTypeYes,
	"no":         romancesValueObject.VoteTypeNo,
	"crush":      romancesValueObject.VoteTypeCrush,
	"compliment": romancesValueObject.VoteTypeCompliment,
}

func (v *AddUserVoteType) MarshalText() ([]byte, error) {
	return []byte(addUserVoteTypeToString[romancesValueObject.VoteType(*v)]), nil
}
func (v *AddUserVoteType) UnmarshalText(b []byte) error {
	vv, ok := addUserVoteTypeFromString[string(b)]
	if !ok {
		return fmt.Errorf("invalid vote type: %q", b)
	}
	*v = AddUserVoteType(vv)
	return nil
}

func (v *AddUserVoteType) Schema(r huma.Registry) *huma.Schema {
	enums := make([]any, 0, len(addUserVoteTypeToString))
	for voteType := range addUserVoteTypeToString {
		enums = append(enums, addUserVoteTypeToString[voteType])
	}

	return &huma.Schema{
		Type: huma.TypeString,
		Enum: enums,
	}
}
