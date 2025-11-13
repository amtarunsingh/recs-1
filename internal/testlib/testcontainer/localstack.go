package testcontainer

import (
	"context"
	"fmt"
	"os"
	"sync"

	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/docker/go-connections/nat"
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
			// Fallback: manually compose endpoint from mapped edge ports
			host, hostErr := container.Host(ctx)
			if hostErr != nil {
				_ = container.Terminate(ctx)
				localstackInitErr = fmt.Errorf("failed to get localstack host: %w (endpoint error: %v)", hostErr, err)
				return
			}
			var mapped nat.Port
			candidatePorts := []string{"4566/tcp", "4510/tcp"}
			for _, p := range candidatePorts {
				if mp, mapErr := container.MappedPort(ctx, nat.Port(p)); mapErr == nil && mp.Port() != "" {
					mapped = mp
					break
				}
			}
			if mapped.Port() == "" {
				_ = container.Terminate(ctx)
				localstackInitErr = fmt.Errorf("failed to get localstack endpoint: %w (no mapped edge port found among %v)", err, candidatePorts)
				return
			}
			endpoint = fmt.Sprintf("http://%s:%s", host, mapped.Port())
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
