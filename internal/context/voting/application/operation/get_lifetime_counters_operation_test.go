package operation

import (
	"context"
	"errors"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/testlib/uuidhelper"
	"testing"

	counterEntity "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/counter/entity"
	sharedValueObject "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/sharedkernel/valueobject"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/testlib/mocks"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

type GetLifetimeCountersOperationUnitTestSuite struct {
	suite.Suite
	activeUserKey sharedValueObject.ActiveUserKey
	ctrl          *gomock.Controller
	countersRepo  *mocks.MockCountersRepository
	ctx           context.Context
}

func TestGetLifetimeCountersOperationUnitSuite(t *testing.T) {
	suite.Run(t, new(GetLifetimeCountersOperationUnitTestSuite))
}

func (s *GetLifetimeCountersOperationUnitTestSuite) SetupSuite() {
	activeUserId := uuidhelper.NewUUID(s.T())
	countryId := uint16(11)

	activeUserKey, err := sharedValueObject.NewActiveUserKey(countryId, activeUserId)
	s.Require().NoError(err)
	s.activeUserKey = activeUserKey
	s.ctx = context.Background()
}

func (s *GetLifetimeCountersOperationUnitTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.countersRepo = mocks.NewMockCountersRepository(s.ctrl)
}

func (s *GetLifetimeCountersOperationUnitTestSuite) newOperation() *GetLifetimeCountersOperation {
	return NewGetLifetimeCountersOperation(s.countersRepo)
}

func (s *GetLifetimeCountersOperationUnitTestSuite) TestGetLifetimeCountersReturnsError() {
	expectedErr := errors.New("database error")

	s.countersRepo.EXPECT().
		GetLifetimeCounter(s.ctx, s.activeUserKey).
		Return(counterEntity.CountersGroup{}, expectedErr)

	operation := s.newOperation()
	counters, err := operation.Run(s.ctx, s.activeUserKey)

	s.Require().Error(err)
	s.Require().ErrorIs(err, expectedErr)
	s.Require().Equal(counterEntity.CountersGroup{}, counters)
}

func (s *GetLifetimeCountersOperationUnitTestSuite) TestGetLifetimeCountersSuccessfully() {
	expectedCounters := counterEntity.CountersGroup{
		IncomingYes: 10,
		IncomingNo:  5,
		OutgoingYes: 8,
		OutgoingNo:  3,
	}

	s.countersRepo.EXPECT().
		GetLifetimeCounter(s.ctx, s.activeUserKey).
		Return(expectedCounters, nil)

	operation := s.newOperation()
	counters, err := operation.Run(s.ctx, s.activeUserKey)

	s.Require().NoError(err)
	s.Require().Equal(expectedCounters, counters)
}
