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

type ChangeUserVoteOperationUnitTestSuite struct {
	suite.Suite
	voteId       sharedValueObject.VoteId
	ctrl         *gomock.Controller
	romancesRepo *mocks.MockRomancesRepository
	countersRepo *mocks.MockCountersRepository
	logger       *slog.Logger
	ctx          context.Context
}

func TestChangeUserVoteOperationUnitSuite(t *testing.T) {
	suite.Run(t, new(ChangeUserVoteOperationUnitTestSuite))
}

func (s *ChangeUserVoteOperationUnitTestSuite) SetupSuite() {
	activeUserId := uuidhelper.NewUUID(s.T())
	peerUserId := uuidhelper.NewUUID(s.T())
	countryId := uint16(11)

	voteId, err := sharedValueObject.NewVoteId(countryId, activeUserId, peerUserId)
	s.Require().NoError(err)
	s.voteId = voteId
	s.ctx = context.Background()
	s.logger = slog.New(slog.NewTextHandler(io.Discard, nil))
}

func (s *ChangeUserVoteOperationUnitTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.romancesRepo = mocks.NewMockRomancesRepository(s.ctrl)
	s.countersRepo = mocks.NewMockCountersRepository(s.ctrl)
}

func (s *ChangeUserVoteOperationUnitTestSuite) newOperation() *ChangeUserVoteOperation {
	return NewChangeUserVoteOperation(s.romancesRepo, s.countersRepo, s.logger)
}

func (s *ChangeUserVoteOperationUnitTestSuite) TestGetRomanceReturnsError() {
	expectedErr := errors.New("database error")

	s.romancesRepo.EXPECT().
		GetRomance(s.ctx, s.voteId).
		Return(romanceEntity.Romance{}, expectedErr)

	operation := s.newOperation()
	vote, err := operation.Run(s.ctx, s.voteId, romancesValueObject.VoteTypeCrush)

	s.Require().Error(err)
	s.Require().ErrorIs(err, expectedErr)
	s.Require().Equal(romanceEntity.Vote{}, vote)
}

func (s *ChangeUserVoteOperationUnitTestSuite) TestChangeActiveUserVoteTypeInRomanceReturnsError() {
	romance := romanceEntity.CreateEmptyRomance(s.voteId)
	now := time.Now()
	romance.ActiveUserVote.VoteType = romancesValueObject.VoteTypeYes
	romance.ActiveUserVote.VotedAt = &now
	romance.ActiveUserVote.CreatedAt = &now
	romance.ActiveUserVote.UpdatedAt = &now

	expectedErr := errors.New("database error")

	s.romancesRepo.EXPECT().
		GetRomance(s.ctx, s.voteId).
		Return(romance, nil)

	s.romancesRepo.EXPECT().
		ChangeActiveUserVoteTypeInRomance(s.ctx, romance, romancesValueObject.VoteTypeCrush).
		Return(romanceEntity.Romance{}, expectedErr)

	operation := s.newOperation()
	vote, err := operation.Run(s.ctx, s.voteId, romancesValueObject.VoteTypeCrush)

	s.Require().Error(err)
	s.Require().ErrorIs(err, expectedErr)
	s.Require().Equal(romanceEntity.Vote{}, vote)
}

func (s *ChangeUserVoteOperationUnitTestSuite) TestVersionConflictRetriesAndSucceeds() {
	romance := romanceEntity.CreateEmptyRomance(s.voteId)
	now := time.Now()
	romance.ActiveUserVote.VoteType = romancesValueObject.VoteTypeYes
	romance.ActiveUserVote.VotedAt = &now
	romance.ActiveUserVote.CreatedAt = &now
	romance.ActiveUserVote.UpdatedAt = &now

	// First call to GetRomance
	s.romancesRepo.EXPECT().
		GetRomance(s.ctx, s.voteId).
		Return(romance, nil)

	// First ChangeActiveUserVoteTypeInRomance fails with version conflict
	s.romancesRepo.EXPECT().
		ChangeActiveUserVoteTypeInRomance(s.ctx, gomock.Any(), romancesValueObject.VoteTypeCrush).
		Return(romanceEntity.Romance{}, romanceDomain.ErrVersionConflict)

	// Second call to GetRomance (retry)
	s.romancesRepo.EXPECT().
		GetRomance(s.ctx, s.voteId).
		Return(romance, nil)

	// Second ChangeActiveUserVoteTypeInRomance succeeds
	updatedRomance := romance
	updatedRomance.ActiveUserVote.VoteType = romancesValueObject.VoteTypeCrush
	s.romancesRepo.EXPECT().
		ChangeActiveUserVoteTypeInRomance(s.ctx, gomock.Any(), romancesValueObject.VoteTypeCrush).
		Return(updatedRomance, nil)

	operation := s.newOperation()
	vote, err := operation.Run(s.ctx, s.voteId, romancesValueObject.VoteTypeCrush)

	s.Require().NoError(err)
	s.Require().Equal(romancesValueObject.VoteTypeCrush, vote.VoteType)
}

func (s *ChangeUserVoteOperationUnitTestSuite) TestInvalidVoteTypeTransition() {
	romance := romanceEntity.CreateEmptyRomance(s.voteId)
	now := time.Now()
	// Crush is a terminal state - can't change to anything
	romance.ActiveUserVote.VoteType = romancesValueObject.VoteTypeCrush
	romance.ActiveUserVote.VotedAt = &now
	romance.ActiveUserVote.CreatedAt = &now
	romance.ActiveUserVote.UpdatedAt = &now

	s.romancesRepo.EXPECT().
		GetRomance(s.ctx, s.voteId).
		Return(romance, nil)

	operation := s.newOperation()
	vote, err := operation.Run(s.ctx, s.voteId, romancesValueObject.VoteTypeYes)

	s.Require().Error(err)
	s.Require().Contains(err.Error(), "wrong vote")
	s.Require().Equal(romanceEntity.Vote{}, vote)
}

func (s *ChangeUserVoteOperationUnitTestSuite) TestChangeVoteSuccessfully() {
	// Test all valid vote transitions from the allowed transitions map
	testCases := []struct {
		name     string
		fromType romancesValueObject.VoteType
		toType   romancesValueObject.VoteType
	}{
		// From Empty
		{name: "Empty to No", fromType: romancesValueObject.VoteTypeEmpty, toType: romancesValueObject.VoteTypeNo},
		{name: "Empty to Yes", fromType: romancesValueObject.VoteTypeEmpty, toType: romancesValueObject.VoteTypeYes},
		{name: "Empty to Crush", fromType: romancesValueObject.VoteTypeEmpty, toType: romancesValueObject.VoteTypeCrush},
		{name: "Empty to Compliment", fromType: romancesValueObject.VoteTypeEmpty, toType: romancesValueObject.VoteTypeCompliment},
		// From No
		{name: "No to Yes", fromType: romancesValueObject.VoteTypeNo, toType: romancesValueObject.VoteTypeYes},
		{name: "No to Crush", fromType: romancesValueObject.VoteTypeNo, toType: romancesValueObject.VoteTypeCrush},
		{name: "No to Compliment", fromType: romancesValueObject.VoteTypeNo, toType: romancesValueObject.VoteTypeCompliment},
		// From Yes
		{name: "Yes to Crush", fromType: romancesValueObject.VoteTypeYes, toType: romancesValueObject.VoteTypeCrush},
		{name: "Yes to Compliment", fromType: romancesValueObject.VoteTypeYes, toType: romancesValueObject.VoteTypeCompliment},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			romance := romanceEntity.CreateEmptyRomance(s.voteId)
			now := time.Now()
			romance.ActiveUserVote.VoteType = tc.fromType
			romance.ActiveUserVote.VotedAt = &now
			romance.ActiveUserVote.CreatedAt = &now
			romance.ActiveUserVote.UpdatedAt = &now

			s.romancesRepo.EXPECT().
				GetRomance(s.ctx, s.voteId).
				Return(romance, nil)

			updatedRomance := romance
			updatedRomance.ActiveUserVote.VoteType = tc.toType
			s.romancesRepo.EXPECT().
				ChangeActiveUserVoteTypeInRomance(s.ctx, romance, tc.toType).
				Return(updatedRomance, nil)

			operation := s.newOperation()
			vote, err := operation.Run(s.ctx, s.voteId, tc.toType)

			s.Require().NoError(err)
			s.Require().Equal(tc.toType, vote.VoteType)
		})
	}
}

func (s *ChangeUserVoteOperationUnitTestSuite) TestIsVoteTypeChangeAllowedExhaustive() {
	// Define all vote types
	allVoteTypes := []romancesValueObject.VoteType{
		romancesValueObject.VoteTypeEmpty,
		romancesValueObject.VoteTypeNo,
		romancesValueObject.VoteTypeYes,
		romancesValueObject.VoteTypeCrush,
		romancesValueObject.VoteTypeCompliment,
	}

	// Define expected valid transitions
	validTransitions := map[romancesValueObject.VoteType][]romancesValueObject.VoteType{
		romancesValueObject.VoteTypeEmpty: {
			romancesValueObject.VoteTypeNo,
			romancesValueObject.VoteTypeYes,
			romancesValueObject.VoteTypeCrush,
			romancesValueObject.VoteTypeCompliment,
		},
		romancesValueObject.VoteTypeNo: {
			romancesValueObject.VoteTypeYes,
			romancesValueObject.VoteTypeCrush,
			romancesValueObject.VoteTypeCompliment,
		},
		romancesValueObject.VoteTypeYes: {
			romancesValueObject.VoteTypeCrush,
			romancesValueObject.VoteTypeCompliment,
		},
		romancesValueObject.VoteTypeCrush:      {},
		romancesValueObject.VoteTypeCompliment: {},
	}

	// Test all possible combinations
	for _, fromType := range allVoteTypes {
		for _, toType := range allVoteTypes {
			testName := fromType.String() + "_to_" + toType.String()

			s.Run(testName, func() {
				vote := romanceEntity.Vote{VoteType: fromType}
				err := isVoteTypeChangeAllowed(vote, toType)

				// Check if this transition should be valid
				expectedValid := false
				if allowedTypes, exists := validTransitions[fromType]; exists {
					for _, allowed := range allowedTypes {
						if allowed == toType {
							expectedValid = true
							break
						}
					}
				}

				if expectedValid {
					s.Require().NoError(err,
						"Expected transition from %s to %s to be allowed, but got error: %v",
						fromType.String(), toType.String(), err)
				} else {
					s.Require().Error(err,
						"Expected transition from %s to %s to be disallowed, but got no error",
						fromType.String(), toType.String())
				}
			})
		}
	}

	// Test unknown/invalid vote type (defensive programming test)
	s.Run("unknown_vote_type", func() {
		// Create a vote with an invalid/unknown vote type (e.g., 99)
		invalidVote := romanceEntity.Vote{VoteType: romancesValueObject.VoteType(99)}
		err := isVoteTypeChangeAllowed(invalidVote, romancesValueObject.VoteTypeYes)

		// Should return error for unknown vote types
		s.Require().Error(err, "Unknown vote type should not be changeable")
	})
}
