package amazon_sns

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/bmbl-bumble2/recs-votes-storage/config"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/shared/platform"
	"os"
)

func GetSnsAwsConfig(config config.Config, logger platform.Logger) aws.Config {
	awsCfg, err := awsConfig.LoadDefaultConfig(context.TODO(),
		awsConfig.WithRegion(config.Aws.Region),
		awsConfig.WithBaseEndpoint(config.Aws.SnsEndpoint),
		awsConfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			config.Aws.AccessKeyId,
			config.Aws.SecretAccessKey,
			"",
		)),
	)
	if err != nil {
		logger.Error(fmt.Sprintf("Unable to load SDK config, %v", err))
		os.Exit(1)
	}
	return awsCfg
}
