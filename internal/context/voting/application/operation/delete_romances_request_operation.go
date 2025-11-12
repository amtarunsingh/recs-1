package operation

import (
	"context"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/application/messaging/message"
	sharedValueObject "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/sharedkernel/valueobject"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/shared/messaging"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/shared/platform"
)

const DeleteRomancesTopic = messaging.Topic("delete-romances.fifo")

type DeleteRomancesRequestOperation struct {
	publisher messaging.Publisher
	logger    platform.Logger
}

func NewDeleteRomancesRequestOperation(
	publisher messaging.Publisher,
	logger platform.Logger,
) *DeleteRomancesRequestOperation {
	return &DeleteRomancesRequestOperation{
		publisher: publisher,
		logger:    logger,
	}
}

func (r *DeleteRomancesRequestOperation) Run(ctx context.Context, userKey sharedValueObject.ActiveUserKey) error {
	return r.publisher.Publish(DeleteRomancesTopic, message.NewDeleteRomancesMessage(userKey))
}
