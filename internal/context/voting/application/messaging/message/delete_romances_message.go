package message

import (
	"fmt"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/sharedkernel/valueobject"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/shared/messaging"
	"github.com/google/uuid"
)

const delRomancesMessageName = "del_romances"

type DeleteRomancesMessage struct {
	ActiveUserId uuid.UUID `json:"active_user_id"`
	CountryId    uint16    `json:"country_id"`
}

func NewDeleteRomancesMessage(activeUserKey valueobject.ActiveUserKey) *DeleteRomancesMessage {
	return &DeleteRomancesMessage{
		ActiveUserId: activeUserKey.ActiveUserId(),
		CountryId:    activeUserKey.CountryId(),
	}
}

func (m *DeleteRomancesMessage) GetDeduplicationId() string {
	return fmt.Sprintf("%s_%d", m.ActiveUserId.String(), m.CountryId)
}

func (m *DeleteRomancesMessage) GetPayload() messaging.Payload {
	payload, err := MarshalMessage(delRomancesMessageName, m)
	if err != nil {
		return nil
	}
	return payload
}

func (m *DeleteRomancesMessage) Load(payload messaging.Payload) error {
	tmp, err := UnmarshalMessage[*DeleteRomancesMessage](payload, delRomancesMessageName)
	if err != nil {
		return err
	}

	*m = *tmp
	return nil
}
