//go:build wireinject

package di

import (
	"github.com/bmbl-bumble2/recs-votes-storage/config"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/app"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/app/api"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/app/bootstrap"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/application"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/application/messaging/handler"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/application/operation"
	countersRepo "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/counter/repository"
	romancesRepo "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/romance/repository"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/infrastructure/persistence"
	storageV1 "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/interface/api/rest/v1"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/shared/messaging"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/shared/platform"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/shared/platform/amazon_sns"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/shared/platform/dynamodb"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/shared/platform/metrics"
	"github.com/google/wire"
)

var PlatformSet = wire.NewSet(
	platform.NewLogger,
)

var MetricsSet = wire.NewSet(
	metrics.NewRegistry,
	metrics.NewMetrics,
)

var ReposSet = wire.NewSet(
	dynamodb.NewDynamoDbClient,
	persistence.NewRomancesRepository,
	persistence.NewCountersRepository,
	wire.Bind(new(romancesRepo.RomancesRepository), new(*persistence.RomancesRepository)),
	wire.Bind(new(countersRepo.CountersRepository), new(*persistence.CountersRepository)),
)

var OperationsSet = wire.NewSet(
	operation.NewGetRomanceOperation,
	operation.NewDeleteRomanceOperation,
	operation.NewGetUserVoteOperation,
	operation.NewAddUserVoteOperation,
	operation.NewChangeUserVoteOperation,
	operation.NewDeleteUserVoteOperation,
	operation.NewGetLifetimeCountersOperation,
	operation.NewGetHourlyCountersOperation,
	operation.NewDeleteRomancesRequestOperation,
	operation.NewDeleteRomancesOperation,
	operation.NewDeleteRomancesGroupOperation,
	application.NewVotingService,
)

func InitializeApiWebServer(config config.Config) (*app.ApiWebServer, error) {
	wire.Build(
		PlatformSet,
		MetricsSet,
		ReposSet,
		amazon_sns.NewSnsPublisher,
		wire.Bind(new(messaging.Publisher), new(*amazon_sns.SnsPublisher)),
		OperationsSet,
		storageV1.NewVotesStorageRoutesRegister,
		api.NewHandlerFactory,
		app.NewApiWebServer,
	)
	return nil, nil
}

func InitializeMessageProcessor(config config.Config) (*app.MessageProcessor, error) {
	wire.Build(
		PlatformSet,
		MetricsSet,
		ReposSet,
		amazon_sns.NewSnsSubscriber,
		amazon_sns.NewSnsPublisher,
		wire.Bind(new(messaging.Subscriber), new(*amazon_sns.SnsSubscriber)),
		wire.Bind(new(messaging.Publisher), new(*amazon_sns.SnsPublisher)),
		handler.NewDeleteRomancesHandler,
		handler.NewDeleteRomancesGroupHandler,
		OperationsSet,
		bootstrap.NewPreparedTopicHandler,
		app.NewTopicListener,
		app.NewMessageProcessor,
	)
	return nil, nil
}
