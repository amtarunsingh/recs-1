package dynamodb

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	appConfig "github.com/bmbl-bumble2/recs-votes-storage/config"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/shared/platform"
	"os"
)

//go:generate mockgen -destination=../../../testlib/mocks/dynamodb_client_mock.go -package=mocks github.com/bmbl-bumble2/recs-votes-storage/internal/shared/platform/dynamodb Client
type Client interface {
	CreateTable(ctx context.Context, params *dynamodb.CreateTableInput, optFns ...func(*dynamodb.Options)) (*dynamodb.CreateTableOutput, error)
	DescribeTable(ctx context.Context, params *dynamodb.DescribeTableInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DescribeTableOutput, error)
	UpdateTimeToLive(ctx context.Context, params *dynamodb.UpdateTimeToLiveInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateTimeToLiveOutput, error)
	PutItem(ctx context.Context, in *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
	GetItem(ctx context.Context, in *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
	UpdateItem(ctx context.Context, in *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error)
	DeleteItem(ctx context.Context, in *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error)
	Query(ctx context.Context, in *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error)
	TransactWriteItems(ctx context.Context, in *dynamodb.TransactWriteItemsInput, optFns ...func(*dynamodb.Options)) (*dynamodb.TransactWriteItemsOutput, error)
	BatchWriteItem(ctx context.Context, params *dynamodb.BatchWriteItemInput, optFns ...func(options *dynamodb.Options)) (*dynamodb.BatchWriteItemOutput, error)
}

func NewDynamoDbClient(conf appConfig.Config, logger platform.Logger) Client {
	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithRegion(conf.Aws.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			conf.Aws.AccessKeyId,
			conf.Aws.SecretAccessKey,
			"",
		)),
		config.WithBaseEndpoint(conf.Aws.DynamoDbEndpoint),
	)
	if err != nil {
		logger.Error(fmt.Sprintf("Unable to load SDK config, %v", err))
		os.Exit(1)
	}

	return dynamodb.NewFromConfig(cfg)
}

func GetDynamodbRegionByCountry(countryId uint16) string {
	return "us-east-2"
}
