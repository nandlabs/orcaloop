package api

import (
	"strconv"

	"oss.nandlabs.io/golly/rest"
	"oss.nandlabs.io/orcaloop-sdk/models"
)

type APIBaseResponse struct {
	Error *models.Error `json:"error,omitempty" yaml:"error,omitempty"`
}

func RespondWithError(ctx rest.ServerContext, code int, message string, err error) {
	var details string
	if err != nil {
		details = err.Error()
	}
	errObj := &APIBaseResponse{
		Error: &models.Error{
			Code:    strconv.Itoa(code),
			Message: message,
			Details: details,
		},
	}
	ctx.WriteJSON(errObj)
	ctx.SetStatusCode(code)
}
