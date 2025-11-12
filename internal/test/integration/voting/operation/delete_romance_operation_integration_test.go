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

type DeleteRomanceOperationIntegrationTestSuite struct {
	suite.Suite
	romancesTableHelper *helper.RomancesTableHelper
	voteId              sharedValueObject.VoteId
	ctx                 context.Context
	romancesRepo        romanceRepository.RomancesRepository
	op                  *operation.DeleteRomanceOperation
}

func TestDeleteRomanceOperationIntegrationSuite(t *testing.T) {
	suite.Run(t, new(DeleteRomanceOperationIntegrationTestSuite))
}

func (s *DeleteRomanceOperationIntegrationTestSuite) SetupSuite() {
	romancesTableHelper, err := helper.NewRomancesTableHelper(ddbClient)
	s.Require().NoError(err)
	s.romancesTableHelper = romancesTableHelper

	err = s.romancesTableHelper.CreateRomancesTable()
	s.Require().NoError(err)

	s.ctx = context.Background()
	s.romancesRepo = newRomancesRepository(ddbClient)
	s.op = operation.NewDeleteRomanceOperation(s.romancesRepo)
}

func (s *DeleteRomanceOperationIntegrationTestSuite) SetupTest() {
	// Create new IDs for each test to ensure test isolation
	activeUserId := uuidhelper.NewUUID(s.T())
	peerUserId := uuidhelper.NewUUID(s.T())
	countryId := uint16(11)

	voteId, err := sharedValueObject.NewVoteId(countryId, activeUserId, peerUserId)
	s.Require().NoError(err)
	s.voteId = voteId
}

func (s *DeleteRomanceOperationIntegrationTestSuite) TearDownTest() {
	// Clean up romance data after each test
	err := s.romancesRepo.DeleteRomance(s.ctx, s.voteId)
	s.Require().NoError(err)
}

func (s *DeleteRomanceOperationIntegrationTestSuite) TestDeleteRomanceWhenRomanceExists() {
	// Setup: Create a romance with votes
	romance := romanceEntity.CreateEmptyRomance(s.voteId)
	votedAt := time.Now().UTC()
	_, err := s.romancesRepo.AddActiveUserVoteToRomance(s.ctx, romance, romancesValueObject.VoteTypeYes, votedAt)
	s.Require().NoError(err)

	// Test: Delete the romance
	err = s.op.Run(s.ctx, s.voteId)

	s.Require().NoError(err)

	// Verify romance was deleted - GetRomance should return empty romance
	romance, err = s.romancesRepo.GetRomance(s.ctx, s.voteId)
	s.Require().NoError(err)
	s.Require().Equal(romancesValueObject.VoteTypeEmpty, romance.ActiveUserVote.VoteType)
	s.Require().Equal(uint32(0), romance.Version)
}

func (s *DeleteRomanceOperationIntegrationTestSuite) TestDeleteRomanceWhenRomanceDoesNotExist() {
	// Test: Delete romance when no romance exists (should succeed without error)
	err := s.op.Run(s.ctx, s.voteId)

	s.Require().NoError(err)
}
