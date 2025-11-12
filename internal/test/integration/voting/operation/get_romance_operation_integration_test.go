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

type GetRomanceOperationIntegrationTestSuite struct {
	suite.Suite
	romancesTableHelper *helper.RomancesTableHelper
	voteId              sharedValueObject.VoteId
	ctx                 context.Context
	romancesRepo        romanceRepository.RomancesRepository
	op                  *operation.GetRomanceOperation
}

func TestGetRomanceOperationIntegrationSuite(t *testing.T) {
	suite.Run(t, new(GetRomanceOperationIntegrationTestSuite))
}

func (s *GetRomanceOperationIntegrationTestSuite) SetupSuite() {
	romancesTableHelper, err := helper.NewRomancesTableHelper(ddbClient)
	s.Require().NoError(err)
	s.romancesTableHelper = romancesTableHelper

	err = s.romancesTableHelper.CreateRomancesTable()
	s.Require().NoError(err)

	s.ctx = context.Background()
	s.romancesRepo = newRomancesRepository(ddbClient)
	s.op = operation.NewGetRomanceOperation(s.romancesRepo)
}

func (s *GetRomanceOperationIntegrationTestSuite) SetupTest() {
	// Create new IDs for each test to ensure test isolation
	activeUserId := uuidhelper.NewUUID(s.T())
	peerUserId := uuidhelper.NewUUID(s.T())
	countryId := uint16(11)

	voteId, err := sharedValueObject.NewVoteId(countryId, activeUserId, peerUserId)
	s.Require().NoError(err)
	s.voteId = voteId
}

func (s *GetRomanceOperationIntegrationTestSuite) TearDownTest() {
	// Clean up romance data after each test
	err := s.romancesRepo.DeleteRomance(s.ctx, s.voteId)
	s.Require().NoError(err)
}

func (s *GetRomanceOperationIntegrationTestSuite) TestGetRomanceWhenRomanceDoesNotExist() {
	romance, err := s.op.Run(s.ctx, s.voteId)

	s.Require().NoError(err)

	// Should return empty romance
	expectedRomance := romanceEntity.CreateEmptyRomance(s.voteId)
	s.Require().Equal(expectedRomance, romance)
}

func (s *GetRomanceOperationIntegrationTestSuite) TestGetRomanceWhenRomanceExists() {
	// Setup: Create a romance with votes from both sides
	activeRomance := romanceEntity.CreateEmptyRomance(s.voteId)
	votedAt := time.Now().UTC()
	_, err := s.romancesRepo.AddActiveUserVoteToRomance(s.ctx, activeRomance, romancesValueObject.VoteTypeYes, votedAt)
	s.Require().NoError(err)

	// Add peer vote
	peerVoteId := s.voteId.ToPeerVoteId()
	peerRomance, err := s.romancesRepo.GetRomance(s.ctx, peerVoteId)
	s.Require().NoError(err)
	_, err = s.romancesRepo.AddActiveUserVoteToRomance(s.ctx, peerRomance, romancesValueObject.VoteTypeNo, votedAt)
	s.Require().NoError(err)

	// Test: Get the romance from active user perspective
	romance, err := s.op.Run(s.ctx, s.voteId)

	s.Require().NoError(err)
	s.Require().Equal(romancesValueObject.VoteTypeYes, romance.ActiveUserVote.VoteType)
	s.Require().Equal(romancesValueObject.VoteTypeNo, romance.PeerUserVote.VoteType)
	s.Require().Equal(uint32(2), romance.Version)
	s.Require().Equal(s.voteId, romance.ActiveUserVote.Id)
}
