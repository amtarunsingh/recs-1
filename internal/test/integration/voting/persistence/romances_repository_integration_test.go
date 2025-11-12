package persistence

import (
	"context"
	"github.com/bmbl-bumble2/recs-votes-storage/config"
	romanceDomain "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/romance"
	romanceEntity "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/romance/entity"
	romanceRepository "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/romance/repository"
	rvo "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/romance/valueobject"
	sharedValueObject "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/domain/sharedkernel/valueobject"
	infraDynamodb "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/infrastructure/persistence"
	platformDynamodb "github.com/bmbl-bumble2/recs-votes-storage/internal/shared/platform/dynamodb"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/shared/timeutil"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/testlib"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/testlib/helper"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/testlib/uuidhelper"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"io"
	"log/slog"
	"testing"
	"time"
)

type RomancesRepositoryTestSuite struct {
	suite.Suite
	romancesTableHelper *helper.RomancesTableHelper
	voteId              sharedValueObject.VoteId
}

func TestMyRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(RomancesRepositoryTestSuite))
}

func (s *RomancesRepositoryTestSuite) SetupSuite() {
	romancesTableHelper, err := helper.NewRomancesTableHelper(ddbClient)
	s.Require().NoError(err)
	s.romancesTableHelper = romancesTableHelper

	err = s.romancesTableHelper.CreateRomancesTable()
	s.Require().NoError(err)

	activeUserId := uuidhelper.NewUUID(s.T())
	peerUserId := uuidhelper.NewUUID(s.T())
	countryId := uint16(11)

	voteId, err := sharedValueObject.NewVoteId(countryId, activeUserId, peerUserId)
	s.Require().NoError(err)
	s.voteId = voteId
}

func (s *RomancesRepositoryTestSuite) TearDownSuite() {
	repo := newRomancesRepository(ddbClient)
	err := repo.DeleteRomance(context.Background(), s.voteId)
	s.Require().NoError(err)
}

func (s *RomancesRepositoryTestSuite) SetupTest() {
	repo := newRomancesRepository(ddbClient)
	err := repo.DeleteRomance(context.Background(), s.voteId)
	s.Require().NoError(err)
}

func (s *RomancesRepositoryTestSuite) TestGetEmptyRomance() {
	repo := newRomancesRepository(ddbClient)
	romance, err := repo.GetRomance(context.Background(), s.voteId)
	s.Require().NoError(err)
	s.assertRomanceInDbMatchesExpected(
		newExpectedRomanceParams(s.voteId, rvo.VoteTypeEmpty, rvo.VoteTypeEmpty, 0),
		romance,
	)
}

func (s *RomancesRepositoryTestSuite) TestAddVoteToEmptyRomance() {
	ctx := context.Background()
	repo := newRomancesRepository(ddbClient)

	romance := romanceEntity.CreateEmptyRomance(s.voteId)

	// Adding a YES vote for the active user
	newRomance, err := repo.AddActiveUserVoteToRomance(ctx, romance, rvo.VoteTypeYes, time.Now())
	s.Require().NoError(err)
	s.assertRomanceInDbMatchesExpected(
		newExpectedRomanceParams(s.voteId, rvo.VoteTypeYes, rvo.VoteTypeEmpty, 1),
		newRomance,
	)
}

func (s *RomancesRepositoryTestSuite) TestGetAndAddRomanceVote() {
	ctx := context.Background()
	repo := newRomancesRepository(ddbClient)

	// step 1: Receiving an empty romance object
	romance, err := repo.GetRomance(ctx, s.voteId)
	s.Require().NoError(err)
	s.assertRomanceInDbMatchesExpected(
		newExpectedRomanceParams(s.voteId, rvo.VoteTypeEmpty, rvo.VoteTypeEmpty, 0),
		romance,
	)

	// step 2: Adding a YES vote for the active user
	newRomance, err := repo.AddActiveUserVoteToRomance(ctx, romance, rvo.VoteTypeYes, time.Now())
	s.Require().NoError(err)
	s.assertRomanceInDbMatchesExpected(
		newExpectedRomanceParams(s.voteId, rvo.VoteTypeYes, rvo.VoteTypeEmpty, 1),
		newRomance,
	)

	// step 3: Getting a romance from the repo (active user side) and checking if it contains the expected values
	romance, err = repo.GetRomance(ctx, s.voteId)
	s.Require().NoError(err)
	assert.Equal(s.T(), newRomance, romance)

	// step 4: Getting a romance from the repo (peer user side) and checking if it contains the expected values
	peerVoteId := s.voteId.ToPeerVoteId()
	peerRomance, err := repo.GetRomance(ctx, peerVoteId)
	s.Require().NoError(err)
	s.assertRomanceInDbMatchesExpected(
		newExpectedRomanceParams(peerVoteId, rvo.VoteTypeEmpty, rvo.VoteTypeYes, 1),
		peerRomance,
	)

	// step 5: Adding a new NO vote from the peer side
	newPeerRomance, err := repo.AddActiveUserVoteToRomance(ctx, peerRomance, rvo.VoteTypeNo, time.Now())
	s.Require().NoError(err)
	s.assertRomanceInDbMatchesExpected(
		newExpectedRomanceParams(peerVoteId, rvo.VoteTypeNo, rvo.VoteTypeYes, 2),
		newPeerRomance,
	)

	// step 6: Comparison of the romance received after the update and from GetRomance method
	peerRomance, err = repo.GetRomance(ctx, s.voteId.ToPeerVoteId())
	s.Require().NoError(err)
	assert.Equal(s.T(), newPeerRomance, peerRomance)
}

func (s *RomancesRepositoryTestSuite) TestAddVoteAndRewrite() {
	ctx := context.Background()
	repo := newRomancesRepository(ddbClient)

	now := time.Now()

	activeUserVoteId := s.voteId
	peerVoteId := s.voteId.ToPeerVoteId()

	// step 1: Adding a NO vote for the peer user
	peerRomance := romanceEntity.CreateEmptyRomance(peerVoteId)
	newPeerRomance, err := repo.AddActiveUserVoteToRomance(ctx, peerRomance, rvo.VoteTypeYes, now)
	s.Require().NoError(err)
	s.assertRomanceInDbMatchesExpected(
		newExpectedRomanceParams(peerVoteId, rvo.VoteTypeYes, rvo.VoteTypeEmpty, 1),
		newPeerRomance,
	)

	// step 2: Rewriting peer vote
	// There is no check at the persistence level for which vote we are inserting,
	// so there may be a NO after YES
	newPeerRomance, err = repo.AddActiveUserVoteToRomance(ctx, newPeerRomance, rvo.VoteTypeNo, now)
	s.Require().NoError(err)
	s.assertRomanceInDbMatchesExpected(
		newExpectedRomanceParams(peerVoteId, rvo.VoteTypeNo, rvo.VoteTypeEmpty, 2),
		newPeerRomance,
	)

	// step 3: Get the same romance from the active user side
	romance, err := repo.GetRomance(ctx, activeUserVoteId)
	s.Require().NoError(err)
	s.assertRomanceInDbMatchesExpected(
		newExpectedRomanceParams(activeUserVoteId, rvo.VoteTypeEmpty, rvo.VoteTypeNo, 2),
		romance,
	)

	// step 4: Adding a YES vote to the active user romance
	newRomance, err := repo.AddActiveUserVoteToRomance(ctx, romance, rvo.VoteTypeYes, now)
	s.Require().NoError(err)
	s.assertRomanceInDbMatchesExpected(
		newExpectedRomanceParams(activeUserVoteId, rvo.VoteTypeYes, rvo.VoteTypeNo, 3),
		newRomance,
	)

	// step 5: Adding a NO vote for the active user
	newRomance, err = repo.AddActiveUserVoteToRomance(ctx, newRomance, rvo.VoteTypeNo, now)
	s.Require().NoError(err)
	s.assertRomanceInDbMatchesExpected(
		newExpectedRomanceParams(activeUserVoteId, rvo.VoteTypeNo, rvo.VoteTypeNo, 4),
		newRomance,
	)
}

func (s *RomancesRepositoryTestSuite) TestAddWithVersionConflictError() {
	ctx := context.Background()
	repo := newRomancesRepository(ddbClient)

	now := time.Now()

	// step 1: Creating romance with wrong version (the version is not synchronized with the db)
	romance := romanceEntity.CreateEmptyRomance(s.voteId)
	romance.Version = 10

	// step 2: Adding vote
	newRomance, err := repo.AddActiveUserVoteToRomance(ctx, romance, rvo.VoteTypeYes, now)
	s.Require().Error(err)
	s.Require().ErrorIs(err, romanceDomain.ErrVersionConflict)
	s.assertNilRomance(newRomance)
}

func (s *RomancesRepositoryTestSuite) TestAddActiveUserVoteWithWrongPeerVotePart() {
	ctx := context.Background()
	repo := newRomancesRepository(ddbClient)

	now := time.Now()

	// step 1: Creating romance with wrong peer vote part (the peer vote part is not synchronized with the db)
	romance := romanceEntity.CreateEmptyRomance(s.voteId)
	romance.PeerUserVote.VoteType = rvo.VoteTypeNo
	romance.PeerUserVote.VotedAt = &now

	// step 2: Adding vote
	newRomance, err := repo.AddActiveUserVoteToRomance(ctx, romance, rvo.VoteTypeYes, now)
	s.Require().NoError(err)
	// The `AddActiveUserVoteToRomance` method returns a synchronized romance
	s.assertRomanceInDbMatchesExpected(
		newExpectedRomanceParams(s.voteId, rvo.VoteTypeYes, rvo.VoteTypeEmpty, 1),
		newRomance,
	)
}

func (s *RomancesRepositoryTestSuite) TestDeleteRomance() {
	ctx := context.Background()
	repo := newRomancesRepository(ddbClient)

	// step 1: Adding new Romance
	romance := romanceEntity.CreateEmptyRomance(s.voteId)
	_, err := repo.AddActiveUserVoteToRomance(ctx, romance, rvo.VoteTypeYes, time.Now())
	s.Require().NoError(err)

	// step 2: Deleting this romance
	err = repo.DeleteRomance(ctx, s.voteId)
	s.Require().NoError(err)

	// step 3: Check romance from active user and peer side
	romance, err = repo.GetRomance(ctx, s.voteId)
	s.Require().NoError(err)
	s.assertEmptyRomance(s.voteId, romance)

	romance, err = repo.GetRomance(ctx, s.voteId.ToPeerVoteId())
	s.Require().NoError(err)
	s.assertEmptyRomance(s.voteId.ToPeerVoteId(), romance)
}

func (s *RomancesRepositoryTestSuite) TestDeleteRomanceFromPeerSide() {
	ctx := context.Background()
	repo := newRomancesRepository(ddbClient)

	// step 1: Adding new Romance
	romance := romanceEntity.CreateEmptyRomance(s.voteId)
	_, err := repo.AddActiveUserVoteToRomance(ctx, romance, rvo.VoteTypeYes, time.Now())
	s.Require().NoError(err)

	// step 2: Deleting this romance from peer side
	err = repo.DeleteRomance(ctx, s.voteId.ToPeerVoteId())
	s.Require().NoError(err)

	// step 3: Check romance from active user and peer side
	romance, err = repo.GetRomance(ctx, s.voteId)
	s.Require().NoError(err)
	s.assertEmptyRomance(s.voteId, romance)

	romance, err = repo.GetRomance(ctx, s.voteId.ToPeerVoteId())
	s.Require().NoError(err)
	s.assertEmptyRomance(s.voteId.ToPeerVoteId(), romance)
}

func (s *RomancesRepositoryTestSuite) TestDeleteNotExistsRomance() {
	ctx := context.Background()
	repo := newRomancesRepository(ddbClient)

	err := repo.DeleteRomance(ctx, s.voteId)
	s.Require().NoError(err)
}

func (s *RomancesRepositoryTestSuite) TestDeleteActiveUserVoteFromRomance() {
	ctx := context.Background()
	repo := newRomancesRepository(ddbClient)

	activeUserVoteId := s.voteId
	peerUserVoteId := s.voteId.ToPeerVoteId()

	// step 1: Adding new Romance and active user vote
	emptyActiveUserRomance := romanceEntity.CreateEmptyRomance(activeUserVoteId)
	_, err := repo.AddActiveUserVoteToRomance(ctx, emptyActiveUserRomance, rvo.VoteTypeYes, time.Now())
	s.Require().NoError(err)

	// step 2: Add vote to peer user side
	peerRomance, err := repo.GetRomance(ctx, peerUserVoteId)
	s.Require().NoError(err)
	peerRomance, err = repo.AddActiveUserVoteToRomance(ctx, peerRomance, rvo.VoteTypeNo, time.Now())
	s.Require().NoError(err)

	s.assertRomanceInDbMatchesExpected(
		newExpectedRomanceParams(peerUserVoteId, rvo.VoteTypeNo, rvo.VoteTypeYes, 2),
		peerRomance,
	)

	// step 3: Deleting vote from peer user side
	err = repo.DeleteActiveUserVoteFromRomance(ctx, peerRomance)
	s.Require().NoError(err)

	peerRomance, err = repo.GetRomance(ctx, peerUserVoteId)
	s.Require().NoError(err)
	s.assertRomanceInDbMatchesExpected(
		newExpectedRomanceParams(peerUserVoteId, rvo.VoteTypeEmpty, rvo.VoteTypeYes, 3),
		peerRomance,
	)

	// step 4: Deleting vote from active user side
	activeUserRomance, err := repo.GetRomance(ctx, activeUserVoteId)
	s.Require().NoError(err)

	err = repo.DeleteActiveUserVoteFromRomance(ctx, activeUserRomance)
	s.Require().NoError(err)

	activeUserRomance, err = repo.GetRomance(ctx, activeUserVoteId)
	s.Require().NoError(err)
	s.assertRomanceInDbMatchesExpected(
		newExpectedRomanceParams(activeUserVoteId, rvo.VoteTypeEmpty, rvo.VoteTypeEmpty, 4),
		activeUserRomance,
	)
}

func (s *RomancesRepositoryTestSuite) TestDeleteActiveUserVoteFromEmptyRomance() {
	ctx := context.Background()
	repo := newRomancesRepository(ddbClient)

	// step 1: Adding new Romance and active user vote
	emptyActiveUserRomance := romanceEntity.CreateEmptyRomance(s.voteId)

	err := repo.DeleteActiveUserVoteFromRomance(ctx, emptyActiveUserRomance)
	s.Require().NoError(err)

	activeUserRomance, err := repo.GetRomance(ctx, s.voteId)
	s.Require().NoError(err)
	s.assertRomanceInDbMatchesExpected(
		newExpectedRomanceParams(s.voteId, rvo.VoteTypeEmpty, rvo.VoteTypeEmpty, 0),
		activeUserRomance,
	)
}

func (s *RomancesRepositoryTestSuite) TestChangeActiveUserVoteTypeInRomance() {
	ctx := context.Background()
	repo := newRomancesRepository(ddbClient)

	activeUserVoteId := s.voteId
	peerUserVoteId := s.voteId.ToPeerVoteId()

	// step 1: Adding new Romance
	romance := romanceEntity.CreateEmptyRomance(activeUserVoteId)
	romance, err := repo.AddActiveUserVoteToRomance(ctx, romance, rvo.VoteTypeNo, time.Now())
	s.Require().NoError(err)

	// step 2: Changing active user vote type
	newRomance, err := repo.ChangeActiveUserVoteTypeInRomance(ctx, romance, rvo.VoteTypeYes)
	s.Require().NoError(err)
	s.assertRomanceInDbMatchesExpected(
		newExpectedRomanceParams(activeUserVoteId, rvo.VoteTypeYes, rvo.VoteTypeEmpty, 2),
		newRomance,
	)

	// step 3: Adding peer vote
	peerRomance, err := repo.GetRomance(ctx, peerUserVoteId)
	s.Require().NoError(err)
	peerRomance, err = repo.AddActiveUserVoteToRomance(ctx, peerRomance, rvo.VoteTypeYes, time.Now())
	s.Require().NoError(err)

	// step 4: Changing peer user vote type
	newPeerRomance, err := repo.ChangeActiveUserVoteTypeInRomance(ctx, peerRomance, rvo.VoteTypeCrush)
	s.Require().NoError(err)
	s.assertRomanceInDbMatchesExpected(
		newExpectedRomanceParams(peerUserVoteId, rvo.VoteTypeCrush, rvo.VoteTypeYes, 4),
		newPeerRomance,
	)
}

func (s *RomancesRepositoryTestSuite) TestChangeActiveUserVoteTypeInEmptyRomance() {
	ctx := context.Background()
	repo := newRomancesRepository(ddbClient)

	romance := romanceEntity.CreateEmptyRomance(s.voteId)
	newRomance, err := repo.ChangeActiveUserVoteTypeInRomance(ctx, romance, rvo.VoteTypeNo)
	s.Require().Error(err)
	s.Require().ErrorIs(err, romanceDomain.ErrVoteNotFound)
	s.assertNilRomance(newRomance)
}

func (s *RomancesRepositoryTestSuite) TestChangeActiveUserVoteTypeInRomanceVariants() {
	cases := []struct {
		fromVote rvo.VoteType
		toVote   rvo.VoteType
	}{
		// there is no vote type check here (only for empty votes)
		{rvo.VoteTypeCompliment, rvo.VoteTypeYes},
		{rvo.VoteTypeCompliment, rvo.VoteTypeCrush},
		{rvo.VoteTypeCrush, rvo.VoteTypeCompliment},
		{rvo.VoteTypeCrush, rvo.VoteTypeNo},
		{rvo.VoteTypeCompliment, rvo.VoteTypeNo},
		{rvo.VoteTypeYes, rvo.VoteTypeNo},
		{rvo.VoteTypeNo, rvo.VoteTypeYes},
	}

	ctx := context.Background()
	repo := newRomancesRepository(ddbClient)

	for _, c := range cases {
		activeUserVoteId := s.voteId

		// step 1: Adding fresh Romance
		err := repo.DeleteRomance(ctx, activeUserVoteId)
		s.Require().NoError(err)
		romance := romanceEntity.CreateEmptyRomance(activeUserVoteId)
		romance, err = repo.AddActiveUserVoteToRomance(ctx, romance, c.fromVote, time.Now())
		s.Require().NoError(err)

		// step 2: Changing active user vote type
		newRomance, err := repo.ChangeActiveUserVoteTypeInRomance(ctx, romance, c.toVote)
		s.Require().NoError(err)
		s.assertRomanceInDbMatchesExpected(
			newExpectedRomanceParams(activeUserVoteId, c.toVote, rvo.VoteTypeEmpty, 2),
			newRomance,
		)
	}
}

func (s *RomancesRepositoryTestSuite) TestChangeActiveUserVoteTypeInRomanceToEmpty() {
	ctx := context.Background()
	repo := newRomancesRepository(ddbClient)

	activeUserVoteId := s.voteId

	// step 1: Adding fresh Romance
	err := repo.DeleteRomance(ctx, activeUserVoteId)
	s.Require().NoError(err)
	romance := romanceEntity.CreateEmptyRomance(activeUserVoteId)
	romance, err = repo.AddActiveUserVoteToRomance(ctx, romance, rvo.VoteTypeYes, time.Now())
	s.Require().NoError(err)

	// step 2: Changing active user vote type to empty
	newRomance, err := repo.ChangeActiveUserVoteTypeInRomance(ctx, romance, rvo.VoteTypeEmpty)
	s.Require().Error(err)
	s.Require().ErrorIs(err, romanceDomain.ErrWrongVote)
	s.assertNilRomance(newRomance)
}

func (s *RomancesRepositoryTestSuite) assertNilRomance(romanceToCheck romanceEntity.Romance) {
	assert.Empty(s.T(), romanceToCheck)
}

func (s *RomancesRepositoryTestSuite) assertEmptyRomance(
	voteId sharedValueObject.VoteId,
	romanceToCheck romanceEntity.Romance,
) {
	expected := map[string]any{
		"activeUserVoteActiveUserId": voteId.ActiveUserId(),
		"activeUserVotePeerUserId":   voteId.PeerUserId(),
		"peerUserVoteActiveUserId":   voteId.PeerUserId(),
		"peerUserVotePeerUserId":     voteId.ActiveUserId(),
		"activeUserVoteType":         rvo.VoteTypeEmpty,
		"activeUserVotedAt":          nil,
		"activeUserCreatedAt":        nil,
		"activeUserUpdatedAt":        nil,
		"peerUserVoteType":           rvo.VoteTypeEmpty,
		"peerUserVotedAt":            nil,
		"peerUserCreatedAt":          nil,
		"peerUserUpdatedAt":          nil,
		"version":                    0,
	}

	activeUserVote := romanceToCheck.ActiveUserVote
	peerUserVote := romanceToCheck.PeerUserVote

	actual := map[string]any{
		"activeUserVoteActiveUserId": activeUserVote.Id.ActiveUserId(),
		"activeUserVotePeerUserId":   activeUserVote.Id.PeerUserId(),
		"peerUserVoteActiveUserId":   peerUserVote.Id.ActiveUserId(),
		"peerUserVotePeerUserId":     peerUserVote.Id.PeerUserId(),
		"activeUserVoteType":         activeUserVote.VoteType,
		"activeUserVotedAt":          activeUserVote.VotedAt,
		"activeUserCreatedAt":        activeUserVote.CreatedAt,
		"activeUserUpdatedAt":        activeUserVote.UpdatedAt,
		"peerUserVoteType":           peerUserVote.VoteType,
		"peerUserVotedAt":            peerUserVote.VotedAt,
		"peerUserCreatedAt":          peerUserVote.CreatedAt,
		"peerUserUpdatedAt":          peerUserVote.UpdatedAt,
		"version":                    romanceToCheck.Version,
	}

	testlib.AssertMap(s.T(), expected, actual)
}

func (s *RomancesRepositoryTestSuite) assertRomanceInDbMatchesExpected(
	expectedParams expectedRomanceParams,
	romanceToCheck romanceEntity.Romance,
) {

	expected := map[string]any{
		"activeUserVoteActiveUserId": expectedParams.ActiveUserId,
		"activeUserVotePeerUserId":   expectedParams.PeerUserId,
		"peerUserVoteActiveUserId":   expectedParams.PeerUserId,
		"peerUserVotePeerUserId":     expectedParams.ActiveUserId,
		"activeUserVoteType":         expectedParams.ActiveUserVoteType,
		"peerUserVoteType":           expectedParams.PeerUserVoteType,
		"version":                    expectedParams.Version,
	}

	activeUserVote := romanceToCheck.ActiveUserVote
	peerUserVote := romanceToCheck.PeerUserVote

	actual := map[string]any{
		"activeUserVoteActiveUserId": activeUserVote.Id.ActiveUserId(),
		"activeUserVotePeerUserId":   activeUserVote.Id.PeerUserId(),
		"peerUserVoteActiveUserId":   peerUserVote.Id.ActiveUserId(),
		"peerUserVotePeerUserId":     peerUserVote.Id.PeerUserId(),
		"activeUserVoteType":         activeUserVote.VoteType,
		"peerUserVoteType":           peerUserVote.VoteType,
		"version":                    romanceToCheck.Version,
	}

	testlib.AssertMap(s.T(), expected, actual)

	voteId := romanceToCheck.ActiveUserVote.Id
	romanceKey := infraDynamodb.NewRomancePrimaryKey(voteId)
	dynamodbRegion := platformDynamodb.GetDynamodbRegionByCountry(voteId.CountryId())
	record, err := s.romancesTableHelper.GetRomanceTableRecord(romanceKey, dynamodbRegion)
	s.Require().NoError(err)
	assertRomanceDbRecord(s.T(), record, romanceToCheck)
}

func newRomancesRepository(client platformDynamodb.Client) romanceRepository.RomancesRepository {
	appConfig := config.Load()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return infraDynamodb.NewRomancesRepository(client, appConfig, logger)
}

func assertRomanceDbRecord(
	t *testing.T,
	record infraDynamodb.RomanceDocumentSchema,
	romance romanceEntity.Romance,
) {
	activeUserVote := romance.ActiveUserVote
	peerUserVote := romance.PeerUserVote

	var (
		pkUserVote romanceEntity.Vote
		skUserVote romanceEntity.Vote
	)

	if record.PkUserId == activeUserVote.Id.ActiveUserId().String() {
		pkUserVote = activeUserVote
		skUserVote = peerUserVote
	} else {
		pkUserVote = peerUserVote
		skUserVote = activeUserVote
	}

	expected := map[string]any{
		"pkUserVoteType":      record.PkUserVoteType,
		"pkUserVotedAt":       timeutil.UnixToTimePtr(record.PkUserVotedAt),
		"pkUserVoteCreatedAt": timeutil.UnixToTimePtr(record.PkUserVoteCreatedAt),
		"pkUserVoteUpdatedAt": timeutil.UnixToTimePtr(record.PkUserVoteUpdatedAt),
		"skUserId":            record.SkUserId,
		"skUserVoteType":      record.SkUserVoteType,
		"skUserVotedAt":       timeutil.UnixToTimePtr(record.SkUserVotedAt),
		"skUserVoteCreatedAt": timeutil.UnixToTimePtr(record.SkUserVoteCreatedAt),
		"skUserVoteUpdatedAt": timeutil.UnixToTimePtr(record.SkUserVoteUpdatedAt),
		"version":             record.Version,
	}

	actual := map[string]any{
		"pkUserVoteType":      pkUserVote.VoteType,
		"pkUserVotedAt":       pkUserVote.VotedAt,
		"pkUserVoteCreatedAt": pkUserVote.CreatedAt,
		"pkUserVoteUpdatedAt": pkUserVote.UpdatedAt,
		"skUserId":            skUserVote.Id.ActiveUserId().String(),
		"skUserVoteType":      skUserVote.VoteType,
		"skUserVotedAt":       skUserVote.VotedAt,
		"skUserVoteCreatedAt": skUserVote.CreatedAt,
		"skUserVoteUpdatedAt": skUserVote.UpdatedAt,
		"version":             romance.Version,
	}

	if romance.Version == 0 {
		actual["skUserId"] = ""
	}

	testlib.AssertMap(t, expected, actual)
}

type expectedRomanceParams struct {
	ActiveUserId       uuid.UUID
	PeerUserId         uuid.UUID
	ActiveUserVoteType rvo.VoteType
	PeerUserVoteType   rvo.VoteType
	Version            uint32
}

func newExpectedRomanceParams(
	voteId sharedValueObject.VoteId,
	activeUserVoteType rvo.VoteType,
	peerUserVoteType rvo.VoteType,
	version uint32,
) expectedRomanceParams {
	return expectedRomanceParams{
		ActiveUserId:       voteId.ActiveUserId(),
		PeerUserId:         voteId.PeerUserId(),
		ActiveUserVoteType: activeUserVoteType,
		PeerUserVoteType:   peerUserVoteType,
		Version:            version,
	}
}
