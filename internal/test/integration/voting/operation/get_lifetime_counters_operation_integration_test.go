package operation

import (
	"context"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/testlib/uuidhelper"
	"testing"
	"time"

	"github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/application/operation"
	counterRepository "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/counter/repository"
	countersValueObject "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/counter/valueobject"
	sharedValueObject "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/sharedkernel/valueobject"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/testlib/helper"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

type GetLifetimeCountersOperationIntegrationTestSuite struct {
	suite.Suite
	romancesTableHelper *helper.RomancesTableHelper
	countersTableHelper *helper.CountersTableHelper
	activeUserId        uuid.UUID
	countryId           uint16
	ctx                 context.Context
	countersRepo        counterRepository.CountersRepository
	op                  *operation.GetLifetimeCountersOperation
}

func TestGetLifetimeCountersOperationIntegrationSuite(t *testing.T) {
	suite.Run(t, new(GetLifetimeCountersOperationIntegrationTestSuite))
}

func (s *GetLifetimeCountersOperationIntegrationTestSuite) SetupSuite() {
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

	s.countryId = uint16(11)
	s.ctx = context.Background()
	s.countersRepo = newCountersRepository(ddbClient)
	s.op = operation.NewGetLifetimeCountersOperation(s.countersRepo)
}

func (s *GetLifetimeCountersOperationIntegrationTestSuite) SetupTest() {
	// Create new activeUserId for each test to ensure test isolation
	activeUserId := uuidhelper.NewUUID(s.T())
	s.activeUserId = activeUserId
}

func (s *GetLifetimeCountersOperationIntegrationTestSuite) TestGetLifetimeCountersWhenCountersExist() {
	userKey, err := sharedValueObject.NewActiveUserKey(s.countryId, s.activeUserId)
	s.Require().NoError(err)

	// Setup: Increment counters directly
	currentTime := time.Now().UTC()
	counterUpdateGroup, err := countersValueObject.NewCounterUpdateGroup(currentTime)
	s.Require().NoError(err)

	for i := 0; i < 5; i++ {
		peerId := uuidhelper.NewUUID(s.T())
		voteId, err := sharedValueObject.NewVoteId(s.countryId, s.activeUserId, peerId)
		s.Require().NoError(err)

		s.countersRepo.IncrYesCounters(s.ctx, voteId, counterUpdateGroup)
	}

	// Test: Get lifetime counters
	countersGroup, err := s.op.Run(s.ctx, userKey)

	s.Require().NoError(err)

	// Verify that IncrYesCounters succeeded by checking the actual counter values
	s.Require().Equal(uint32(5), countersGroup.OutgoingYes, "Expected 5 outgoing yes votes to be recorded")
	s.Require().Equal(uint32(0), countersGroup.OutgoingNo, "Expected no outgoing no votes")
	s.Require().Equal(uint32(0), countersGroup.IncomingYes, "Expected no incoming yes votes")
	s.Require().Equal(uint32(0), countersGroup.IncomingNo, "Expected no incoming no votes")
}

func (s *GetLifetimeCountersOperationIntegrationTestSuite) TestGetLifetimeCountersWhenCountersDoNotExist() {
	userKey, err := sharedValueObject.NewActiveUserKey(s.countryId, s.activeUserId)
	s.Require().NoError(err)

	// Test: Get lifetime counters for user with no counters
	countersGroup, err := s.op.Run(s.ctx, userKey)

	s.Require().NoError(err)

	// Should return zero counters when no counters exist
	s.Require().Equal(uint32(0), countersGroup.OutgoingYes)
	s.Require().Equal(uint32(0), countersGroup.OutgoingNo)
	s.Require().Equal(uint32(0), countersGroup.IncomingYes)
	s.Require().Equal(uint32(0), countersGroup.IncomingNo)
}
