package valueobject

import (
	"errors"
	"github.com/google/uuid"
)

type ActiveUserKey struct {
	countryId    uint16
	activeUserId uuid.UUID
}

func NewActiveUserKey(countryId uint16, activeUserId uuid.UUID) (ActiveUserKey, error) {
	if countryId == 0 {
		return ActiveUserKey{}, errors.New("countryId must be non-zero")
	}
	if activeUserId == uuid.Nil {
		return ActiveUserKey{}, errors.New("activeUserId must not be empty")
	}

	return ActiveUserKey{
		countryId:    countryId,
		activeUserId: activeUserId,
	}, nil
}

func (id ActiveUserKey) CountryId() uint16 {
	return id.countryId
}

func (id ActiveUserKey) ActiveUserId() uuid.UUID {
	return id.activeUserId
}
