package command

import (
	"github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/interface/api/rest/v1/contract"
	"github.com/google/uuid"
	"time"
)

type VoteAdd struct {
	CountryId uint16 `path:"country_id" doc:"Current active user country ID"`
	Body      struct {
		ActiveUserId uuid.UUID                `json:"active_user_id" format:"uuid" doc:"Active User Id"`
		PeerId       uuid.UUID                `json:"peer_id" format:"uuid" doc:"Peer user ID"`
		VoteType     contract.AddUserVoteType `json:"vote_type"`
		VotedAt      time.Time                `json:"voted_at"`
	}
}

type ChangeVoteType struct {
	CountryId    uint16    `path:"country_id" doc:"Current active user country ID"`
	ActiveUserId uuid.UUID `path:"active_user_id" format:"uuid" doc:"Active User Id"`
	PeerId       uuid.UUID `path:"peer_id" format:"uuid" doc:"Peer user ID"`
	Body         struct {
		NewType contract.ChangeUserVoteType `json:"new_vote_type"`
	}
}

type DeleteVote struct {
	CountryId    uint16    `path:"country_id" doc:"Current active user country ID"`
	ActiveUserId uuid.UUID `path:"active_user_id" format:"uuid" doc:"Active User Id"`
	PeerId       uuid.UUID `path:"peer_id" format:"uuid" doc:"Peer user ID"`
}
