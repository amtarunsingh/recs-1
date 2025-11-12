package persistence

import (
	"context"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/testlib/uuidhelper"
	"io"
	"log/slog"
	"testing"

	"github.com/bmbl-bumble2/recs-votes-storage/config"
	counterEntity "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/counter/entity"
	countersRepository "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/counter/repository"
	sharedValueObject "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/sharedkernel/valueobject"
	infraDynamodb "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/infrastructure/persistence"
	platformDynamodb "github.com/bmbl-bumble2/recs-votes-storage/internal/shared/platform/dynamodb"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/testlib"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/testlib/helper"
	"github.com/stretchr/testify/suite"
)

type CountersRepositoryTestSuite struct {
	suite.Suite
	countersTableHelper *helper.CountersTableHelper
	activeUserKey       sharedValueObject.ActiveUserKey
}

func TestMyCountersRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(CountersRepositoryTestSuite))
}

func (s *CountersRepositoryTestSuite) SetupSuite() {
	countersTableHelper, err := helper.NewCountersTableHelper(ddbClient)
	s.Require().NoError(err)
	s.countersTableHelper = countersTableHelper

	err = s.countersTableHelper.CreateCountersTable()
	s.Require().NoError(err)

	activeUserId := uuidhelper.NewUUID(s.T())
	countryId := uint16(11)

	activeUserKey, err := sharedValueObject.NewActiveUserKey(countryId, activeUserId)
	s.Require().NoError(err)
	s.activeUserKey = activeUserKey
}

func (s *CountersRepositoryTestSuite) TearDownSuite() {
	err := s.countersTableHelper.DeleteAllUserRecords(s.activeUserKey)
	s.Require().NoError(err)
}

func (s *CountersRepositoryTestSuite) SetupTest() {
	err := s.countersTableHelper.DeleteAllUserRecords(s.activeUserKey)
	s.Require().NoError(err)
}

func (s *CountersRepositoryTestSuite) TestGetEmptyLifetimeCounter() {
	repo := newCountersRepository(ddbClient)
	countersGroup, err := repo.GetLifetimeCounter(context.Background(), s.activeUserKey)
	s.Require().NoError(err)
	s.assertEmptyCountersGroup(s.activeUserKey, countersGroup)
}

func newCountersRepository(client platformDynamodb.Client) countersRepository.CountersRepository {
	appConfig := config.Load()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return infraDynamodb.NewCountersRepository(client, appConfig, logger)
}

// func (s *CountersRepositoryTestSuite) assertNilCountersGroup(countersGroup counterEntity.CountersGroup) {
// 	assert.Empty(s.T(), countersGroup)
// }

func (s *CountersRepositoryTestSuite) assertEmptyCountersGroup(
	activeUserKey sharedValueObject.ActiveUserKey,
	countersGroup counterEntity.CountersGroup,
) {
	expected := map[string]any{
		"activeUserKey":     activeUserKey,
		"hourUnixTimestamp": 0,
		"outgoingYes":       0,
		"outgoingNo":        0,
		"incomingYes":       0,
		"incomingNo":        0,
	}

	actual := map[string]any{
		"activeUserKey":     countersGroup.ActiveUserKey,
		"hourUnixTimestamp": countersGroup.HourUnixTimestamp,
		"outgoingYes":       countersGroup.OutgoingYes,
		"outgoingNo":        countersGroup.OutgoingNo,
		"incomingYes":       countersGroup.IncomingYes,
		"incomingNo":        countersGroup.IncomingNo,
	}

	testlib.AssertMap(s.T(), expected, actual)
}
