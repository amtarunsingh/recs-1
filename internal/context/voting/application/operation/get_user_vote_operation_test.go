package operation

import (
	"context"
	"errors"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/testlib/uuidhelper"
	"testing"

	romanceEntity "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/romance/entity"
	sharedValueObject "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/sharedkernel/valueobject"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/testlib/mocks"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

type GetUserVoteOperationUnitTestSuite struct {
	suite.Suite
	voteId       sharedValueObject.VoteId
	ctrl         *gomock.Controller
	romancesRepo *mocks.MockRomancesRepository
	ctx          context.Context
}

func TestGetUserVoteOperationUnitSuite(t *testing.T) {
	suite.Run(t, new(GetUserVoteOperationUnitTestSuite))
}

func (s *GetUserVoteOperationUnitTestSuite) SetupSuite() {
	activeUserId := uuidhelper.NewUUID(s.T())
	peerUserId := uuidhelper.NewUUID(s.T())
	countryId := uint16(11)

	voteId, err := sharedValueObject.NewVoteId(countryId, activeUserId, peerUserId)
	s.Require().NoError(err)
	s.voteId = voteId
	s.ctx = context.Background()
}

func (s *GetUserVoteOperationUnitTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.romancesRepo = mocks.NewMockRomancesRepository(s.ctrl)
}

func (s *GetUserVoteOperationUnitTestSuite) newOperation() *GetUserVoteOperation {
	return NewGetUserVoteOperation(s.romancesRepo)
}

func (s *GetUserVoteOperationUnitTestSuite) TestGetRomanceReturnsError() {
	expectedErr := errors.New("database error")

	s.romancesRepo.EXPECT().
		GetRomance(s.ctx, s.voteId).
		Return(romanceEntity.Romance{}, expectedErr)

	operation := s.newOperation()
	vote, err := operation.Run(s.ctx, s.voteId)

	s.Require().Error(err)
	s.Require().ErrorIs(err, expectedErr)
	s.Require().Equal(romanceEntity.Vote{}, vote)
}

func (s *GetUserVoteOperationUnitTestSuite) TestGetUserVoteSuccessfully() {
	romance := romanceEntity.CreateEmptyRomance(s.voteId)

	s.romancesRepo.EXPECT().
		GetRomance(s.ctx, s.voteId).
		Return(romance, nil)

	operation := s.newOperation()
	vote, err := operation.Run(s.ctx, s.voteId)

	s.Require().NoError(err)
	s.Require().Equal(romance.ActiveUserVote, vote)
}
