package valueobject

type VoteType uint8

const (
	VoteTypeEmpty VoteType = iota
	VoteTypeYes
	VoteTypeNo
	VoteTypeCrush
	VoteTypeCompliment
)

var UserVoteTypeToString = map[VoteType]string{
	VoteTypeEmpty:      "empty",
	VoteTypeYes:        "yes",
	VoteTypeNo:         "no",
	VoteTypeCrush:      "crush",
	VoteTypeCompliment: "compliment",
}

func (v VoteType) IsPositive() bool {
	return v == VoteTypeYes || v == VoteTypeCrush || v == VoteTypeCompliment
}

func (v VoteType) IsNegative() bool {
	return v == VoteTypeNo
}

func (v VoteType) IsEmpty() bool {
	return v == VoteTypeEmpty
}

func (v VoteType) String() string {
	return UserVoteTypeToString[v]
}
