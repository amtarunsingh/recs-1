package operation

import (
	"context"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/testlib/uuidhelper"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/application/operation"
	romanceEntity "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/romance/entity"
	romancesValueObject "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/romance/valueobject"
	sharedValueObject "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/sharedkernel/valueobject"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/testlib/helper"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

type DeleteRomancesGroupOperationIntegrationTestSuite struct {
	suite.Suite
	romancesTableHelper *helper.RomancesTableHelper
	activeUserId        uuid.UUID
	countryId           uint16
	ctx                 context.Context
}

func TestDeleteRomancesGroupOperationIntegrationSuite(t *testing.T) {
	suite.Run(t, new(DeleteRomancesGroupOperationIntegrationTestSuite))
}

func (s *DeleteRomancesGroupOperationIntegrationTestSuite) SetupSuite() {
	romancesTableHelper, err := helper.NewRomancesTableHelper(ddbClient)
	s.Require().NoError(err)
	s.romancesTableHelper = romancesTableHelper

	err = s.romancesTableHelper.CreateRomancesTable()
	s.Require().NoError(err)

	s.countryId = uint16(11)
	s.ctx = context.Background()
}

func (s *DeleteRomancesGroupOperationIntegrationTestSuite) SetupTest() {
	// Create new activeUserId for each test to ensure test isolation
	activeUserId := uuidhelper.NewUUID(s.T())
	s.activeUserId = activeUserId
}

func (s *DeleteRomancesGroupOperationIntegrationTestSuite) TestDeleteRomancesGroupWithMultipleRomances() {
	repo := newRomancesRepository(ddbClient)
	op := operation.NewDeleteRomancesGroupOperation(repo, slog.New(slog.NewTextHandler(io.Discard, nil)))

	// Setup: Create 3 romances with different peers
	peerIds := []uuid.UUID{}
	for i := 0; i < 3; i++ {
		peerId := uuidhelper.NewUUID(s.T())
		peerIds = append(peerIds, peerId)

		voteId, err := sharedValueObject.NewVoteId(s.countryId, s.activeUserId, peerId)
		s.Require().NoError(err)

		romance := romanceEntity.CreateEmptyRomance(voteId)
		votedAt := time.Now().UTC()
		_, err = repo.AddActiveUserVoteToRomance(s.ctx, romance, romancesValueObject.VoteTypeYes, votedAt)
		s.Require().NoError(err)
	}

	userKey, err := sharedValueObject.NewActiveUserKey(s.countryId, s.activeUserId)
	s.Require().NoError(err)

	// Test: Delete all romances in the group
	err = op.Run(s.ctx, userKey, peerIds)

	s.Require().NoError(err)

	// Verify all romances were deleted
	for _, peerId := range peerIds {
		voteId, err := sharedValueObject.NewVoteId(s.countryId, s.activeUserId, peerId)
		s.Require().NoError(err)

		romance, err := repo.GetRomance(s.ctx, voteId)
		s.Require().NoError(err)
		s.Require().Equal(romancesValueObject.VoteTypeEmpty, romance.ActiveUserVote.VoteType)
		s.Require().Equal(uint32(0), romance.Version)
	}
}

func (s *DeleteRomancesGroupOperationIntegrationTestSuite) TestDeleteRomancesGroupWithEmptyPeerIds() {
	repo := newRomancesRepository(ddbClient)
	op := operation.NewDeleteRomancesGroupOperation(repo, slog.New(slog.NewTextHandler(io.Discard, nil)))

	userKey, err := sharedValueObject.NewActiveUserKey(s.countryId, s.activeUserId)
	s.Require().NoError(err)

	// Test: Delete with empty peer IDs (should succeed without error)
	err = op.Run(s.ctx, userKey, []uuid.UUID{})

	s.Require().NoError(err)
}
