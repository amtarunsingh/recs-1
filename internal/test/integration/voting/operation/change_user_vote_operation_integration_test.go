package operation

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/application/operation"
	counterRepository "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/counter/repository"
	romanceDomain "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/romance"
	romanceEntity "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/romance/entity"
	romanceRepository "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/romance/repository"
	romancesValueObject "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/romance/valueobject"
	sharedValueObject "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/sharedkernel/valueobject"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/testlib/helper"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/testlib/uuidhelper"
	"github.com/stretchr/testify/suite"
)

type ChangeUserVoteOperationIntegrationTestSuite struct {
	suite.Suite
	romancesTableHelper *helper.RomancesTableHelper
	countersTableHelper *helper.CountersTableHelper
	voteId              sharedValueObject.VoteId
	ctx                 context.Context
	romancesRepo        romanceRepository.RomancesRepository
	countersRepo        counterRepository.CountersRepository
	op                  *operation.ChangeUserVoteOperation
}

func TestChangeUserVoteOperationIntegrationSuite(t *testing.T) {
	suite.Run(t, new(ChangeUserVoteOperationIntegrationTestSuite))
}

func (s *ChangeUserVoteOperationIntegrationTestSuite) SetupSuite() {
	romancesTableHelper, err := helper.NewRomancesTableHelper(ddbClient)
	s.Require().NoError(err)
	s.romancesTableHelper = romancesTableHelper

	countersTableHelper, err := helper.NewCountersTableHelper(ddbClient)
	s.Require().NoError(err)
	s.countersTableHelper = countersTableHelper

	err = s.romancesTableHelper.CreateRomancesTable()
	s.Require().NoError(err)

	err = s.countersTableHelper.CreateCountersTable()
	s.Require().NoError(err)

	s.ctx = context.Background()
	s.romancesRepo = newRomancesRepository(ddbClient)
	s.countersRepo = newCountersRepository(ddbClient)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	s.op = operation.NewChangeUserVoteOperation(s.romancesRepo, s.countersRepo, logger)
}

func (s *ChangeUserVoteOperationIntegrationTestSuite) SetupTest() {
	// Create new IDs for each test to ensure test isolation
	activeUserId := uuidhelper.NewUUID(s.T())
	peerUserId := uuidhelper.NewUUID(s.T())
	countryId := uint16(11)

	voteId, err := sharedValueObject.NewVoteId(countryId, activeUserId, peerUserId)
	s.Require().NoError(err)
	s.voteId = voteId
}

func (s *ChangeUserVoteOperationIntegrationTestSuite) TearDownTest() {
	// Clean up romance data after each test
	err := s.romancesRepo.DeleteRomance(s.ctx, s.voteId)
	s.Require().NoError(err)
}

func (s *ChangeUserVoteOperationIntegrationTestSuite) TestChangeInvalidVoteTransition() {
	// Setup: Add a CRUSH vote (terminal state) directly via repository
	romance := romanceEntity.CreateEmptyRomance(s.voteId)
	votedAt := time.Now().UTC()
	_, err := s.romancesRepo.AddActiveUserVoteToRomance(s.ctx, romance, romancesValueObject.VoteTypeCrush, votedAt)
	s.Require().NoError(err)

	// Test: Try to change to YES vote (invalid transition from Crush)
	_, err = s.op.Run(s.ctx, s.voteId, romancesValueObject.VoteTypeYes)

	s.Require().Error(err)
	s.Require().ErrorIs(err, romanceDomain.ErrWrongVote)
	s.Require().Contains(err.Error(), "vote type change from `crush` to `yes` is not allowed")
}

func (s *ChangeUserVoteOperationIntegrationTestSuite) TestValidVoteTransition() {
	// Setup: Add a NO vote directly via repository (without incrementing counters)
	romance := romanceEntity.CreateEmptyRomance(s.voteId)
	votedAt := time.Now().UTC()
	_, err := s.romancesRepo.AddActiveUserVoteToRomance(s.ctx, romance, romancesValueObject.VoteTypeNo, votedAt)
	s.Require().NoError(err)

	// Verify initial counters are all zero (repository doesn't increment counters)
	activeUserKey, err := sharedValueObject.NewActiveUserKey(s.voteId.CountryId(), s.voteId.ActiveUserId())
	s.Require().NoError(err)
	countersBefore, err := s.countersRepo.GetLifetimeCounter(s.ctx, activeUserKey)
	s.Require().NoError(err)
	s.Require().Equal(uint32(0), countersBefore.OutgoingYes)
	s.Require().Equal(uint32(0), countersBefore.OutgoingNo)
	s.Require().Equal(uint32(0), countersBefore.IncomingYes)
	s.Require().Equal(uint32(0), countersBefore.IncomingNo)

	// Test: Change to YES vote (valid transition)
	vote, err := s.op.Run(s.ctx, s.voteId, romancesValueObject.VoteTypeYes)

	s.Require().NoError(err)
	s.Require().Equal(romancesValueObject.VoteTypeYes, vote.VoteType)

	// Verify in database
	romance, err = s.romancesRepo.GetRomance(s.ctx, s.voteId)
	s.Require().NoError(err)
	s.Require().Equal(romancesValueObject.VoteTypeYes, romance.ActiveUserVote.VoteType)
	s.Require().Equal(uint32(2), romance.Version)

	// Verify counters remain at zero (change operation should not modify counters)
	countersAfter, err := s.countersRepo.GetLifetimeCounter(s.ctx, activeUserKey)
	s.Require().NoError(err)
	s.Require().Equal(uint32(0), countersAfter.OutgoingYes)
	s.Require().Equal(uint32(0), countersAfter.OutgoingNo)
	s.Require().Equal(uint32(0), countersAfter.IncomingYes)
	s.Require().Equal(uint32(0), countersAfter.IncomingNo)
}
