package v1

import (
	"context"
	apiResponse "github.com/bmbl-bumble2/recs-votes-storage/internal/app/api/response"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/application"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/interface/api/rest/v1/command"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/interface/api/rest/v1/query"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/interface/api/rest/v1/response"
	"github.com/danielgtaylor/huma/v2"
	"net/http"
)

type VotesStorageRoutesRegister struct {
	votesService *application.VotingService
}

func NewVotesStorageRoutesRegister(
	votesService *application.VotingService,
) VotesStorageRoutesRegister {
	return VotesStorageRoutesRegister{
		votesService: votesService,
	}
}

func (v VotesStorageRoutesRegister) RegisterV1Routes(grp *huma.Group) {
	registerRomancesRoutes(grp, v.votesService)
	registerVotesRoutes(grp, v.votesService)
	registerCountersRoutes(grp, v.votesService)
}

func registerRomancesRoutes(
	grp *huma.Group,
	votesService *application.VotingService,
) {
	grp = huma.NewGroup(grp, "/romances")
	grp.UseSimpleModifier(func(op *huma.Operation) {
		op.Tags = []string{"Romances"}
	})

	// GET /v1/romances/{country_id}/{active_user_id}/{peer_id}
	huma.Register(grp, huma.Operation{
		OperationID: "get-romance",
		Method:      http.MethodGet,
		Path:        "/{country_id}/{active_user_id}/{peer_id}",
		Summary:     "Get romance from the active user's perspective",
		Description: "Each user in a pair can take the role of either the active user or the peer, " +
			"and the order of users in the request determines how the romance object " +
			"will be constructed and returned.",
	}, func(reqCtx context.Context, get *query.RomanceGet) (*response.RomanceGetResponse, error) {
		romance, err := votesService.GetRomance(reqCtx, *get)
		if err != nil {
			return nil, response.ToApiError(err)
		}
		resp := response.CreateRomanceGetResponseFromVoteEntity(romance)
		return resp, nil
	})

	// DELETE /v1/romances/{country_id}/{active_user_id}/{peer_id}
	huma.Register(grp, huma.Operation{
		OperationID: "delete-romance",
		Method:      http.MethodDelete,
		Path:        "/{country_id}/{active_user_id}/{peer_id}",
		Summary:     "Delete romance",
	}, func(reqCtx context.Context, command *command.DeleteRomance) (*struct{}, error) {
		err := votesService.DeleteRomance(reqCtx, *command)
		if err != nil {
			return nil, response.ToApiError(err)
		}
		return nil, nil
	})

	// DELETE /v1/romances/{country_id}/{active_user_id}
	huma.Register(grp, huma.Operation{
		OperationID: "delete-romances",
		Method:      http.MethodDelete,
		Path:        "/{country_id}/{active_user_id}",
		Summary:     "Delete all active user romances",
	}, func(reqCtx context.Context, command *command.DeleteRomances) (*struct{}, error) {
		err := votesService.DeleteRomancesRequest(reqCtx, *command)
		if err != nil {
			return nil, response.ToApiError(err)
		}
		return nil, nil
	})
}

func registerVotesRoutes(
	grp *huma.Group,
	votesService *application.VotingService,
) {
	grp = huma.NewGroup(grp, "/votes")
	grp.UseSimpleModifier(func(op *huma.Operation) {
		op.Tags = []string{"Votes"}
	})

	// GET /v1/votes/{country_id}/{active_user_id}/{peer_id}
	huma.Register(grp, huma.Operation{
		OperationID: "get-vote",
		Method:      http.MethodGet,
		Path:        "/{country_id}/{active_user_id}/{peer_id}",
		Summary:     "Get vote from the active user's perspective",
	}, func(reqCtx context.Context, get *query.VoteGet) (*response.VoteGetResponse, error) {
		vote, err := votesService.GetUserVote(reqCtx, *get)
		if err != nil {
			return nil, response.ToApiError(err)
		}
		resp := response.CreateVoteGetResponseFromVoteEntity(vote)
		return resp, nil
	})

	// POST /v1/votes/{country_id}
	huma.Register(grp, huma.Operation{
		OperationID: "add-vote",
		Method:      http.MethodPost,
		Path:        "/{country_id}",
		Summary:     "Add new vote",
	}, func(reqCtx context.Context, command *command.VoteAdd) (*response.VoteAddResponse, error) {
		vote, err := votesService.AddUserVote(reqCtx, *command)
		if err != nil {
			return nil, response.ToApiError(err)
		}
		resp := response.CreateVoteAddResponseFromVoteEntity(vote)
		return resp, nil
	})

	// PATCH /v1/votes/{country_id}/{active_user_id}/{peer_id}/change-contract
	huma.Register(grp, huma.Operation{
		OperationID: "change-vote",
		Method:      http.MethodPatch,
		Path:        "/{country_id}/{active_user_id}/{peer_id}/change-contract",
		Summary:     "Change active user vote contract",
		Responses:   apiResponse.GenerateErrorResponsesGroup(grp, 404),
	}, func(reqCtx context.Context, command *command.ChangeVoteType) (*response.ChangeVoteResponse, error) {
		vote, err := votesService.ChangeUserVote(reqCtx, *command)
		if err != nil {
			return nil, response.ToApiError(err)
		}
		resp := response.CreateChangeVoteResponseFromVoteEntity(vote)
		return resp, nil
	})

	// DELETE /v1/votes/{country_id}/{active_user_id}/{peer_id}
	huma.Register(grp, huma.Operation{
		OperationID: "delete-vote",
		Method:      http.MethodDelete,
		Path:        "/{country_id}/{active_user_id}/{peer_id}",
		Summary:     "Delete active user vote",
	}, func(reqCtx context.Context, command *command.DeleteVote) (*struct{}, error) {
		err := votesService.DeleteUserVote(reqCtx, *command)
		if err != nil {
			return nil, response.ToApiError(err)
		}
		return nil, nil
	})
}

func registerCountersRoutes(
	grp *huma.Group,
	votesService *application.VotingService,
) {
	grp = huma.NewGroup(grp, "/counters")
	grp.UseSimpleModifier(func(op *huma.Operation) {
		op.Tags = []string{"Counters"}
	})

	// GET /v1/counters/{country_id}/{active_user_id}/lifetime
	huma.Register(grp, huma.Operation{
		OperationID: "get-lifetime-counters",
		Method:      http.MethodGet,
		Path:        "/{country_id}/{active_user_id}/lifetime",
		Summary:     "Get lifetime counters for the active user",
	}, func(reqCtx context.Context, query *query.LifetimeCountersGet) (*response.LifetimeCountersGetResponse, error) {
		countersGroup, err := votesService.GetLifetimeCounters(reqCtx, *query)
		if err != nil {
			return nil, response.ToApiError(err)
		}
		resp := response.CreateLifetimeCountersGetResponseFromCountersGroup(countersGroup)
		return resp, nil
	})

	// GET /v1/counters/{country_id}/{active_user_id}/hourly
	huma.Register(grp, huma.Operation{
		OperationID: "get-hourly-counters",
		Method:      http.MethodGet,
		Path:        "/{country_id}/{active_user_id}/hourly",
		Summary:     "Get hourly counters for the active user",
	}, func(reqCtx context.Context, query *query.HourlyCountersGet) (*response.HourlyCountersGetResponse, error) {
		countersGroup, err := votesService.GetHourlyCounters(reqCtx, *query)
		if err != nil {
			return nil, response.ToApiError(err)
		}
		resp := response.CreateHourlyCountersGetResponseFromCountersGroup(countersGroup)
		return resp, nil
	})
}
