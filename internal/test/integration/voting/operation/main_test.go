package operation

import (
	"context"
	"io"
	"log"
	"log/slog"
	"os"
	"testing"

	"github.com/bmbl-bumble2/recs-votes-storage/config"
	counterRepository "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/counter/repository"
	romanceRepository "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/romance/repository"
	infraDynamodb "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/infrastructure/persistence"
	platformDynamodb "github.com/bmbl-bumble2/recs-votes-storage/internal/shared/platform/dynamodb"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/testlib/testcontainer"
)

var (
	ddbClient platformDynamodb.Client
)

var appConfig = config.Load()

func TestMain(m *testing.M) {
	ctx := context.Background()
	localstack, err := testcontainer.SetupLocalStack(ctx, appConfig.Aws.Region)
	if err != nil {
		log.Fatalf("failed to run localstack: %v", err)
	}
	ddbClient = localstack.Client

	// Run tests and ensure cleanup happens before os.Exit
	code := m.Run()

	// Ensure LocalStack container is properly terminated after all tests
	if err := localstack.Container.Terminate(ctx); err != nil {
		log.Printf("failed to terminate localstack: %v", err)
	}

	os.Exit(code)
}

func newRomancesRepository(client platformDynamodb.Client) romanceRepository.RomancesRepository {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return infraDynamodb.NewRomancesRepository(client, appConfig, logger)
}

func newCountersRepository(client platformDynamodb.Client) counterRepository.CountersRepository {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return infraDynamodb.NewCountersRepository(client, appConfig, logger)
}
