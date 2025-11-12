package bootstrap

import (
	"github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/application/messaging/handler"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/application/operation"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/shared/messaging"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/shared/platform"
)

func NewPreparedTopicHandler(
	deleteRomancesHandler *handler.DeleteRomancesHandler,
	deleteRomancesGroupHandler *handler.DeleteRomancesGroupHandler,
	logger platform.Logger,
) *messaging.TopicHandler {
	reg := messaging.NewTopicHandler(logger)
	messaging.RegisterTopicHandler(
		reg,
		operation.DeleteRomancesTopic,
		deleteRomancesHandler,
	)
	messaging.RegisterTopicHandler(
		reg,
		operation.DeleteRomancesGroupTopic,
		deleteRomancesGroupHandler,
	)

	return reg
}
