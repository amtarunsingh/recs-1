package persistence

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/bmbl-bumble2/recs-votes-storage/config"
	romanceDomain "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/romance"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/romance/entity"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/romance/valueobject"
	sharedValueObject "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/sharedkernel/valueobject"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/shared/platform"
	platformDynamoDb "github.com/bmbl-bumble2/recs-votes-storage/internal/shared/platform/dynamodb"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/shared/timeutil"
	"github.com/google/uuid"
)

const (
	RomancesTableName           = "Romances"
	PkUserIdAttrName            = "a"
	SkUserIdAttrName            = "b"
	pkUserVoteTypeAttrName      = "e"
	pkUserVotedAtAttrName       = "g"
	pkUserVoteCreatedAtAttrName = "h"
	pkUserVoteUpdatedAtAttrName = "i"
	skUserVoteTypeAttrName      = "l"
	skUserVotedAtAttrName       = "n"
	skUserVoteCreatedAtAttrName = "o"
	skUserVoteUpdatedAtAttrName = "p"
	versionAttrName             = "v"
)

type RomancesRepository struct {
	dynamoDbClient platformDynamoDb.Client
	config         config.Config
	logger         platform.Logger
}

type RomanceDocumentSchema struct {
	PkUserId            string `dynamodbav:"a"`
	SkUserId            string `dynamodbav:"b"`
	PkUserVoteType      uint8  `dynamodbav:"e"`
	PkUserVotedAt       *int32 `dynamodbav:"g"`
	PkUserVoteCreatedAt *int32 `dynamodbav:"h"`
	PkUserVoteUpdatedAt *int32 `dynamodbav:"i"`
	SkUserVoteType      uint8  `dynamodbav:"l"`
	SkUserVotedAt       *int32 `dynamodbav:"n"`
	SkUserVoteCreatedAt *int32 `dynamodbav:"o"`
	SkUserVoteUpdatedAt *int32 `dynamodbav:"p"`
	Version             uint32 `dynamodbav:"v"`
}

func NewRomancesRepository(
	dynamoDbClient platformDynamoDb.Client,
	config config.Config,
	logger platform.Logger,
) *RomancesRepository {
	return &RomancesRepository{
		dynamoDbClient: dynamoDbClient,
		config:         config,
		logger:         logger,
	}
}

func (r *RomancesRepository) getRomancesTableKey(key RomancePrimaryKey) map[string]types.AttributeValue {
	return map[string]types.AttributeValue{
		PkUserIdAttrName: &types.AttributeValueMemberS{Value: key.Pk.String()},
		SkUserIdAttrName: &types.AttributeValueMemberS{Value: key.Sk.String()},
	}
}

func (r *RomancesRepository) GetRomance(ctx context.Context, voteId sharedValueObject.VoteId) (entity.Romance, error) {

	romanceKey := NewRomancePrimaryKey(voteId)
	out, err := r.dynamoDbClient.GetItem(ctx, &dynamodb.GetItemInput{
		Key:            r.getRomancesTableKey(romanceKey),
		TableName:      aws.String(RomancesTableName),
		ConsistentRead: aws.Bool(true),
	}, func(o *dynamodb.Options) {
		o.Region = platformDynamoDb.GetDynamodbRegionByCountry(voteId.CountryId())
	})

	if err != nil {
		return entity.Romance{}, err
	}

	if out == nil || len(out.Item) == 0 {
		return entity.CreateEmptyRomance(voteId), nil
	}

	romanceItem := &RomanceDocumentSchema{}
	if err = attributevalue.UnmarshalMap(out.Item, romanceItem); err != nil {
		return entity.Romance{}, err
	}

	r.logger.Debug(fmt.Sprintf("Got romance from dynamodb: %+v", romanceItem))

	return r.transformRomanceItemToEntity(voteId.CountryId(), voteId.ActiveUserId(), *romanceItem)
}

func (r *RomancesRepository) GetAllPeersForActiveUser(
	ctx context.Context,
	userKey sharedValueObject.ActiveUserKey,
) (<-chan uuid.UUID, error) {
	out := make(chan uuid.UUID, 64)

	var (
		lastEvaluatedKey map[string]types.AttributeValue
		err              error
	)

	go func() {
		defer close(out)

		queryFn := func(indexName *string, pkName string) {
			for {

				input := &dynamodb.QueryInput{
					TableName:              aws.String(RomancesTableName),
					KeyConditionExpression: aws.String("#pk = :uid"),
					ExpressionAttributeNames: map[string]string{
						"#pk": pkName,
					},
					ExpressionAttributeValues: map[string]types.AttributeValue{
						":uid": &types.AttributeValueMemberS{Value: userKey.ActiveUserId().String()},
					},
					ExclusiveStartKey: lastEvaluatedKey,
				}

				if indexName != nil {
					input.IndexName = indexName
				}

				queryOutput, err := r.dynamoDbClient.Query(ctx, input)
				if err != nil {
					return
				}

				if len(queryOutput.Items) == 0 {
					return
				}

				var debugIndexName string
				if indexName == nil {
					debugIndexName = "PK"
				} else {
					debugIndexName = *indexName
				}

				r.logger.Debug(
					fmt.Sprintf(
						"Got peerIds from Romances table. Count: %d, used idx: %s, field: %s",
						len(queryOutput.Items),
						debugIndexName,
						pkName,
					),
				)

				for _, item := range queryOutput.Items {
					romanceItem := &RomanceDocumentSchema{}
					if err = attributevalue.UnmarshalMap(item, romanceItem); err != nil {
						return
					}

					romance, err := r.transformRomanceItemToEntity(
						userKey.CountryId(),
						userKey.ActiveUserId(),
						*romanceItem,
					)
					if err != nil {
						return
					}

					select {
					case out <- romance.ActiveUserVote.Id.PeerUserId():
					case <-ctx.Done():
						return
					}
				}

				if queryOutput.LastEvaluatedKey == nil {
					break
				}
				lastEvaluatedKey = queryOutput.LastEvaluatedKey
			}
		}

		queryFn(nil, PkUserIdAttrName)

		indexName := aws.String("gsiByMaxMinUser")
		queryFn(indexName, SkUserIdAttrName)
	}()

	return out, err
}

func (r *RomancesRepository) AddActiveUserVoteToRomance(
	ctx context.Context,
	romance entity.Romance,
	voteType valueobject.VoteType,
	votedAt time.Time,
) (entity.Romance, error) {

	activeUserId := romance.ActiveUserVote.Id.ActiveUserId()
	countryId := romance.ActiveUserVote.Id.CountryId()

	romanceKey := NewRomancePrimaryKey(romance.ActiveUserVote.Id)
	now := time.Now()

	exprNames := map[string]string{
		"#version": versionAttrName,
		"#ttl":     platformDynamoDb.TtlAttrName,
	}

	if romanceKey.isPartitionKey(activeUserId) {
		exprNames["#voteType"] = pkUserVoteTypeAttrName
		exprNames["#votedAt"] = pkUserVotedAtAttrName
		exprNames["#voteCreatedAt"] = pkUserVoteCreatedAtAttrName
	} else {
		exprNames["#voteType"] = skUserVoteTypeAttrName
		exprNames["#votedAt"] = skUserVotedAtAttrName
		exprNames["#voteCreatedAt"] = skUserVoteCreatedAtAttrName
	}

	currentVersion := int64(romance.Version)
	ttlSeconds := r.getTtlSecondsForVotesPair(voteType, romance.PeerUserVote.VoteType)

	exprValues := map[string]types.AttributeValue{
		":voteType":  &types.AttributeValueMemberN{Value: strconv.Itoa(int(voteType))},
		":votedAt":   &types.AttributeValueMemberN{Value: strconv.FormatInt(votedAt.Unix(), 10)},
		":createdAt": &types.AttributeValueMemberN{Value: strconv.FormatInt(now.Unix(), 10)},
		":v":         &types.AttributeValueMemberN{Value: strconv.FormatInt(currentVersion+1, 10)},
		":ttl":       &types.AttributeValueMemberN{Value: strconv.FormatInt(ttlSeconds, 10)},
	}

	var conditionExpression string
	if romance.Version == 0 {
		conditionExpression = "attribute_not_exists(a) AND attribute_not_exists(b)"
	} else {
		conditionExpression = "#version = :expectedV"
		exprValues[":expectedV"] = &types.AttributeValueMemberN{Value: strconv.FormatInt(currentVersion, 10)}
	}

	updateExpr := aws.String("SET #voteType = :voteType, #votedAt = :votedAt, #voteCreatedAt = :createdAt, #version = :v, #ttl = :ttl")

	out, err := r.dynamoDbClient.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		Key:                       r.getRomancesTableKey(romanceKey),
		TableName:                 aws.String(RomancesTableName),
		UpdateExpression:          updateExpr,
		ExpressionAttributeNames:  exprNames,
		ExpressionAttributeValues: exprValues,
		ConditionExpression:       aws.String(conditionExpression),
		ReturnValues:              types.ReturnValueAllNew,
	}, func(o *dynamodb.Options) {
		o.Region = platformDynamoDb.GetDynamodbRegionByCountry(countryId)
	})

	if err != nil {
		var condCheckErr *types.ConditionalCheckFailedException
		if errors.As(err, &condCheckErr) {
			return entity.Romance{}, romanceDomain.ErrVersionConflict
		}

		return entity.Romance{}, err
	}

	romanceItem := &RomanceDocumentSchema{}
	if err = attributevalue.UnmarshalMap(out.Attributes, romanceItem); err != nil {
		return entity.Romance{}, err
	}

	r.logger.Debug(fmt.Sprintf("Updated romance in dynamodb: %+v", romanceItem))

	return r.transformRomanceItemToEntity(countryId, activeUserId, *romanceItem)
}

func (r *RomancesRepository) DeleteRomance(
	ctx context.Context,
	voteId sharedValueObject.VoteId,
) error {
	romanceKey := NewRomancePrimaryKey(voteId)
	out, err := r.dynamoDbClient.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		Key:       r.getRomancesTableKey(romanceKey),
		TableName: aws.String(RomancesTableName),
	}, func(o *dynamodb.Options) {
		o.Region = platformDynamoDb.GetDynamodbRegionByCountry(voteId.CountryId())
	})

	if err != nil {
		return err
	}

	r.logger.Debug(fmt.Sprintf("Romance deleted from dynamodb: %+v", out))
	return nil
}

func (r *RomancesRepository) DeleteRomancesGroup(
	ctx context.Context,
	userKey sharedValueObject.ActiveUserKey,
	peerIds []uuid.UUID,
) error {
	var batch []types.WriteRequest
	var keysLog []uuid.UUID
	for _, peerId := range peerIds {
		voteId, err := sharedValueObject.NewVoteId(userKey.CountryId(), userKey.ActiveUserId(), peerId)
		if err != nil {
			return err
		}
		batch = append(batch, types.WriteRequest{
			DeleteRequest: &types.DeleteRequest{
				Key: r.getRomancesTableKey(NewRomancePrimaryKey(voteId)),
			},
		})

		keysLog = append(keysLog, peerId)

		if len(batch) == 25 {
			_, err := r.dynamoDbClient.BatchWriteItem(ctx, &dynamodb.BatchWriteItemInput{
				RequestItems: map[string][]types.WriteRequest{
					RomancesTableName: batch,
				},
			})
			if err != nil {
				return err
			}
			batch = batch[:0]
			r.logger.Debug(
				fmt.Sprintf(
					"Deleted records from Romances talbe. Count: %d, ActiveIserId: %s, CountryId: %d, PeerIds: %+v",
					len(keysLog),
					userKey.ActiveUserId(),
					userKey.CountryId(),
					keysLog,
				),
			)
			keysLog = nil
		}
	}

	if len(batch) > 0 {
		_, err := r.dynamoDbClient.BatchWriteItem(ctx, &dynamodb.BatchWriteItemInput{
			RequestItems: map[string][]types.WriteRequest{
				RomancesTableName: batch,
			},
		})
		if err != nil {
			return err
		}
		r.logger.Debug(
			fmt.Sprintf(
				"Deleted records from Romances talbe. Count: %d, ActiveIserId: %s, CountryId: %d, PeerIds: %+v",
				len(keysLog),
				userKey.ActiveUserId(),
				userKey.CountryId(),
				keysLog,
			),
		)
	}

	return nil
}

func (r *RomancesRepository) DeleteActiveUserVoteFromRomance(ctx context.Context, romance entity.Romance) error {
	if romance.IsEmpty() {
		return nil
	}

	activeUserId := romance.ActiveUserVote.Id.ActiveUserId()
	countryId := romance.ActiveUserVote.Id.CountryId()

	romanceKey := NewRomancePrimaryKey(romance.ActiveUserVote.Id)
	exprNames := map[string]string{
		"#version": versionAttrName,
		"#ttl":     platformDynamoDb.TtlAttrName,
	}

	if romanceKey.isPartitionKey(activeUserId) {
		exprNames["#voteType"] = pkUserVoteTypeAttrName
		exprNames["#votedAt"] = pkUserVotedAtAttrName
		exprNames["#voteCreatedAt"] = pkUserVoteCreatedAtAttrName
		exprNames["#voteUpdatedAt"] = pkUserVoteUpdatedAtAttrName
	} else {
		exprNames["#voteType"] = skUserVoteTypeAttrName
		exprNames["#votedAt"] = skUserVotedAtAttrName
		exprNames["#voteCreatedAt"] = skUserVoteCreatedAtAttrName
		exprNames["#voteUpdatedAt"] = skUserVoteUpdatedAtAttrName
	}

	currentVersion := int64(romance.Version)
	ttlSeconds := r.getTtlSecondsForVotesPair(valueobject.VoteTypeEmpty, romance.PeerUserVote.VoteType)

	exprValues := map[string]types.AttributeValue{
		":v":   &types.AttributeValueMemberN{Value: strconv.FormatInt(currentVersion+1, 10)},
		":ttl": &types.AttributeValueMemberN{Value: strconv.FormatInt(ttlSeconds, 10)},
	}

	conditionExpression := "#version = :expectedV"
	exprValues[":expectedV"] = &types.AttributeValueMemberN{Value: strconv.FormatInt(currentVersion, 10)}

	updateExpr := aws.String("SET #version = :v, #ttl = :ttl REMOVE #voteType, #votedAt, #voteCreatedAt, #voteUpdatedAt")

	out, err := r.dynamoDbClient.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		Key:                       r.getRomancesTableKey(romanceKey),
		TableName:                 aws.String(RomancesTableName),
		UpdateExpression:          updateExpr,
		ExpressionAttributeNames:  exprNames,
		ExpressionAttributeValues: exprValues,
		ConditionExpression:       aws.String(conditionExpression),
		ReturnValues:              types.ReturnValueAllNew,
	}, func(o *dynamodb.Options) {
		o.Region = platformDynamoDb.GetDynamodbRegionByCountry(countryId)
	})

	if err != nil {
		var condCheckErr *types.ConditionalCheckFailedException
		if errors.As(err, &condCheckErr) {
			return romanceDomain.ErrVersionConflict
		}

		return err
	}

	r.logger.Debug(fmt.Sprintf("Deleted romance vote from dynamodb: %+v", out))
	return nil
}

func (r *RomancesRepository) ChangeActiveUserVoteTypeInRomance(
	ctx context.Context,
	romance entity.Romance,
	newVoteType valueobject.VoteType,
) (entity.Romance, error) {
	if romance.ActiveUserVote.VoteType.IsEmpty() {
		return entity.Romance{}, romanceDomain.ErrVoteNotFound
	}

	if newVoteType.IsEmpty() {
		return entity.Romance{}, romanceDomain.ErrWrongVote
	}

	activeUserId := romance.ActiveUserVote.Id.ActiveUserId()
	countryId := romance.ActiveUserVote.Id.CountryId()

	romanceKey := NewRomancePrimaryKey(romance.ActiveUserVote.Id)
	now := time.Now()

	exprNames := map[string]string{
		"#version": versionAttrName,
		"#ttl":     platformDynamoDb.TtlAttrName,
	}

	if romanceKey.isPartitionKey(activeUserId) {
		exprNames["#voteType"] = pkUserVoteTypeAttrName
		exprNames["#voteUpdatedAt"] = pkUserVoteUpdatedAtAttrName
	} else {
		exprNames["#voteType"] = skUserVoteTypeAttrName
		exprNames["#voteUpdatedAt"] = skUserVoteUpdatedAtAttrName
	}

	currentVersion := int64(romance.Version)
	ttlSeconds := r.getTtlSecondsForVotesPair(newVoteType, romance.PeerUserVote.VoteType)

	exprValues := map[string]types.AttributeValue{
		":voteType":  &types.AttributeValueMemberN{Value: strconv.Itoa(int(newVoteType))},
		":updatedAt": &types.AttributeValueMemberN{Value: strconv.FormatInt(now.Unix(), 10)},
		":v":         &types.AttributeValueMemberN{Value: strconv.FormatInt(currentVersion+1, 10)},
		":ttl":       &types.AttributeValueMemberN{Value: strconv.FormatInt(ttlSeconds, 10)},
	}

	conditionExpression := "#version = :expectedV"
	exprValues[":expectedV"] = &types.AttributeValueMemberN{Value: strconv.FormatInt(currentVersion, 10)}

	updateExpr := aws.String("SET #voteType = :voteType, #voteUpdatedAt = :updatedAt, #version = :v, #ttl = :ttl")

	out, err := r.dynamoDbClient.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		Key:                       r.getRomancesTableKey(romanceKey),
		TableName:                 aws.String(RomancesTableName),
		UpdateExpression:          updateExpr,
		ExpressionAttributeNames:  exprNames,
		ExpressionAttributeValues: exprValues,
		ConditionExpression:       aws.String(conditionExpression),
		ReturnValues:              types.ReturnValueAllNew,
	}, func(o *dynamodb.Options) {
		o.Region = platformDynamoDb.GetDynamodbRegionByCountry(countryId)
	})

	if err != nil {
		var condCheckErr *types.ConditionalCheckFailedException
		if errors.As(err, &condCheckErr) {
			return entity.Romance{}, romanceDomain.ErrVersionConflict
		}

		return entity.Romance{}, err
	}

	romanceItem := &RomanceDocumentSchema{}
	if err = attributevalue.UnmarshalMap(out.Attributes, romanceItem); err != nil {
		return entity.Romance{}, err
	}

	r.logger.Debug(fmt.Sprintf("Updated romance in dynamodb: %+v", romanceItem))

	return r.transformRomanceItemToEntity(countryId, activeUserId, *romanceItem)
}

func (r *RomancesRepository) transformRomanceItemToEntity(
	countryId uint16,
	activeUserId uuid.UUID,
	romanceItem RomanceDocumentSchema,
) (entity.Romance, error) {

	pkUserId, err := uuid.Parse(romanceItem.PkUserId)
	if err != nil {
		return entity.Romance{}, err
	}

	skUserId, err := uuid.Parse(romanceItem.SkUserId)
	if err != nil {
		return entity.Romance{}, err
	}

	pkUserVote := entity.Vote{
		VoteType:  valueobject.VoteType(romanceItem.PkUserVoteType),
		VotedAt:   timeutil.UnixToTimePtr(romanceItem.PkUserVotedAt),
		CreatedAt: timeutil.UnixToTimePtr(romanceItem.PkUserVoteCreatedAt),
		UpdatedAt: timeutil.UnixToTimePtr(romanceItem.PkUserVoteUpdatedAt),
	}

	skUserVote := entity.Vote{
		VoteType:  valueobject.VoteType(romanceItem.SkUserVoteType),
		VotedAt:   timeutil.UnixToTimePtr(romanceItem.SkUserVotedAt),
		CreatedAt: timeutil.UnixToTimePtr(romanceItem.SkUserVoteCreatedAt),
		UpdatedAt: timeutil.UnixToTimePtr(romanceItem.SkUserVoteUpdatedAt),
	}

	var peerUserId uuid.UUID
	if activeUserId == pkUserId {
		peerUserId = skUserId
	} else {
		peerUserId = pkUserId
	}

	activeUserVoteId, err := sharedValueObject.NewVoteId(countryId, activeUserId, peerUserId)
	if err != nil {
		return entity.Romance{}, err
	}
	peerUserVoteId := activeUserVoteId.ToPeerVoteId()

	var resultActiveUserVote entity.Vote
	var resultPeerUserVote entity.Vote

	if activeUserId == pkUserId {
		resultActiveUserVote = pkUserVote
		resultActiveUserVote.Id = activeUserVoteId
		resultPeerUserVote = skUserVote
		resultPeerUserVote.Id = peerUserVoteId
	} else {
		resultActiveUserVote = skUserVote
		resultActiveUserVote.Id = activeUserVoteId
		resultPeerUserVote = pkUserVote
		resultPeerUserVote.Id = peerUserVoteId
	}

	return entity.Romance{
		ActiveUserVote: resultActiveUserVote,
		PeerUserVote:   resultPeerUserVote,
		Version:        romanceItem.Version,
	}, nil
}

func (r *RomancesRepository) getTtlSecondsForVotesPair(
	activeUserVoteType valueobject.VoteType,
	peerUserVoteType valueobject.VoteType,
) int64 {

	if activeUserVoteType.IsNegative() || peerUserVoteType.IsNegative() {
		return r.config.Romances.DeadRomanceTtlSeconds
	}

	if activeUserVoteType.IsPositive() && peerUserVoteType.IsPositive() {
		return r.config.Romances.MutualRomanceTtlSeconds
	}

	return r.config.Romances.NonMutualRomanceTtlSeconds
}

type RomancePrimaryKey struct {
	Pk uuid.UUID
	Sk uuid.UUID
}

func NewRomancePrimaryKey(voteId sharedValueObject.VoteId) RomancePrimaryKey {
	activeUserId := voteId.ActiveUserId()
	peerUserId := voteId.PeerUserId()

	if bytes.Compare(activeUserId[:], peerUserId[:]) == -1 {
		return RomancePrimaryKey{
			Pk: activeUserId,
			Sk: peerUserId,
		}
	} else {
		return RomancePrimaryKey{
			Pk: peerUserId,
			Sk: activeUserId,
		}
	}
}

func (r *RomancePrimaryKey) isPartitionKey(someUuid uuid.UUID) bool {
	return someUuid == r.Pk
}
