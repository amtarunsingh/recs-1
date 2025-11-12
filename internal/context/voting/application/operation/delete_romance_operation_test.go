package operation

import (
	"context"
	"errors"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/testlib/uuidhelper"
	"testing"

	sharedValueObject "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/sharedkernel/valueobject"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/testlib/mocks"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

type DeleteRomanceOperationUnitTestSuite struct {
	suite.Suite
	voteId       sharedValueObject.VoteId
	ctrl         *gomock.Controller
	romancesRepo *mocks.MockRomancesRepository
	ctx          context.Context
}

func TestDeleteRomanceOperationUnitSuite(t *testing.T) {
	suite.Run(t, new(DeleteRomanceOperationUnitTestSuite))
}

func (s *DeleteRomanceOperationUnitTestSuite) SetupSuite() {
	activeUserId := uuidhelper.NewUUID(s.T())
	peerUserId := uuidhelper.NewUUID(s.T())
	countryId := uint16(11)

	voteId, err := sharedValueObject.NewVoteId(countryId, activeUserId, peerUserId)
	s.Require().NoError(err)
	s.voteId = voteId
	s.ctx = context.Background()
}

func (s *DeleteRomanceOperationUnitTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.romancesRepo = mocks.NewMockRomancesRepository(s.ctrl)
}

func (s *DeleteRomanceOperationUnitTestSuite) newOperation() *DeleteRomanceOperation {
	return NewDeleteRomanceOperation(s.romancesRepo)
}

func (s *DeleteRomanceOperationUnitTestSuite) TestDeleteRomanceReturnsError() {
	expectedErr := errors.New("database error")

	s.romancesRepo.EXPECT().
		DeleteRomance(s.ctx, s.voteId).
		Return(expectedErr)

	operation := s.newOperation()
	err := operation.Run(s.ctx, s.voteId)

	s.Require().Error(err)
	s.Require().ErrorIs(err, expectedErr)
}

func (s *DeleteRomanceOperationUnitTestSuite) TestDeleteRomanceSuccessfully() {
	s.romancesRepo.EXPECT().
		DeleteRomance(s.ctx, s.voteId).
		Return(nil)

	operation := s.newOperation()
	err := operation.Run(s.ctx, s.voteId)

	s.Require().NoError(err)
}
