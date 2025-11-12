package message

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/sharedkernel/valueobject"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/shared/messaging"
	"github.com/google/uuid"
)

const delRomancesGroupMessageName = "del_romances_group"

type DeleteRomancesGroupMessage struct {
	ActiveUserId uuid.UUID   `json:"active_user_id"`
	CountryId    uint16      `json:"country_id"`
	PeerIds      []uuid.UUID `json:"peer_ids"`
}

func NewDeleteRomancesGroupMessage(activeUserKey valueobject.ActiveUserKey, peerIds []uuid.UUID) *DeleteRomancesGroupMessage {
	return &DeleteRomancesGroupMessage{
		ActiveUserId: activeUserKey.ActiveUserId(),
		PeerIds:      peerIds,
		CountryId:    activeUserKey.CountryId(),
	}
}

func (m *DeleteRomancesGroupMessage) GetDeduplicationId() string {
	hashBytes := md5.Sum([]byte(fmt.Sprintf("%s_%d_%v", m.ActiveUserId.String(), m.CountryId, m.PeerIds)))
	return hex.EncodeToString(hashBytes[:])
}

func (m *DeleteRomancesGroupMessage) GetPayload() messaging.Payload {
	payload, err := MarshalMessage(delRomancesGroupMessageName, m)
	if err != nil {
		return nil
	}
	return payload
}

func (m *DeleteRomancesGroupMessage) Load(payload messaging.Payload) error {
	tmp, err := UnmarshalMessage[*DeleteRomancesGroupMessage](payload, delRomancesGroupMessageName)
	if err != nil {
		return err
	}
	*m = *tmp
	return nil
}
