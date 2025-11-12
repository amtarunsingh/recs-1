package contract

import (
	romancesValueObject "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/romance/valueobject"
	huma "github.com/danielgtaylor/huma/v2"
)

type ReadUserVoteType romancesValueObject.VoteType

func (v *ReadUserVoteType) MarshalText() ([]byte, error) {
	return []byte(romancesValueObject.UserVoteTypeToString[romancesValueObject.VoteType(*v)]), nil
}

func (v *ReadUserVoteType) Schema(r huma.Registry) *huma.Schema {
	enums := make([]any, 0, len(romancesValueObject.UserVoteTypeToString))
	for voteType := range romancesValueObject.UserVoteTypeToString {
		enums = append(enums, romancesValueObject.UserVoteTypeToString[voteType])
	}

	return &huma.Schema{
		Type: huma.TypeString,
		Enum: enums,
	}
}
