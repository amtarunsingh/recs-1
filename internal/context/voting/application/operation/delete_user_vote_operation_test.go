package operation

import (
	"context"
	"errors"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/testlib/uuidhelper"
	"io"
	"log/slog"
	"testing"

	romanceDomain "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/romance"
	romanceEntity "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/romance/entity"
	sharedValueObject "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/sharedkernel/valueobject"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/testlib/mocks"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

type DeleteUserVoteOperationUnitTestSuite struct {
	suite.Suite
	voteId       sharedValueObject.VoteId
	ctrl         *gomock.Controller
	romancesRepo *mocks.MockRomancesRepository
	countersRepo *mocks.MockCountersRepository
	logger       *slog.Logger
	ctx          context.Context
}

func TestDeleteUserVoteOperationUnitSuite(t *testing.T) {
	suite.Run(t, new(DeleteUserVoteOperationUnitTestSuite))
}

func (s *DeleteUserVoteOperationUnitTestSuite) SetupSuite() {
	activeUserId := uuidhelper.NewUUID(s.T())
	peerUserId := uuidhelper.NewUUID(s.T())
	countryId := uint16(11)

	voteId, err := sharedValueObject.NewVoteId(countryId, activeUserId, peerUserId)
	s.Require().NoError(err)
	s.voteId = voteId
	s.ctx = context.Background()
	s.logger = slog.New(slog.NewTextHandler(io.Discard, nil))
}

func (s *DeleteUserVoteOperationUnitTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.romancesRepo = mocks.NewMockRomancesRepository(s.ctrl)
	s.countersRepo = mocks.NewMockCountersRepository(s.ctrl)
}

func (s *DeleteUserVoteOperationUnitTestSuite) newOperation() *DeleteUserVoteOperation {
	return NewDeleteUserVoteOperation(s.romancesRepo, s.countersRepo, s.logger)
}

func (s *DeleteUserVoteOperationUnitTestSuite) TestGetRomanceReturnsError() {
	expectedErr := errors.New("database error")

	s.romancesRepo.EXPECT().
		GetRomance(s.ctx, s.voteId).
		Return(romanceEntity.Romance{}, expectedErr)

	operation := s.newOperation()
	err := operation.Run(s.ctx, s.voteId)

	s.Require().Error(err)
	s.Require().ErrorIs(err, expectedErr)
}

func (s *DeleteUserVoteOperationUnitTestSuite) TestDeleteActiveUserVoteFromRomanceReturnsError() {
	romance := romanceEntity.CreateEmptyRomance(s.voteId)
	expectedErr := errors.New("database error")

	s.romancesRepo.EXPECT().
		GetRomance(s.ctx, s.voteId).
		Return(romance, nil)

	s.romancesRepo.EXPECT().
		DeleteActiveUserVoteFromRomance(s.ctx, romance).
		Return(expectedErr)

	operation := s.newOperation()
	err := operation.Run(s.ctx, s.voteId)

	s.Require().Error(err)
	s.Require().ErrorIs(err, expectedErr)
}

func (s *DeleteUserVoteOperationUnitTestSuite) TestVersionConflictRetriesAndSucceeds() {
	romance := romanceEntity.CreateEmptyRomance(s.voteId)

	// First call to GetRomance
	s.romancesRepo.EXPECT().
		GetRomance(s.ctx, s.voteId).
		Return(romance, nil)

	// First DeleteActiveUserVoteFromRomance fails with version conflict
	s.romancesRepo.EXPECT().
		DeleteActiveUserVoteFromRomance(s.ctx, romance).
		Return(romanceDomain.ErrVersionConflict)

	// Second call to GetRomance (retry)
	s.romancesRepo.EXPECT().
		GetRomance(s.ctx, s.voteId).
		Return(romance, nil)

	// Second DeleteActiveUserVoteFromRomance succeeds
	s.romancesRepo.EXPECT().
		DeleteActiveUserVoteFromRomance(s.ctx, romance).
		Return(nil)

	operation := s.newOperation()
	err := operation.Run(s.ctx, s.voteId)

	s.Require().NoError(err)
}

func (s *DeleteUserVoteOperationUnitTestSuite) TestDeleteVoteSuccessfully() {
	romance := romanceEntity.CreateEmptyRomance(s.voteId)

	s.romancesRepo.EXPECT().
		GetRomance(s.ctx, s.voteId).
		Return(romance, nil)

	s.romancesRepo.EXPECT().
		DeleteActiveUserVoteFromRomance(s.ctx, romance).
		Return(nil)

	operation := s.newOperation()
	err := operation.Run(s.ctx, s.voteId)

	s.Require().NoError(err)
}
