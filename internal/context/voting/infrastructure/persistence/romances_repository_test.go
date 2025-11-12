package persistence

import (
	"context"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/testlib/uuidhelper"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/bmbl-bumble2/recs-votes-storage/config"
	romanceEntity "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/romance/entity"
	rvo "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/romance/valueobject"
	sharedValueObject "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/sharedkernel/valueobject"
	platformDynamodb "github.com/bmbl-bumble2/recs-votes-storage/internal/shared/platform/dynamodb"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/testlib/mocks"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

type RomancesRepositoryUnitTestSuite struct {
	suite.Suite
	voteId sharedValueObject.VoteId
}

func TestRomancesRepositoryUnitSuite(t *testing.T) {
	suite.Run(t, new(RomancesRepositoryUnitTestSuite))
}

func (s *RomancesRepositoryUnitTestSuite) SetupSuite() {
	activeUserId := uuidhelper.NewUUID(s.T())
	peerUserId := uuidhelper.NewUUID(s.T())
	countryId := uint16(11)

	voteId, err := sharedValueObject.NewVoteId(countryId, activeUserId, peerUserId)
	s.Require().NoError(err)
	s.voteId = voteId
}

func (s *RomancesRepositoryUnitTestSuite) TestGetRomanceWithDbException() {
	ctrl := gomock.NewController(s.T())
	mock := mocks.NewMockClient(ctrl)

	ctx := context.Background()

	mock.EXPECT().
		GetItem(ctx, gomock.Any(), gomock.Any()).
		Return(nil, &types.InvalidEndpointException{})

	repo := newRomancesRepository(mock)

	romance, err := repo.GetRomance(ctx, s.voteId)
	s.Require().Error(err)
	s.assertEmptyRomance(romance)
}

func (s *RomancesRepositoryUnitTestSuite) TestAddVoteWithDbException() {
	ctrl := gomock.NewController(s.T())
	mock := mocks.NewMockClient(ctrl)

	ctx := context.Background()
	romance := romanceEntity.CreateEmptyRomance(s.voteId)

	expectedErr := &types.InvalidEndpointException{}

	mock.EXPECT().
		UpdateItem(ctx, gomock.Any(), gomock.Any()).
		Return(nil, expectedErr)

	repo := newRomancesRepository(mock)

	_, err := repo.AddActiveUserVoteToRomance(ctx, romance, rvo.VoteTypeYes, time.Now())
	s.Require().Error(err)
	s.Require().ErrorIs(err, expectedErr)
}

func (s *RomancesRepositoryUnitTestSuite) TestDeleteRomanceWithDbException() {
	ctrl := gomock.NewController(s.T())
	mock := mocks.NewMockClient(ctrl)

	ctx := context.Background()

	expectedErr := &types.InvalidEndpointException{}

	mock.EXPECT().
		DeleteItem(ctx, gomock.Any(), gomock.Any()).
		Return(nil, expectedErr)

	repo := newRomancesRepository(mock)

	err := repo.DeleteRomance(ctx, s.voteId)
	s.Require().Error(err)
	s.Require().ErrorIs(err, expectedErr)
}

func (s *RomancesRepositoryUnitTestSuite) TestDeleteActiveUserVoteWithDbException() {
	ctrl := gomock.NewController(s.T())
	mock := mocks.NewMockClient(ctrl)

	ctx := context.Background()

	// Manually construct a romance with a vote (simulating existing state)
	romance := romanceEntity.CreateEmptyRomance(s.voteId)
	now := time.Now()
	romance.ActiveUserVote.VoteType = rvo.VoteTypeYes
	romance.ActiveUserVote.VotedAt = &now
	romance.ActiveUserVote.CreatedAt = &now
	romance.ActiveUserVote.UpdatedAt = &now
	romance.Version = 1

	expectedErr := &types.InvalidEndpointException{}

	mock.EXPECT().
		UpdateItem(ctx, gomock.Any(), gomock.Any()).
		Return(nil, expectedErr)

	repo := newRomancesRepository(mock)

	err := repo.DeleteActiveUserVoteFromRomance(ctx, romance)
	s.Require().Error(err)
	s.Require().ErrorIs(err, expectedErr)
}

func (s *RomancesRepositoryUnitTestSuite) TestChangeActiveUserVoteTypeInRomanceWithDbError() {
	ctrl := gomock.NewController(s.T())
	mock := mocks.NewMockClient(ctrl)

	ctx := context.Background()

	// Manually construct a romance with a vote (simulating existing state)
	romance := romanceEntity.CreateEmptyRomance(s.voteId)
	now := time.Now()
	romance.ActiveUserVote.VoteType = rvo.VoteTypeNo
	romance.ActiveUserVote.VotedAt = &now
	romance.ActiveUserVote.CreatedAt = &now
	romance.ActiveUserVote.UpdatedAt = &now
	romance.Version = 1

	expectedErr := &types.InvalidEndpointException{}

	mock.EXPECT().
		UpdateItem(ctx, gomock.Any(), gomock.Any()).
		Return(nil, expectedErr)

	repo := newRomancesRepository(mock)

	newRomance, err := repo.ChangeActiveUserVoteTypeInRomance(ctx, romance, rvo.VoteTypeYes)
	s.Require().Error(err)
	s.Require().ErrorIs(err, expectedErr)
	s.assertEmptyRomance(newRomance)
}

// Helper methods
func (s *RomancesRepositoryUnitTestSuite) assertEmptyRomance(romanceToCheck romanceEntity.Romance) {
	s.Require().Equal(romanceEntity.Romance{}, romanceToCheck)
}

func newRomancesRepository(client platformDynamodb.Client) *RomancesRepository {
	appConfig := config.Load()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return NewRomancesRepository(client, appConfig, logger)
}
