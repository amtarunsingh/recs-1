package operation

import (
	"context"
	"errors"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/testlib/uuidhelper"
	"io"
	"log/slog"
	"testing"

	sharedValueObject "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/sharedkernel/valueobject"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/testlib/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

type DeleteRomancesGroupOperationUnitTestSuite struct {
	suite.Suite
	activeUserKey sharedValueObject.ActiveUserKey
	ctrl          *gomock.Controller
	romancesRepo  *mocks.MockRomancesRepository
	logger        *slog.Logger
	ctx           context.Context
}

func TestDeleteRomancesGroupOperationUnitSuite(t *testing.T) {
	suite.Run(t, new(DeleteRomancesGroupOperationUnitTestSuite))
}

func (s *DeleteRomancesGroupOperationUnitTestSuite) SetupSuite() {
	activeUserId := uuidhelper.NewUUID(s.T())
	countryId := uint16(11)

	activeUserKey, err := sharedValueObject.NewActiveUserKey(countryId, activeUserId)
	s.Require().NoError(err)
	s.activeUserKey = activeUserKey
	s.ctx = context.Background()
	s.logger = slog.New(slog.NewTextHandler(io.Discard, nil))
}

func (s *DeleteRomancesGroupOperationUnitTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.romancesRepo = mocks.NewMockRomancesRepository(s.ctrl)
}

func (s *DeleteRomancesGroupOperationUnitTestSuite) newOperation() *DeleteRomancesGroupOperation {
	return NewDeleteRomancesGroupOperation(s.romancesRepo, s.logger)
}

func (s *DeleteRomancesGroupOperationUnitTestSuite) TestDeleteRomancesGroupReturnsError() {
	expectedErr := errors.New("database error")
	peerIds := []uuid.UUID{uuidhelper.NewUUID(s.T()), uuidhelper.NewUUID(s.T()), uuidhelper.NewUUID(s.T())}

	s.romancesRepo.EXPECT().
		DeleteRomancesGroup(s.ctx, s.activeUserKey, peerIds).
		Return(expectedErr)

	operation := s.newOperation()
	err := operation.Run(s.ctx, s.activeUserKey, peerIds)

	s.Require().Error(err)
	s.Require().ErrorIs(err, expectedErr)
}

func (s *DeleteRomancesGroupOperationUnitTestSuite) TestDeleteRomancesGroupSuccessfully() {
	peerIds := []uuid.UUID{uuidhelper.NewUUID(s.T()), uuidhelper.NewUUID(s.T()), uuidhelper.NewUUID(s.T())}

	s.romancesRepo.EXPECT().
		DeleteRomancesGroup(s.ctx, s.activeUserKey, peerIds).
		Return(nil)

	operation := s.newOperation()
	err := operation.Run(s.ctx, s.activeUserKey, peerIds)

	s.Require().NoError(err)
}
