package operation

import (
	"context"
	"errors"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/testlib/uuidhelper"
	"io"
	"log/slog"
	"testing"

	"github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/application/messaging/message"
	sharedValueObject "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/sharedkernel/valueobject"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/testlib/mocks"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

type DeleteRomancesRequestOperationUnitTestSuite struct {
	suite.Suite
	activeUserKey sharedValueObject.ActiveUserKey
	ctrl          *gomock.Controller
	publisher     *mocks.MockPublisher
	logger        *slog.Logger
	ctx           context.Context
}

func TestDeleteRomancesRequestOperationUnitSuite(t *testing.T) {
	suite.Run(t, new(DeleteRomancesRequestOperationUnitTestSuite))
}

func (s *DeleteRomancesRequestOperationUnitTestSuite) SetupSuite() {
	activeUserId := uuidhelper.NewUUID(s.T())
	countryId := uint16(11)

	activeUserKey, err := sharedValueObject.NewActiveUserKey(countryId, activeUserId)
	s.Require().NoError(err)
	s.activeUserKey = activeUserKey
	s.ctx = context.Background()
	s.logger = slog.New(slog.NewTextHandler(io.Discard, nil))
}

func (s *DeleteRomancesRequestOperationUnitTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.publisher = mocks.NewMockPublisher(s.ctrl)
}

func (s *DeleteRomancesRequestOperationUnitTestSuite) newOperation() *DeleteRomancesRequestOperation {
	return NewDeleteRomancesRequestOperation(s.publisher, s.logger)
}

func (s *DeleteRomancesRequestOperationUnitTestSuite) TestPublishReturnsError() {
	expectedErr := errors.New("publish error")
	expectedMessage := message.NewDeleteRomancesMessage(s.activeUserKey)

	s.publisher.EXPECT().
		Publish(DeleteRomancesTopic, expectedMessage).
		Return(expectedErr)

	operation := s.newOperation()
	err := operation.Run(s.ctx, s.activeUserKey)

	s.Require().Error(err)
	s.Require().ErrorIs(err, expectedErr)
}

func (s *DeleteRomancesRequestOperationUnitTestSuite) TestDeleteRomancesRequestSuccessfully() {
	expectedMessage := message.NewDeleteRomancesMessage(s.activeUserKey)

	s.publisher.EXPECT().
		Publish(DeleteRomancesTopic, expectedMessage).
		Return(nil)

	operation := s.newOperation()
	err := operation.Run(s.ctx, s.activeUserKey)

	s.Require().NoError(err)
}
