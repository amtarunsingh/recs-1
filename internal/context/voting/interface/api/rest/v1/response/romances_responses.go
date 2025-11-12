package response

import (
	"github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/romance/entity"
)

type Romance struct {
	ActiveUserVote Vote `json:"active_user_vote" doc:"Active user vote"`
	PeerUserVote   Vote `json:"peer_vote" doc:"Peer user vote"`
}

type RomanceGetResponse struct {
	Body Romance
}

func CreateRomanceGetResponseFromVoteEntity(vote entity.Romance) *RomanceGetResponse {
	resp := &RomanceGetResponse{
		Body: Romance{
			ActiveUserVote: NewVoteFromEntity(vote.ActiveUserVote),
			PeerUserVote:   NewVoteFromEntity(vote.PeerUserVote),
		},
	}
	return resp
}
