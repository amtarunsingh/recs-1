package helper

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	sharedValueObject "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/sharedkernel/valueobject"
	infraDynamodb "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/infrastructure/persistence"
	platformDynamodb "github.com/bmbl-bumble2/recs-votes-storage/internal/shared/platform/dynamodb"
	"time"
)

type CountersTableHelper struct {
	ddbClient platformDynamodb.Client
}

func NewCountersTableHelper(client platformDynamodb.Client) (*CountersTableHelper, error) {
	return &CountersTableHelper{
		ddbClient: client,
	}, nil
}

func (c *CountersTableHelper) CreateCountersTable() error {
	ctx := context.Background()
	table := aws.String(infraDynamodb.CountersTableName)

	_, err := c.ddbClient.CreateTable(ctx, &dynamodb.CreateTableInput{
		TableName: table,
		AttributeDefinitions: []ddbtypes.AttributeDefinition{
			{AttributeName: aws.String(infraDynamodb.UserIdAttrName), AttributeType: ddbtypes.ScalarAttributeTypeS},
			{AttributeName: aws.String(infraDynamodb.HourUnixTimestampAttrName), AttributeType: ddbtypes.ScalarAttributeTypeN},
		},
		KeySchema: []ddbtypes.KeySchemaElement{
			{AttributeName: aws.String(infraDynamodb.UserIdAttrName), KeyType: ddbtypes.KeyTypeHash},
			{AttributeName: aws.String(infraDynamodb.HourUnixTimestampAttrName), KeyType: ddbtypes.KeyTypeRange},
		},
		BillingMode: ddbtypes.BillingModePayPerRequest,
	})

	var condCheckErr *ddbtypes.ResourceInUseException
	if err != nil && !errors.As(err, &condCheckErr) {
		return err
	}

	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		out, err := c.ddbClient.DescribeTable(ctx, &dynamodb.DescribeTableInput{TableName: table})
		if err == nil && out.Table != nil && out.Table.TableStatus == ddbtypes.TableStatusActive {
			_, err = c.ddbClient.UpdateTimeToLive(ctx, &dynamodb.UpdateTimeToLiveInput{
				TableName: table,
				TimeToLiveSpecification: &ddbtypes.TimeToLiveSpecification{
					Enabled:       aws.Bool(true),
					AttributeName: aws.String("ttl"),
				},
			})
			if err != nil {
				return fmt.Errorf("failed to enable TTL: %w", err)
			}
			return nil
		}
		time.Sleep(200 * time.Millisecond)
	}
	return fmt.Errorf("table %s not ACTIVE in time", *table)
}

func (c *CountersTableHelper) DeleteAllUserRecords(activeUserKey sharedValueObject.ActiveUserKey) error {
	ctx := context.Background()

	var lastEvaluatedKey map[string]ddbtypes.AttributeValue

	for {
		queryOutput, err := c.ddbClient.Query(ctx, &dynamodb.QueryInput{
			TableName:              aws.String(infraDynamodb.CountersTableName),
			KeyConditionExpression: aws.String(infraDynamodb.UserIdAttrName + " = :uid"),
			ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
				":uid": &ddbtypes.AttributeValueMemberS{Value: activeUserKey.ActiveUserId().String()},
			},
			ExclusiveStartKey:    lastEvaluatedKey,
			ProjectionExpression: aws.String(infraDynamodb.UserIdAttrName),
		})
		if err != nil {
			return err
		}

		if len(queryOutput.Items) == 0 {
			break
		}

		var batch []ddbtypes.WriteRequest
		for _, item := range queryOutput.Items {
			batch = append(batch, ddbtypes.WriteRequest{
				DeleteRequest: &ddbtypes.DeleteRequest{
					Key: map[string]ddbtypes.AttributeValue{
						infraDynamodb.UserIdAttrName: item[infraDynamodb.UserIdAttrName],
					},
				},
			})

			if len(batch) == 25 {
				_, err = c.ddbClient.BatchWriteItem(ctx, &dynamodb.BatchWriteItemInput{
					RequestItems: map[string][]ddbtypes.WriteRequest{
						infraDynamodb.CountersTableName: batch,
					},
				})
				if err != nil {
					return err
				}
				batch = batch[:0]
			}
		}

		if len(batch) > 0 {
			_, err = c.ddbClient.BatchWriteItem(ctx, &dynamodb.BatchWriteItemInput{
				RequestItems: map[string][]ddbtypes.WriteRequest{
					infraDynamodb.CountersTableName: batch,
				},
			})
			if err != nil {
				return err
			}
		}

		if queryOutput.LastEvaluatedKey == nil {
			break
		}
		lastEvaluatedKey = queryOutput.LastEvaluatedKey
	}

	return nil
}
