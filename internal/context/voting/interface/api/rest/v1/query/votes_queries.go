package query

import "github.com/google/uuid"

type VoteGet struct {
	CountryId    uint16    `path:"country_id" doc:"Current active user country ID"`
	ActiveUserId uuid.UUID `path:"active_user_id" format:"uuid" doc:"Active User Id"`
	PeerId       uuid.UUID `path:"peer_id" format:"uuid" doc:"Peer user ID"`
}
