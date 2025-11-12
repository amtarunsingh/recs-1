package operation

import (
	"context"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/testlib/uuidhelper"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/application/operation"
	counterRepository "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/counter/repository"
	romanceDomain "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/romance"
	romanceRepository "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/romance/repository"
	romancesValueObject "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/romance/valueobject"
	sharedValueObject "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/sharedkernel/valueobject"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/testlib/helper"
	"github.com/stretchr/testify/suite"
)

type AddUserVoteOperationIntegrationTestSuite struct {
	suite.Suite
	romancesTableHelper *helper.RomancesTableHelper
	countersTableHelper *helper.CountersTableHelper
	voteId              sharedValueObject.VoteId
	ctx                 context.Context
	romancesRepo        romanceRepository.RomancesRepository
	countersRepo        counterRepository.CountersRepository
	op                  *operation.AddUserVoteOperation
}

func TestAddUserVoteOperationIntegrationSuite(t *testing.T) {
	suite.Run(t, new(AddUserVoteOperationIntegrationTestSuite))
}

func (s *AddUserVoteOperationIntegrationTestSuite) SetupSuite() {
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
	s.op = operation.NewAddUserVoteOperation(s.romancesRepo, s.countersRepo, slog.New(slog.NewTextHandler(io.Discard, nil)))
}

func (s *AddUserVoteOperationIntegrationTestSuite) SetupTest() {
	// Create new IDs for each test to ensure test isolation
	activeUserId := uuidhelper.NewUUID(s.T())
	peerUserId := uuidhelper.NewUUID(s.T())
	countryId := uint16(11)

	voteId, err := sharedValueObject.NewVoteId(countryId, activeUserId, peerUserId)
	s.Require().NoError(err)
	s.voteId = voteId
}

func (s *AddUserVoteOperationIntegrationTestSuite) TearDownTest() {
	// Clean up romance data after each test
	err := s.romancesRepo.DeleteRomance(s.ctx, s.voteId)
	s.Require().NoError(err)
}

func (s *AddUserVoteOperationIntegrationTestSuite) TestAddFirstYesVote() {
	votedAt := time.Now().UTC()
	vote, err := s.op.Run(s.ctx, s.voteId, romancesValueObject.VoteTypeYes, votedAt)

	s.Require().NoError(err)
	s.Require().Equal(romancesValueObject.VoteTypeYes, vote.VoteType)
	s.Require().NotNil(vote.VotedAt)
	s.Require().Equal(s.voteId, vote.Id)

	// Verify romance was created in database
	romance, err := s.romancesRepo.GetRomance(s.ctx, s.voteId)
	s.Require().NoError(err)
	s.Require().Equal(romancesValueObject.VoteTypeYes, romance.ActiveUserVote.VoteType)
	s.Require().Equal(uint32(1), romance.Version)

	// Verify outgoing yes counter was incremented
	activeUserKey, err := sharedValueObject.NewActiveUserKey(s.voteId.CountryId(), s.voteId.ActiveUserId())
	s.Require().NoError(err)
	counters, err := s.countersRepo.GetLifetimeCounter(s.ctx, activeUserKey)
	s.Require().NoError(err)
	s.Require().Equal(uint32(1), counters.OutgoingYes)
	s.Require().Equal(uint32(0), counters.OutgoingNo)
}

func (s *AddUserVoteOperationIntegrationTestSuite) TestAddFirstNoVote() {
	votedAt := time.Now().UTC()
	vote, err := s.op.Run(s.ctx, s.voteId, romancesValueObject.VoteTypeNo, votedAt)

	s.Require().NoError(err)
	s.Require().Equal(romancesValueObject.VoteTypeNo, vote.VoteType)
	s.Require().NotNil(vote.VotedAt)
	s.Require().Equal(s.voteId, vote.Id)

	// Verify romance was created in database
	romance, err := s.romancesRepo.GetRomance(s.ctx, s.voteId)
	s.Require().NoError(err)
	s.Require().Equal(romancesValueObject.VoteTypeNo, romance.ActiveUserVote.VoteType)
	s.Require().Equal(uint32(1), romance.Version)

	// Verify outgoing no counter was incremented
	activeUserKey, err := sharedValueObject.NewActiveUserKey(s.voteId.CountryId(), s.voteId.ActiveUserId())
	s.Require().NoError(err)
	counters, err := s.countersRepo.GetLifetimeCounter(s.ctx, activeUserKey)
	s.Require().NoError(err)
	s.Require().Equal(uint32(0), counters.OutgoingYes)
	s.Require().Equal(uint32(1), counters.OutgoingNo)
}

func (s *AddUserVoteOperationIntegrationTestSuite) TestAddInvalidVoteTransition() {
	// Setup: Add a CRUSH vote (terminal state)
	votedAt := time.Now().UTC()
	_, err := s.op.Run(s.ctx, s.voteId, romancesValueObject.VoteTypeCrush, votedAt)
	s.Require().NoError(err)

	// Test: Try to add a YES vote (invalid transition from Crush)
	_, err = s.op.Run(s.ctx, s.voteId, romancesValueObject.VoteTypeYes, votedAt)

	s.Require().Error(err)
	s.Require().ErrorIs(err, romanceDomain.ErrWrongVote)
	s.Require().Contains(err.Error(), "vote type change from `crush` to `yes` is not allowed")
}

func (s *AddUserVoteOperationIntegrationTestSuite) TestValidVoteTransition() {
	// Setup: Add a NO vote
	votedAt := time.Now().UTC()
	_, err := s.op.Run(s.ctx, s.voteId, romancesValueObject.VoteTypeNo, votedAt)
	s.Require().NoError(err)

	// Test: Change to YES vote (valid transition)
	vote, err := s.op.Run(s.ctx, s.voteId, romancesValueObject.VoteTypeYes, votedAt)

	s.Require().NoError(err)
	s.Require().Equal(romancesValueObject.VoteTypeYes, vote.VoteType)

	// Verify in database
	romance, err := s.romancesRepo.GetRomance(s.ctx, s.voteId)
	s.Require().NoError(err)
	s.Require().Equal(romancesValueObject.VoteTypeYes, romance.ActiveUserVote.VoteType)
	s.Require().Equal(uint32(2), romance.Version)

	// Verify counters reflect both transitions (No from first vote, Yes from second vote)
	activeUserKey, err := sharedValueObject.NewActiveUserKey(s.voteId.CountryId(), s.voteId.ActiveUserId())
	s.Require().NoError(err)
	counters, err := s.countersRepo.GetLifetimeCounter(s.ctx, activeUserKey)
	s.Require().NoError(err)
	s.Require().Equal(uint32(1), counters.OutgoingYes)
	s.Require().Equal(uint32(1), counters.OutgoingNo) // AddUserVoteOperation only increments, never decrements
}
