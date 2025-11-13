package testcontainer

import (
	"context"
	"fmt"
	"os"
	"sync"

	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/localstack"
)

type LocalStackContainer struct {
	Container *localstack.LocalStackContainer
	Client    *dynamodb.Client
	Endpoint  string
}

var (
	localstackOnce     sync.Once
	localstackInitErr  error
	localstackInstance *LocalStackContainer
)

func SetupLocalStack(ctx context.Context, region string) (*LocalStackContainer, error) {
	localstackOnce.Do(func() {
		container, err := localstack.Run(ctx,
			"localstack/localstack:4.9.0",
			testcontainers.WithEnv(map[string]string{
				"SERVICES":   "dynamodb",
				"DEBUG":      "0",
				"LS_LOG":     "error",
				"AWS_REGION": region,
			}),
		)
		if err != nil {
			localstackInitErr = fmt.Errorf("failed to start localstack container: %w", err)
			return
		}

		endpoint, err := container.Endpoint(ctx, "http")
		if err != nil {
			_ = container.Terminate(ctx)
			localstackInitErr = fmt.Errorf("failed to get localstack endpoint: %w", err)
			return
		}

		cfg, err := awsConfig.LoadDefaultConfig(
			ctx,
			awsConfig.WithRegion(region),
			awsConfig.WithBaseEndpoint(endpoint),
			awsConfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("dummy", "dummy", "")),
		)
		if err != nil {
			_ = container.Terminate(ctx)
			localstackInitErr = fmt.Errorf("failed to create aws config: %w", err)
			return
		}

		localstackInstance = &LocalStackContainer{
			Container: container,
			Client:    dynamodb.NewFromConfig(cfg),
			Endpoint:  endpoint,
		}
	})

	if localstackInitErr != nil {
		fmt.Fprintf(os.Stderr, "LocalStack initialization failed: %v\n", localstackInitErr)
		os.Exit(1)
	}

	return localstackInstance, nil
}
