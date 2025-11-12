package operation

import (
	"context"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/testlib/uuidhelper"
	"testing"
	"time"

	"github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/application/operation"
	romanceEntity "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/romance/entity"
	romanceRepository "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/romance/repository"
	romancesValueObject "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/romance/valueobject"
	sharedValueObject "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/sharedkernel/valueobject"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/testlib/helper"
	"github.com/stretchr/testify/suite"
)

type GetUserVoteOperationIntegrationTestSuite struct {
	suite.Suite
	romancesTableHelper *helper.RomancesTableHelper
	voteId              sharedValueObject.VoteId
	ctx                 context.Context
	romancesRepo        romanceRepository.RomancesRepository
	op                  *operation.GetUserVoteOperation
}

func TestGetUserVoteOperationIntegrationSuite(t *testing.T) {
	suite.Run(t, new(GetUserVoteOperationIntegrationTestSuite))
}

func (s *GetUserVoteOperationIntegrationTestSuite) SetupSuite() {
	romancesTableHelper, err := helper.NewRomancesTableHelper(ddbClient)
	s.Require().NoError(err)
	s.romancesTableHelper = romancesTableHelper

	err = s.romancesTableHelper.CreateRomancesTable()
	s.Require().NoError(err)

	s.ctx = context.Background()
	s.romancesRepo = newRomancesRepository(ddbClient)
	s.op = operation.NewGetUserVoteOperation(s.romancesRepo)
}

func (s *GetUserVoteOperationIntegrationTestSuite) SetupTest() {
	// Create new IDs for each test to ensure test isolation
	activeUserId := uuidhelper.NewUUID(s.T())
	peerUserId := uuidhelper.NewUUID(s.T())
	countryId := uint16(11)

	voteId, err := sharedValueObject.NewVoteId(countryId, activeUserId, peerUserId)
	s.Require().NoError(err)
	s.voteId = voteId
}

func (s *GetUserVoteOperationIntegrationTestSuite) TearDownTest() {
	// Clean up romance data after each test
	err := s.romancesRepo.DeleteRomance(s.ctx, s.voteId)
	s.Require().NoError(err)
}

func (s *GetUserVoteOperationIntegrationTestSuite) TestGetVoteWhenVoteDoesNotExist() {
	vote, err := s.op.Run(s.ctx, s.voteId)

	s.Require().NoError(err)

	expectedVote := romanceEntity.Vote{
		Id:        s.voteId,
		VoteType:  romancesValueObject.VoteTypeEmpty,
		VotedAt:   nil,
		CreatedAt: nil,
		UpdatedAt: nil,
	}
	s.Require().Equal(expectedVote, vote)
}

func (s *GetUserVoteOperationIntegrationTestSuite) TestGetVoteWhenVoteExists() {
	// Setup: Create a romance with a vote
	romance := romanceEntity.CreateEmptyRomance(s.voteId)
	votedAt := time.Now().UTC()
	updatedRomance, err := s.romancesRepo.AddActiveUserVoteToRomance(s.ctx, romance, romancesValueObject.VoteTypeYes, votedAt)
	s.Require().NoError(err)

	// Test: Get the vote
	vote, err := s.op.Run(s.ctx, s.voteId)

	s.Require().NoError(err)
	s.Require().Equal(romancesValueObject.VoteTypeYes, vote.VoteType)
	s.Require().Equal(s.voteId, vote.Id)
	// Verify the vote matches what was returned from AddActiveUserVoteToRomance
	s.Require().Equal(updatedRomance.ActiveUserVote, vote)
}
