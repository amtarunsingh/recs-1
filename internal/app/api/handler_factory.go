package api

import (
	"context"
	"github.com/bmbl-bumble2/recs-votes-storage/config"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/app/api/response"
	votingV1 "github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/interface/api/rest/v1"
	huma "github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"net/http"
)

type HandlerFactory struct {
	votesStorageRoutesRegister votingV1.VotesStorageRoutesRegister
}

func NewHandlerFactory(
	votesStorageRoutesRegister votingV1.VotesStorageRoutesRegister,
) HandlerFactory {
	return HandlerFactory{
		votesStorageRoutesRegister: votesStorageRoutesRegister,
	}
}

func (s HandlerFactory) NewHumaApiServerHandler() http.Handler {
	handler := http.NewServeMux()
	api := humago.New(handler, huma.DefaultConfig(config.ProjectName, config.ProjectVersion))
	grp := huma.NewGroup(api, "/v1")

	s.registerHealthCheck(api)
	s.setApiErrorSchema()
	s.registerDefaultOpenApiErrorsResponses(grp, 400, 422, 500)

	s.votesStorageRoutesRegister.RegisterV1Routes(grp)

	return handler
}

func (s HandlerFactory) registerHealthCheck(api huma.API) {
	huma.Register(api, huma.Operation{
		Method:  http.MethodGet,
		Path:    "/health",
		Summary: "Health check",
		Tags:    []string{"System"},
	}, func(ctx context.Context, input *struct{}) (*struct{}, error) {
		return nil, nil
	})
}

func (s HandlerFactory) registerDefaultOpenApiErrorsResponses(grp *huma.Group, codes ...int) {
	grp.UseSimpleModifier(func(op *huma.Operation) {
		for code, resp := range response.GenerateErrorResponsesGroup(grp, codes...) {
			op.Responses[code] = resp
		}
	})
}

func (s HandlerFactory) setApiErrorSchema() {
	huma.NewError = func(status int, message string, errs ...error) huma.StatusError {

		details := make([]*huma.ErrorDetail, len(errs))
		for i := 0; i < len(errs); i++ {
			if converted, ok := errs[i].(huma.ErrorDetailer); ok {
				details[i] = converted.ErrorDetail()
			} else {
				if errs[i] == nil {
					continue
				}
				details[i] = &huma.ErrorDetail{Message: errs[i].Error()}
			}
		}

		return &response.HumaApiError{
			Status:  status,
			Message: message,
			Errors:  details,
		}
	}
}
