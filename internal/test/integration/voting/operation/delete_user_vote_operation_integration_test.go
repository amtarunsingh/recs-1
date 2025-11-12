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
	romanceEntity "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/romance/entity"
	romanceRepository "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/romance/repository"
	romancesValueObject "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/romance/valueobject"
	sharedValueObject "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/sharedkernel/valueobject"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/testlib/helper"
	"github.com/stretchr/testify/suite"
)

type DeleteUserVoteOperationIntegrationTestSuite struct {
	suite.Suite
	romancesTableHelper *helper.RomancesTableHelper
	voteId              sharedValueObject.VoteId
	ctx                 context.Context
	romancesRepo        romanceRepository.RomancesRepository
	countersRepo        counterRepository.CountersRepository
	op                  *operation.DeleteUserVoteOperation
}

func TestDeleteUserVoteOperationIntegrationSuite(t *testing.T) {
	suite.Run(t, new(DeleteUserVoteOperationIntegrationTestSuite))
}

func (s *DeleteUserVoteOperationIntegrationTestSuite) SetupSuite() {
	romancesTableHelper, err := helper.NewRomancesTableHelper(ddbClient)
	s.Require().NoError(err)
	s.romancesTableHelper = romancesTableHelper

	err = s.romancesTableHelper.CreateRomancesTable()
	s.Require().NoError(err)

	s.ctx = context.Background()
	s.romancesRepo = newRomancesRepository(ddbClient)
	s.countersRepo = newCountersRepository(ddbClient)
	s.op = operation.NewDeleteUserVoteOperation(s.romancesRepo, s.countersRepo, slog.New(slog.NewTextHandler(io.Discard, nil)))
}

func (s *DeleteUserVoteOperationIntegrationTestSuite) SetupTest() {
	// Create new IDs for each test to ensure test isolation
	activeUserId := uuidhelper.NewUUID(s.T())
	peerUserId := uuidhelper.NewUUID(s.T())
	countryId := uint16(11)

	voteId, err := sharedValueObject.NewVoteId(countryId, activeUserId, peerUserId)
	s.Require().NoError(err)
	s.voteId = voteId
}

func (s *DeleteUserVoteOperationIntegrationTestSuite) TearDownTest() {
	// Clean up romance data after each test
	err := s.romancesRepo.DeleteRomance(s.ctx, s.voteId)
	s.Require().NoError(err)
}

func (s *DeleteUserVoteOperationIntegrationTestSuite) TestDeleteVoteWhenVoteExists() {
	// Setup: Create a vote
	romance := romanceEntity.CreateEmptyRomance(s.voteId)
	votedAt := time.Now().UTC()
	_, err := s.romancesRepo.AddActiveUserVoteToRomance(s.ctx, romance, romancesValueObject.VoteTypeYes, votedAt)
	s.Require().NoError(err)

	// Test: Delete the vote
	err = s.op.Run(s.ctx, s.voteId)

	s.Require().NoError(err)

	// Verify vote was deleted (should be back to empty vote)
	romance, err = s.romancesRepo.GetRomance(s.ctx, s.voteId)
	s.Require().NoError(err)

	expectedVote := romanceEntity.Vote{
		Id:        s.voteId,
		VoteType:  romancesValueObject.VoteTypeEmpty,
		VotedAt:   nil,
		CreatedAt: nil,
		UpdatedAt: nil,
	}
	s.Require().Equal(expectedVote, romance.ActiveUserVote)
}

func (s *DeleteUserVoteOperationIntegrationTestSuite) TestDeleteVoteWhenVoteDoesNotExist() {
	// Test: Delete vote when no vote exists (should succeed without error)
	err := s.op.Run(s.ctx, s.voteId)

	s.Require().NoError(err)
}
