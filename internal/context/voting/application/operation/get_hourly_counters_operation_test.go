package operation

import (
	"context"
	"errors"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/testlib/uuidhelper"
	"testing"

	counterEntity "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/counter/entity"
	countersValueObject "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/counter/valueobject"
	sharedValueObject "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/sharedkernel/valueobject"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/testlib/mocks"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

type GetHourlyCountersOperationUnitTestSuite struct {
	suite.Suite
	activeUserKey sharedValueObject.ActiveUserKey
	ctrl          *gomock.Controller
	countersRepo  *mocks.MockCountersRepository
	ctx           context.Context
}

func TestGetHourlyCountersOperationUnitSuite(t *testing.T) {
	suite.Run(t, new(GetHourlyCountersOperationUnitTestSuite))
}

func (s *GetHourlyCountersOperationUnitTestSuite) SetupSuite() {
	activeUserId := uuidhelper.NewUUID(s.T())
	countryId := uint16(11)

	activeUserKey, err := sharedValueObject.NewActiveUserKey(countryId, activeUserId)
	s.Require().NoError(err)
	s.activeUserKey = activeUserKey
	s.ctx = context.Background()
}

func (s *GetHourlyCountersOperationUnitTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.countersRepo = mocks.NewMockCountersRepository(s.ctrl)
}

func (s *GetHourlyCountersOperationUnitTestSuite) newOperation() *GetHourlyCountersOperation {
	return NewGetHourlyCountersOperation(s.countersRepo)
}

func (s *GetHourlyCountersOperationUnitTestSuite) TestGetHourlyCountersReturnsError() {
	expectedErr := errors.New("database error")
	hoursOffsetGroups, err := countersValueObject.NewHoursOffsetGroups([]uint8{1, 2, 3})
	s.Require().NoError(err)

	s.countersRepo.EXPECT().
		GetHourlyCounters(s.ctx, s.activeUserKey, hoursOffsetGroups).
		Return(map[uint8]*counterEntity.CountersGroup{}, expectedErr)

	operation := s.newOperation()
	counters, err := operation.Run(s.ctx, s.activeUserKey, hoursOffsetGroups)

	s.Require().Error(err)
	s.Require().ErrorIs(err, expectedErr)
	s.Require().Equal(map[uint8]*counterEntity.CountersGroup{}, counters)
}

func (s *GetHourlyCountersOperationUnitTestSuite) TestGetHourlyCountersSuccessfully() {
	hoursOffsetGroups, err := countersValueObject.NewHoursOffsetGroups([]uint8{1, 2, 3})
	s.Require().NoError(err)
	expectedCounters := map[uint8]*counterEntity.CountersGroup{
		1: {
			IncomingYes: 5,
			IncomingNo:  2,
			OutgoingYes: 3,
			OutgoingNo:  1,
		},
		2: {
			IncomingYes: 8,
			IncomingNo:  4,
			OutgoingYes: 6,
			OutgoingNo:  2,
		},
		3: {
			IncomingYes: 3,
			IncomingNo:  1,
			OutgoingYes: 2,
			OutgoingNo:  0,
		},
	}

	s.countersRepo.EXPECT().
		GetHourlyCounters(s.ctx, s.activeUserKey, hoursOffsetGroups).
		Return(expectedCounters, nil)

	operation := s.newOperation()
	counters, err := operation.Run(s.ctx, s.activeUserKey, hoursOffsetGroups)

	s.Require().NoError(err)
	s.Require().Equal(expectedCounters, counters)
}
