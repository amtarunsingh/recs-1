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
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

type DeleteRomancesOperationUnitTestSuite struct {
	suite.Suite
	activeUserKey sharedValueObject.ActiveUserKey
	ctrl          *gomock.Controller
	romancesRepo  *mocks.MockRomancesRepository
	publisher     *mocks.MockPublisher
	logger        *slog.Logger
	ctx           context.Context
}

func TestDeleteRomancesOperationUnitSuite(t *testing.T) {
	suite.Run(t, new(DeleteRomancesOperationUnitTestSuite))
}

func (s *DeleteRomancesOperationUnitTestSuite) SetupSuite() {
	activeUserId := uuidhelper.NewUUID(s.T())
	countryId := uint16(11)

	activeUserKey, err := sharedValueObject.NewActiveUserKey(countryId, activeUserId)
	s.Require().NoError(err)
	s.activeUserKey = activeUserKey
	s.ctx = context.Background()
	s.logger = slog.New(slog.NewTextHandler(io.Discard, nil))
}

func (s *DeleteRomancesOperationUnitTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.romancesRepo = mocks.NewMockRomancesRepository(s.ctrl)
	s.publisher = mocks.NewMockPublisher(s.ctrl)
}

func (s *DeleteRomancesOperationUnitTestSuite) newOperation() *DeleteRomancesOperation {
	return NewDeleteRomancesOperation(s.romancesRepo, s.publisher, s.logger)
}

func (s *DeleteRomancesOperationUnitTestSuite) TestGetAllPeersForActiveUserReturnsError() {
	expectedErr := errors.New("database error")

	s.romancesRepo.EXPECT().
		GetAllPeersForActiveUser(s.ctx, s.activeUserKey).
		Return(nil, expectedErr)

	operation := s.newOperation()
	err := operation.Run(s.ctx, s.activeUserKey)

	s.Require().Error(err)
	s.Require().ErrorIs(err, expectedErr)
}

func (s *DeleteRomancesOperationUnitTestSuite) TestPublishReturnsErrorOnFirstBatch() {
	// Create a channel with a full batch of peer IDs
	peerIdsChan := make(chan uuid.UUID, getRomancesGroupLimit)
	peerIds := make([]uuid.UUID, getRomancesGroupLimit)
	for i := 0; i < getRomancesGroupLimit; i++ {
		peerIds[i] = uuidhelper.NewUUID(s.T())
		peerIdsChan <- peerIds[i]
	}
	close(peerIdsChan)

	expectedErr := errors.New("publish error")
	expectedMessage := message.NewDeleteRomancesGroupMessage(s.activeUserKey, peerIds)

	s.romancesRepo.EXPECT().
		GetAllPeersForActiveUser(s.ctx, s.activeUserKey).
		Return(peerIdsChan, nil)

	s.publisher.EXPECT().
		Publish(DeleteRomancesGroupTopic, expectedMessage).
		Return(expectedErr)

	operation := s.newOperation()
	err := operation.Run(s.ctx, s.activeUserKey)

	s.Require().Error(err)
	s.Require().ErrorIs(err, expectedErr)
}

func (s *DeleteRomancesOperationUnitTestSuite) TestPublishReturnsErrorOnRemainder() {
	// Create a channel with less than a full batch of peer IDs
	remainderSize := getRomancesGroupLimit - 10
	peerIdsChan := make(chan uuid.UUID, remainderSize)
	peerIds := make([]uuid.UUID, remainderSize)
	for i := 0; i < remainderSize; i++ {
		peerIds[i] = uuidhelper.NewUUID(s.T())
		peerIdsChan <- peerIds[i]
	}
	close(peerIdsChan)

	expectedErr := errors.New("publish error")
	expectedMessage := message.NewDeleteRomancesGroupMessage(s.activeUserKey, peerIds)

	s.romancesRepo.EXPECT().
		GetAllPeersForActiveUser(s.ctx, s.activeUserKey).
		Return(peerIdsChan, nil)

	s.publisher.EXPECT().
		Publish(DeleteRomancesGroupTopic, expectedMessage).
		Return(expectedErr)

	operation := s.newOperation()
	err := operation.Run(s.ctx, s.activeUserKey)

	s.Require().Error(err)
	s.Require().ErrorIs(err, expectedErr)
}

func (s *DeleteRomancesOperationUnitTestSuite) TestDeleteRomancesWithExactlyOneBatchSuccessfully() {
	// Create a channel with a full batch of peer IDs
	peerIdsChan := make(chan uuid.UUID, getRomancesGroupLimit)
	peerIds := make([]uuid.UUID, getRomancesGroupLimit)
	for i := 0; i < getRomancesGroupLimit; i++ {
		peerIds[i] = uuidhelper.NewUUID(s.T())
		peerIdsChan <- peerIds[i]
	}
	close(peerIdsChan)

	expectedMessage := message.NewDeleteRomancesGroupMessage(s.activeUserKey, peerIds)

	s.romancesRepo.EXPECT().
		GetAllPeersForActiveUser(s.ctx, s.activeUserKey).
		Return(peerIdsChan, nil)

	s.publisher.EXPECT().
		Publish(DeleteRomancesGroupTopic, expectedMessage).
		Return(nil)

	operation := s.newOperation()
	err := operation.Run(s.ctx, s.activeUserKey)

	s.Require().NoError(err)
}

func (s *DeleteRomancesOperationUnitTestSuite) TestDeleteRomancesWithRemainderSuccessfully() {
	// Create a channel with more than a full batch of peer IDs (one full batch + 5 remainder)
	peerIdsChan := make(chan uuid.UUID, getRomancesGroupLimit+5)
	firstBatch := make([]uuid.UUID, getRomancesGroupLimit)
	remainder := make([]uuid.UUID, 5)

	for i := 0; i < getRomancesGroupLimit; i++ {
		firstBatch[i] = uuidhelper.NewUUID(s.T())
		peerIdsChan <- firstBatch[i]
	}
	for i := 0; i < 5; i++ {
		remainder[i] = uuidhelper.NewUUID(s.T())
		peerIdsChan <- remainder[i]
	}
	close(peerIdsChan)

	firstMessage := message.NewDeleteRomancesGroupMessage(s.activeUserKey, firstBatch)
	remainderMessage := message.NewDeleteRomancesGroupMessage(s.activeUserKey, remainder)

	s.romancesRepo.EXPECT().
		GetAllPeersForActiveUser(s.ctx, s.activeUserKey).
		Return(peerIdsChan, nil)

	s.publisher.EXPECT().
		Publish(DeleteRomancesGroupTopic, firstMessage).
		Return(nil)

	s.publisher.EXPECT().
		Publish(DeleteRomancesGroupTopic, remainderMessage).
		Return(nil)

	operation := s.newOperation()
	err := operation.Run(s.ctx, s.activeUserKey)

	s.Require().NoError(err)
}

func (s *DeleteRomancesOperationUnitTestSuite) TestDeleteRomancesWithMultipleBatchesSuccessfully() {
	// Create a channel with two batches of peer IDs
	peerIdsChan := make(chan uuid.UUID, getRomancesGroupLimit*2)
	firstBatch := make([]uuid.UUID, getRomancesGroupLimit)
	secondBatch := make([]uuid.UUID, getRomancesGroupLimit)

	for i := 0; i < getRomancesGroupLimit; i++ {
		firstBatch[i] = uuidhelper.NewUUID(s.T())
		peerIdsChan <- firstBatch[i]
	}
	for i := 0; i < getRomancesGroupLimit; i++ {
		secondBatch[i] = uuidhelper.NewUUID(s.T())
		peerIdsChan <- secondBatch[i]
	}
	close(peerIdsChan)

	firstMessage := message.NewDeleteRomancesGroupMessage(s.activeUserKey, firstBatch)
	secondMessage := message.NewDeleteRomancesGroupMessage(s.activeUserKey, secondBatch)

	s.romancesRepo.EXPECT().
		GetAllPeersForActiveUser(s.ctx, s.activeUserKey).
		Return(peerIdsChan, nil)

	s.publisher.EXPECT().
		Publish(DeleteRomancesGroupTopic, firstMessage).
		Return(nil)

	s.publisher.EXPECT().
		Publish(DeleteRomancesGroupTopic, secondMessage).
		Return(nil)

	operation := s.newOperation()
	err := operation.Run(s.ctx, s.activeUserKey)

	s.Require().NoError(err)
}

func (s *DeleteRomancesOperationUnitTestSuite) TestDeleteRomancesWithNoPeersSuccessfully() {
	// Create an empty channel
	peerIdsChan := make(chan uuid.UUID)
	close(peerIdsChan)

	s.romancesRepo.EXPECT().
		GetAllPeersForActiveUser(s.ctx, s.activeUserKey).
		Return(peerIdsChan, nil)

	// No publish should be called

	operation := s.newOperation()
	err := operation.Run(s.ctx, s.activeUserKey)

	s.Require().NoError(err)
}
