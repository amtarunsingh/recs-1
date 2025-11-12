package response

import (
	huma "github.com/danielgtaylor/huma/v2"
	"net/http"
	"reflect"
	"strconv"
)

type HumaApiError struct {
	Status  int                 `json:"status" example:"400" doc:"HTTP status code"`
	Message string              `json:"message" example:"Property foo is required but is missing." doc:"A human-readable explanation specific to this occurrence of the problem."`
	Errors  []*huma.ErrorDetail `json:"errors,omitempty" doc:"Optional list of individual error details"`
}

func (e *HumaApiError) Error() string {
	return e.Message
}

func (e *HumaApiError) GetStatus() int {
	return e.Status
}

func GenerateErrorResponsesGroup(grp *huma.Group, codes ...int) map[string]*huma.Response {
	responses := map[string]*huma.Response{}
	for _, code := range codes {
		responses[strconv.Itoa(code)] = GenerateErrorResponse(grp, code)
	}
	return responses
}

func GenerateErrorResponse(grp *huma.Group, code int) *huma.Response {
	reg := grp.OpenAPI().Components.Schemas
	errSchema := huma.SchemaFromType(reg, reflect.TypeOf(HumaApiError{}))
	return &huma.Response{
		Description: http.StatusText(code),
		Content: map[string]*huma.MediaType{
			"application/json": {
				Schema: errSchema,
			},
		},
	}
}
