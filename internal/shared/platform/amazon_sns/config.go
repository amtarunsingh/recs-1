package amazon_sns

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/bmbl-bumble2/recs-votes-storage/config"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/shared/platform"
)

func GetSnsAwsConfig(config config.Config, logger platform.Logger) aws.Config {
	opts := []func(*awsConfig.LoadOptions) error{
		awsConfig.WithRegion(config.Aws.Region),
	}
	if config.Aws.AccessKeyId != "" && config.Aws.SecretAccessKey != "" {
		opts = append(opts, awsConfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(config.Aws.AccessKeyId, config.Aws.SecretAccessKey, "")))
	}
	if config.Aws.SnsEndpoint != "" {
		opts = append(opts, awsConfig.WithBaseEndpoint(config.Aws.SnsEndpoint))
	}
	awsCfg, err := awsConfig.LoadDefaultConfig(context.TODO(), opts...)
	if err != nil {
		logger.Error(fmt.Sprintf("Unable to load SDK config, %v", err))
		os.Exit(1)
	}
	return awsCfg
}
