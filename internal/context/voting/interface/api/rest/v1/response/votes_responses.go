package response

import (
	"github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/romance/entity"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/interface/api/rest/v1/contract"
	"time"
)

type Vote struct {
	VoteType  contract.ReadUserVoteType `json:"vote_type"`
	VotedAt   *time.Time                `json:"voted_at" doc:"Vote time"`
	CreatedAt *time.Time                `json:"created_at" doc:"Vote creation time"`
	UpdatedAt *time.Time                `json:"updated_at" doc:"Vote update time"`
}

func NewVoteFromEntity(vote entity.Vote) Vote {
	return Vote{
		VoteType:  contract.ReadUserVoteType(vote.VoteType),
		VotedAt:   vote.VotedAt,
		CreatedAt: vote.CreatedAt,
		UpdatedAt: vote.UpdatedAt,
	}
}

type VoteGetResponse struct {
	Body Vote
}

type VoteAddResponse struct {
	Body Vote
}

type ChangeVoteResponse struct {
	Body Vote
}

func CreateVoteGetResponseFromVoteEntity(vote entity.Vote) *VoteGetResponse {
	return &VoteGetResponse{
		Body: NewVoteFromEntity(vote),
	}
}

func CreateVoteAddResponseFromVoteEntity(vote entity.Vote) *VoteAddResponse {
	return &VoteAddResponse{
		Body: NewVoteFromEntity(vote),
	}
}

func CreateChangeVoteResponseFromVoteEntity(vote entity.Vote) *ChangeVoteResponse {
	return &ChangeVoteResponse{
		Body: NewVoteFromEntity(vote),
	}
}
