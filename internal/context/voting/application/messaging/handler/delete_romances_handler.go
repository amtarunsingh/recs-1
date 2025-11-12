package handler

import (
	"context"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/application"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/application/messaging/message"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/interface/api/rest/v1/command"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/shared/platform"
)

type DeleteRomancesHandler struct {
	name          string
	votingService *application.VotingService
	logger        platform.Logger
}

func NewDeleteRomancesHandler(
	votingService *application.VotingService,
	logger platform.Logger,
) *DeleteRomancesHandler {
	return &DeleteRomancesHandler{
		name:          string(DeleteRomancesHandlerName),
		votingService: votingService,
		logger:        logger,
	}
}

func (h *DeleteRomancesHandler) GetName() string {
	return h.name
}

func (h *DeleteRomancesHandler) Handle(ctx context.Context, message *message.DeleteRomancesMessage) error {
	c := command.DeleteRomances{
		ActiveUserId: message.ActiveUserId,
		CountryId:    message.CountryId,
	}

	err := h.votingService.DeleteRomances(ctx, c)
	if err != nil {
		return err
	}
	return nil
}
