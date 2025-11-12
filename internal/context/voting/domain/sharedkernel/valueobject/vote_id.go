package valueobject

import (
	"errors"
	"github.com/google/uuid"
)

type VoteId struct {
	activeUserKey ActiveUserKey
	peerUserId    uuid.UUID
}

func NewVoteId(countryId uint16, activeUserId uuid.UUID, peerUserId uuid.UUID) (VoteId, error) {
	activeUserKey, err := NewActiveUserKey(countryId, activeUserId)
	if err != nil {
		return VoteId{}, err
	}

	if peerUserId == uuid.Nil {
		return VoteId{}, errors.New("peerUserId must not be empty")
	}
	if activeUserId == peerUserId {
		return VoteId{}, errors.New("activeUserId and peerUserId must differ")
	}

	return VoteId{
		activeUserKey: activeUserKey,
		peerUserId:    peerUserId,
	}, nil
}

func (id VoteId) CountryId() uint16 {
	return id.activeUserKey.countryId
}

func (id VoteId) ActiveUserId() uuid.UUID {
	return id.activeUserKey.activeUserId
}

func (id VoteId) PeerUserId() uuid.UUID {
	return id.peerUserId
}

func (id VoteId) ToPeerVoteId() VoteId {
	return VoteId{
		activeUserKey: ActiveUserKey{
			activeUserId: id.peerUserId,
			countryId:    id.activeUserKey.countryId,
		},
		peerUserId: id.activeUserKey.activeUserId,
	}
}
