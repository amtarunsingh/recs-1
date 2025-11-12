package operation

import (
	"context"
	"errors"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/testlib/uuidhelper"
	"io"
	"log/slog"
	"testing"
	"time"

	romanceDomain "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/romance"
	romanceEntity "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/romance/entity"
	romancesValueObject "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/romance/valueobject"
	sharedValueObject "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/sharedkernel/valueobject"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/testlib/mocks"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

type AddUserVoteOperationUnitTestSuite struct {
	suite.Suite
	voteId       sharedValueObject.VoteId
	ctrl         *gomock.Controller
	romancesRepo *mocks.MockRomancesRepository
	countersRepo *mocks.MockCountersRepository
	logger       *slog.Logger
	ctx          context.Context
}

func TestAddUserVoteOperationUnitSuite(t *testing.T) {
	suite.Run(t, new(AddUserVoteOperationUnitTestSuite))
}

func (s *AddUserVoteOperationUnitTestSuite) SetupSuite() {
	activeUserId := uuidhelper.NewUUID(s.T())
	peerUserId := uuidhelper.NewUUID(s.T())
	countryId := uint16(11)

	voteId, err := sharedValueObject.NewVoteId(countryId, activeUserId, peerUserId)
	s.Require().NoError(err)
	s.voteId = voteId
	s.ctx = context.Background()
	s.logger = slog.New(slog.NewTextHandler(io.Discard, nil))
}

func (s *AddUserVoteOperationUnitTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.romancesRepo = mocks.NewMockRomancesRepository(s.ctrl)
	s.countersRepo = mocks.NewMockCountersRepository(s.ctrl)
}

func (s *AddUserVoteOperationUnitTestSuite) newOperation() *AddUserVoteOperation {
	return NewAddUserVoteOperation(s.romancesRepo, s.countersRepo, s.logger)
}

func (s *AddUserVoteOperationUnitTestSuite) TestGetRomanceReturnsError() {
	expectedErr := errors.New("database error")

	s.romancesRepo.EXPECT().
		GetRomance(s.ctx, s.voteId).
		Return(romanceEntity.Romance{}, expectedErr)

	operation := s.newOperation()
	vote, err := operation.Run(s.ctx, s.voteId, romancesValueObject.VoteTypeYes, time.Now())

	s.Require().Error(err)
	s.Require().ErrorIs(err, expectedErr)
	s.Require().Equal(romanceEntity.Vote{}, vote)
}

func (s *AddUserVoteOperationUnitTestSuite) TestAddActiveUserVoteToRomanceReturnsError() {
	romance := romanceEntity.CreateEmptyRomance(s.voteId)
	expectedErr := errors.New("database error")
	votedAt := time.Now()

	s.romancesRepo.EXPECT().
		GetRomance(s.ctx, s.voteId).
		Return(romance, nil)

	s.romancesRepo.EXPECT().
		AddActiveUserVoteToRomance(s.ctx, romance, romancesValueObject.VoteTypeYes, votedAt).
		Return(romanceEntity.Romance{}, expectedErr)

	operation := s.newOperation()
	vote, err := operation.Run(s.ctx, s.voteId, romancesValueObject.VoteTypeYes, votedAt)

	s.Require().Error(err)
	s.Require().ErrorIs(err, expectedErr)
	s.Require().Equal(romanceEntity.Vote{}, vote)
}

func (s *AddUserVoteOperationUnitTestSuite) TestVersionConflictRetriesAndSucceeds() {
	romance := romanceEntity.CreateEmptyRomance(s.voteId)
	votedAt := time.Now()

	// First call to GetRomance
	s.romancesRepo.EXPECT().
		GetRomance(s.ctx, s.voteId).
		Return(romance, nil)

	// First AddActiveUserVoteToRomance fails with version conflict
	s.romancesRepo.EXPECT().
		AddActiveUserVoteToRomance(s.ctx, gomock.Any(), romancesValueObject.VoteTypeYes, gomock.Any()).
		Return(romanceEntity.Romance{}, romanceDomain.ErrVersionConflict)

	// Second call to GetRomance (retry)
	s.romancesRepo.EXPECT().
		GetRomance(s.ctx, s.voteId).
		Return(romance, nil)

	// Second AddActiveUserVoteToRomance succeeds
	updatedRomance := romance
	updatedRomance.ActiveUserVote.VoteType = romancesValueObject.VoteTypeYes
	updatedRomance.ActiveUserVote.VotedAt = &votedAt
	s.romancesRepo.EXPECT().
		AddActiveUserVoteToRomance(s.ctx, gomock.Any(), romancesValueObject.VoteTypeYes, gomock.Any()).
		Return(updatedRomance, nil)

	// Counter should be incremented (no return value)
	s.countersRepo.EXPECT().
		IncrYesCounters(s.ctx, s.voteId, gomock.Any())

	operation := s.newOperation()
	vote, err := operation.Run(s.ctx, s.voteId, romancesValueObject.VoteTypeYes, votedAt)

	s.Require().NoError(err)
	s.Require().Equal(romancesValueObject.VoteTypeYes, vote.VoteType)
	s.Require().Equal(&votedAt, vote.VotedAt)
}

func (s *AddUserVoteOperationUnitTestSuite) TestAddVoteSuccessfully() {
	testCases := []struct {
		name             string
		voteType         romancesValueObject.VoteType
		expectYesCounter bool
		expectNoCounter  bool
	}{
		{
			name:             "Add Yes vote",
			voteType:         romancesValueObject.VoteTypeYes,
			expectYesCounter: true,
			expectNoCounter:  false,
		},
		{
			name:             "Add No vote",
			voteType:         romancesValueObject.VoteTypeNo,
			expectYesCounter: false,
			expectNoCounter:  true,
		},
		{
			name:             "Add Crush vote",
			voteType:         romancesValueObject.VoteTypeCrush,
			expectYesCounter: true,
			expectNoCounter:  false,
		},
		{
			name:             "Add Compliment vote",
			voteType:         romancesValueObject.VoteTypeCompliment,
			expectYesCounter: true,
			expectNoCounter:  false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			romance := romanceEntity.CreateEmptyRomance(s.voteId)
			votedAt := time.Now()

			s.romancesRepo.EXPECT().
				GetRomance(s.ctx, s.voteId).
				Return(romance, nil)

			updatedRomance := romance
			updatedRomance.ActiveUserVote.VoteType = tc.voteType
			updatedRomance.ActiveUserVote.VotedAt = &votedAt
			s.romancesRepo.EXPECT().
				AddActiveUserVoteToRomance(s.ctx, romance, tc.voteType, votedAt).
				Return(updatedRomance, nil)

			if tc.expectYesCounter {
				s.countersRepo.EXPECT().
					IncrYesCounters(s.ctx, s.voteId, gomock.Any())
			}
			if tc.expectNoCounter {
				s.countersRepo.EXPECT().
					IncrNoCounters(s.ctx, s.voteId, gomock.Any())
			}

			operation := s.newOperation()
			vote, err := operation.Run(s.ctx, s.voteId, tc.voteType, votedAt)

			s.Require().NoError(err)
			s.Require().Equal(tc.voteType, vote.VoteType)
			s.Require().Equal(&votedAt, vote.VotedAt)
		})
	}
}
