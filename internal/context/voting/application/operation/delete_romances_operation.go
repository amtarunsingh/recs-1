package operation

import (
	"context"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/application/messaging/message"
	romancesRepo "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/romance/repository"
	sharedValueObject "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/sharedkernel/valueobject"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/shared/messaging"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/shared/platform"
	"github.com/google/uuid"
)

const DeleteRomancesGroupTopic = messaging.Topic("delete-romances-group.fifo")
const getRomancesGroupLimit = 25

type DeleteRomancesOperation struct {
	romancesRepository romancesRepo.RomancesRepository
	publisher          messaging.Publisher
	logger             platform.Logger
}

func NewDeleteRomancesOperation(
	romancesRepository romancesRepo.RomancesRepository,
	publisher messaging.Publisher,
	logger platform.Logger,
) *DeleteRomancesOperation {
	return &DeleteRomancesOperation{
		romancesRepository: romancesRepository,
		publisher:          publisher,
		logger:             logger,
	}
}

func (r *DeleteRomancesOperation) Run(ctx context.Context, userKey sharedValueObject.ActiveUserKey) error {
	var peerIds []uuid.UUID

	peerIdsChan, err := r.romancesRepository.GetAllPeersForActiveUser(ctx, userKey)
	if err != nil {
		r.logger.Error(err.Error())
		return err
	}
	peerIds = []uuid.UUID{}

	for peerId := range peerIdsChan {
		peerIds = append(peerIds, peerId)
		if len(peerIds) == getRomancesGroupLimit {
			err = r.publisher.Publish(DeleteRomancesGroupTopic, message.NewDeleteRomancesGroupMessage(userKey, peerIds))
			if err != nil {
				r.logger.Error(err.Error())
				return err
			}
			peerIds = []uuid.UUID{}
		}
	}

	if len(peerIds) > 0 {
		err = r.publisher.Publish(DeleteRomancesGroupTopic, message.NewDeleteRomancesGroupMessage(userKey, peerIds))
		if err != nil {
			r.logger.Error(err.Error())
			return err
		}
	}

	return nil
}
