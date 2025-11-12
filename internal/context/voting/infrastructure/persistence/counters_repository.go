package persistence

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/bmbl-bumble2/recs-votes-storage/config"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/counter/entity"
	countersValueObject "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/counter/valueobject"
	sharedValueObject "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/sharedkernel/valueobject"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/shared/platform"
	platformDynamoDb "github.com/bmbl-bumble2/recs-votes-storage/internal/shared/platform/dynamodb"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/shared/timeutil"
	"github.com/google/uuid"
	"strconv"
	"time"
)

const (
	CountersTableName         = "Counters"
	LifetimeCounterKey        = 0
	UserIdAttrName            = "u"
	HourUnixTimestampAttrName = "h"
	incomingYesAttrName       = "iy"
	incomingNoAttrName        = "in"
	outgoingYesAttrName       = "oy"
	outgoingNoAttrName        = "on"
)

type CountersRepository struct {
	dynamoDbClient platformDynamoDb.Client
	config         config.Config
	logger         platform.Logger
}

type CountersDocumentSchema struct {
	UserId            string `dynamodbav:"u"`
	HourUnixTimestamp int32  `dynamodbav:"h"`
	IncomingYes       uint32 `dynamodbav:"iy"`
	IncomingNo        uint32 `dynamodbav:"in"`
	OutgoingYes       uint32 `dynamodbav:"oy"`
	OutgoingNo        uint32 `dynamodbav:"on"`
}

func NewCountersRepository(
	dynamoDbClient platformDynamoDb.Client,
	config config.Config,
	logger platform.Logger,
) *CountersRepository {
	return &CountersRepository{
		dynamoDbClient: dynamoDbClient,
		config:         config,
		logger:         logger,
	}
}

func (c *CountersRepository) GetLifetimeCounter(
	ctx context.Context,
	activeUserKey sharedValueObject.ActiveUserKey,
) (entity.CountersGroup, error) {
	out, err := c.dynamoDbClient.GetItem(ctx, &dynamodb.GetItemInput{
		Key:            c.getCountersTableKey(activeUserKey.ActiveUserId(), LifetimeCounterKey),
		TableName:      aws.String(CountersTableName),
		ConsistentRead: aws.Bool(true),
	}, func(o *dynamodb.Options) {
		o.Region = platformDynamoDb.GetDynamodbRegionByCountry(activeUserKey.CountryId())
	})

	if err != nil {
		return entity.CountersGroup{}, err
	}

	c.logger.Debug(fmt.Sprintf("Got counter from dynamodb: %+v", out))

	if len(out.Item) == 0 {
		return entity.CountersGroup{}, nil
	}

	countersGroupItem := &CountersDocumentSchema{}
	if err = attributevalue.UnmarshalMap(out.Item, countersGroupItem); err != nil {
		return entity.CountersGroup{}, err
	}

	return c.transformCountersGroupItemToEntity(activeUserKey.CountryId(), *countersGroupItem)
}

func (c *CountersRepository) GetHourlyCounters(
	ctx context.Context,
	activeUserKey sharedValueObject.ActiveUserKey,
	hoursOffsetGroups countersValueObject.HoursOffsetGroups,
) (map[uint8]*entity.CountersGroup, error) {

	result := map[uint8]*entity.CountersGroup{}
	maxHour := uint8(0)

	for _, hour := range hoursOffsetGroups.Values() {
		hourUnixTimestamp := timeutil.HourStart(time.Now().UTC().Add(time.Duration(hour) * time.Hour * -1)).Unix()
		result[hour] = &entity.CountersGroup{
			ActiveUserKey:     activeUserKey,
			HourUnixTimestamp: int32(hourUnixTimestamp),
		}
		if hour > maxHour {
			maxHour = hour
		}
	}

	timeFilter := timeutil.HourStart(time.Now().UTC().Add(time.Duration(maxHour) * time.Hour * -1))

	out, err := c.dynamoDbClient.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(CountersTableName),
		KeyConditionExpression: aws.String("u = :pk AND h >= :sk"),
		ConsistentRead:         aws.Bool(true),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: activeUserKey.ActiveUserId().String()},
			":sk": &types.AttributeValueMemberN{Value: strconv.FormatInt(timeFilter.Unix(), 10)},
		},
	}, func(o *dynamodb.Options) {
		o.Region = platformDynamoDb.GetDynamodbRegionByCountry(activeUserKey.CountryId())
	})

	if err != nil {
		return map[uint8]*entity.CountersGroup{}, err
	}

	c.logger.Debug(fmt.Sprintf("Got counters from dynamodb: %+v", out))

	if len(out.Items) == 0 {
		return result, nil
	}

	var countersGroup entity.CountersGroup
	for _, item := range out.Items {
		countersGroupItem := &CountersDocumentSchema{}
		if err = attributevalue.UnmarshalMap(item, &countersGroupItem); err != nil {
			return map[uint8]*entity.CountersGroup{}, err
		}

		countersGroup, err = c.transformCountersGroupItemToEntity(activeUserKey.CountryId(), *countersGroupItem)
		if err != nil {
			return map[uint8]*entity.CountersGroup{}, err
		}

		for _, group := range result {
			if countersGroup.HourUnixTimestamp < group.HourUnixTimestamp {
				continue
			}

			group.IncomingNo += countersGroup.IncomingNo
			group.IncomingYes += countersGroup.IncomingYes
			group.OutgoingNo += countersGroup.OutgoingNo
			group.OutgoingYes += countersGroup.OutgoingYes
		}
	}

	return result, nil
}

func (c *CountersRepository) IncrYesCounters(
	ctx context.Context,
	voteId sharedValueObject.VoteId,
	counterUpdateGroup countersValueObject.CounterUpdateGroup,
) {
	err := c.incrCounters(ctx, voteId, counterUpdateGroup, outgoingYesAttrName, incomingYesAttrName)
	if err != nil {
		c.logger.Error(fmt.Sprintf("incrYesCounters error: %s", err))
	}
}

func (c *CountersRepository) IncrNoCounters(
	ctx context.Context,
	voteId sharedValueObject.VoteId,
	counterUpdateGroup countersValueObject.CounterUpdateGroup,
) {
	err := c.incrCounters(ctx, voteId, counterUpdateGroup, outgoingNoAttrName, incomingNoAttrName)
	if err != nil {
		c.logger.Error(fmt.Sprintf("incrNoCounters error: %s", err))
	}
}

func (c *CountersRepository) incrCounters(
	ctx context.Context,
	voteId sharedValueObject.VoteId,
	counterUpdateGroup countersValueObject.CounterUpdateGroup,
	activeUserCounter string,
	peerUserCounter string,
) error {

	eventStartHourTime := counterUpdateGroup.HourStartTime().Unix()
	ttl := eventStartHourTime + c.config.Counters.TtlSeconds

	lifetimeValues := map[string]types.AttributeValue{
		":zero": &types.AttributeValueMemberN{Value: "0"},
		":incr": &types.AttributeValueMemberN{Value: "1"},
	}

	hourlyValues := map[string]types.AttributeValue{
		":zero": &types.AttributeValueMemberN{Value: "0"},
		":incr": &types.AttributeValueMemberN{Value: "1"},
		":ttl":  &types.AttributeValueMemberN{Value: strconv.FormatInt(ttl, 10)},
	}

	hourlyUpdateExpression := aws.String("SET #counterIndex = if_not_exists(#counterIndex, :zero) + :incr, #ttl = :ttl")
	lifetimeUpdateExpression := aws.String("SET #counterIndex = if_not_exists(#counterIndex, :zero) + :incr")

	input := &dynamodb.TransactWriteItemsInput{
		TransactItems: []types.TransactWriteItem{
			{
				Update: &types.Update{
					TableName:        aws.String(CountersTableName),
					Key:              c.getCountersTableKey(voteId.ActiveUserId(), eventStartHourTime),
					UpdateExpression: hourlyUpdateExpression,
					ExpressionAttributeNames: map[string]string{
						"#counterIndex": activeUserCounter,
						"#ttl":          platformDynamoDb.TtlAttrName,
					},
					ExpressionAttributeValues: hourlyValues,
				},
			},
			{
				Update: &types.Update{
					TableName:        aws.String(CountersTableName),
					Key:              c.getCountersTableKey(voteId.ActiveUserId(), LifetimeCounterKey),
					UpdateExpression: lifetimeUpdateExpression,
					ExpressionAttributeNames: map[string]string{
						"#counterIndex": activeUserCounter,
					},
					ExpressionAttributeValues: lifetimeValues,
				},
			},
			{
				Update: &types.Update{
					TableName:        aws.String(CountersTableName),
					Key:              c.getCountersTableKey(voteId.PeerUserId(), eventStartHourTime),
					UpdateExpression: hourlyUpdateExpression,
					ExpressionAttributeNames: map[string]string{
						"#counterIndex": peerUserCounter,
						"#ttl":          platformDynamoDb.TtlAttrName,
					},
					ExpressionAttributeValues: hourlyValues,
				},
			},
			{
				Update: &types.Update{
					TableName:        aws.String(CountersTableName),
					Key:              c.getCountersTableKey(voteId.PeerUserId(), LifetimeCounterKey),
					UpdateExpression: lifetimeUpdateExpression,
					ExpressionAttributeNames: map[string]string{
						"#counterIndex": peerUserCounter,
					},
					ExpressionAttributeValues: lifetimeValues,
				},
			},
		},
	}

	_, err := c.dynamoDbClient.TransactWriteItems(
		ctx,
		input,
		func(o *dynamodb.Options) {
			o.Region = platformDynamoDb.GetDynamodbRegionByCountry(voteId.CountryId())
		},
	)

	if err == nil {
		c.logger.Debug(fmt.Sprintf("Counters updated for users: %s and %s", voteId.ActiveUserId(), voteId.PeerUserId()))
	}

	return err
}

func (c *CountersRepository) transformCountersGroupItemToEntity(
	countryId uint16,
	countersItem CountersDocumentSchema,
) (entity.CountersGroup, error) {
	userId, err := uuid.Parse(countersItem.UserId)
	if err != nil {
		return entity.CountersGroup{}, err
	}

	activeUserKey, err := sharedValueObject.NewActiveUserKey(countryId, userId)
	if err != nil {
		return entity.CountersGroup{}, err
	}

	return entity.CountersGroup{
		ActiveUserKey:     activeUserKey,
		HourUnixTimestamp: countersItem.HourUnixTimestamp,
		IncomingYes:       countersItem.IncomingYes,
		IncomingNo:        countersItem.IncomingNo,
		OutgoingYes:       countersItem.OutgoingYes,
		OutgoingNo:        countersItem.OutgoingNo,
	}, nil
}

func (c *CountersRepository) getCountersTableKey(
	activeUserId uuid.UUID,
	dayStartTimeUnixTimestamp int64,
) map[string]types.AttributeValue {
	return map[string]types.AttributeValue{
		UserIdAttrName:            &types.AttributeValueMemberS{Value: activeUserId.String()},
		HourUnixTimestampAttrName: &types.AttributeValueMemberN{Value: strconv.FormatInt(dayStartTimeUnixTimestamp, 10)},
	}
}
