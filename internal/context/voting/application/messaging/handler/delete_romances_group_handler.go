package handler

import (
	"context"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/application"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/application/messaging/message"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/sharedkernel/valueobject"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/shared/platform"
)

type DeleteRomancesGroupHandler struct {
	name          string
	votingService *application.VotingService
	logger        platform.Logger
}

func NewDeleteRomancesGroupHandler(
	votingService *application.VotingService,
	logger platform.Logger,
) *DeleteRomancesGroupHandler {
	return &DeleteRomancesGroupHandler{
		name:          string(DeleteRomancesGroupHandlerName),
		votingService: votingService,
		logger:        logger,
	}
}

func (h *DeleteRomancesGroupHandler) GetName() string {
	return h.name
}

func (h *DeleteRomancesGroupHandler) Handle(ctx context.Context, message *message.DeleteRomancesGroupMessage) error {
	userKey, err := valueobject.NewActiveUserKey(message.CountryId, message.ActiveUserId)
	if err != nil {
		return err
	}

	return h.votingService.DeleteRomancesGroup(ctx, userKey, message.PeerIds)
}
