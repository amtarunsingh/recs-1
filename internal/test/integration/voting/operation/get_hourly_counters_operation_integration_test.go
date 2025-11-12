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

type GetHourlyCountersOperationIntegrationTestSuite struct {
	suite.Suite
	romancesTableHelper *helper.RomancesTableHelper
	countersTableHelper *helper.CountersTableHelper
	activeUserId        uuid.UUID
	countryId           uint16
	ctx                 context.Context
	countersRepo        counterRepository.CountersRepository
	op                  *operation.GetHourlyCountersOperation
}

func TestGetHourlyCountersOperationIntegrationSuite(t *testing.T) {
	suite.Run(t, new(GetHourlyCountersOperationIntegrationTestSuite))
}

func (s *GetHourlyCountersOperationIntegrationTestSuite) SetupSuite() {
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
	s.op = operation.NewGetHourlyCountersOperation(s.countersRepo)
}

func (s *GetHourlyCountersOperationIntegrationTestSuite) SetupTest() {
	// Create new activeUserId for each test to ensure test isolation
	activeUserId := uuidhelper.NewUUID(s.T())
	s.activeUserId = activeUserId
}

func (s *GetHourlyCountersOperationIntegrationTestSuite) TestGetHourlyCountersWhenCountersExist() {

	userKey, err := sharedValueObject.NewActiveUserKey(s.countryId, s.activeUserId)
	s.Require().NoError(err)

	// Setup: Increment counters directly for 1 hour ago
	oneHourAgo := time.Now().UTC().Add(-1 * time.Hour)
	counterUpdateGroup, err := countersValueObject.NewCounterUpdateGroup(oneHourAgo)
	s.Require().NoError(err)

	for i := 0; i < 3; i++ {
		peerId := uuidhelper.NewUUID(s.T())
		voteId, err := sharedValueObject.NewVoteId(s.countryId, s.activeUserId, peerId)
		s.Require().NoError(err)

		s.countersRepo.IncrYesCounters(s.ctx, voteId, counterUpdateGroup)
	}

	// Test: Get hourly counters for 1 hour ago (offset 1 means 1 hour ago)
	hoursOffsetGroups, err := countersValueObject.NewHoursOffsetGroups([]uint8{1})
	s.Require().NoError(err)

	countersGroups, err := s.op.Run(s.ctx, userKey, hoursOffsetGroups)

	s.Require().NoError(err)
	s.Require().Len(countersGroups, 1)

	// Verify that IncrYesCounters succeeded by checking the actual counter values
	group := countersGroups[1]
	s.Require().NotNil(group, "Counter group should not be nil")
	s.Require().Equal(uint32(3), group.OutgoingYes, "Expected 3 outgoing yes votes to be recorded")
	s.Require().Equal(uint32(0), group.OutgoingNo, "Expected no outgoing no votes")
	s.Require().Equal(uint32(0), group.IncomingYes, "Expected no incoming yes votes")
	s.Require().Equal(uint32(0), group.IncomingNo, "Expected no incoming no votes")
}

func (s *GetHourlyCountersOperationIntegrationTestSuite) TestGetHourlyCountersWhenCountersDoNotExist() {
	userKey, err := sharedValueObject.NewActiveUserKey(s.countryId, s.activeUserId)
	s.Require().NoError(err)

	// Test: Get hourly counters for user with no counters
	hoursOffsetGroups, err := countersValueObject.NewHoursOffsetGroups([]uint8{1, 2, 3})
	s.Require().NoError(err)

	countersGroups, err := s.op.Run(s.ctx, userKey, hoursOffsetGroups)

	s.Require().NoError(err)
	// Repository returns map with zero-value counters for requested hours even when no data exists
	s.Require().Len(countersGroups, 3)
	for _, group := range countersGroups {
		s.Require().Equal(uint32(0), group.OutgoingYes)
		s.Require().Equal(uint32(0), group.OutgoingNo)
		s.Require().Equal(uint32(0), group.IncomingYes)
		s.Require().Equal(uint32(0), group.IncomingNo)
	}
}
