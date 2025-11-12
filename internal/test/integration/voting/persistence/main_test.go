package persistence

import (
	"context"
	"github.com/bmbl-bumble2/recs-votes-storage/config"
	platformDynamodb "github.com/bmbl-bumble2/recs-votes-storage/internal/shared/platform/dynamodb"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/testlib/testcontainer"
	"log"
	"os"
	"testing"
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
