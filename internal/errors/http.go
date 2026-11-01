package errors

import "errors"

var (
	ErrBindingRequestBody      = errors.New("could not bind the request body")
	ErrBindingExternalResponse = errors.New("could not bind the external response")
	ErrRequestingToExternalAPI = errors.New("could not complete the request to an external api")
	ErrInvalidUUID             = errors.New("invalid uuid format")
	ErrInvalidQueryParameters  = errors.New("invalid query parameter")
	ErrInvalidParameters       = errors.New("invalid parameters")
	ErrInvalidDateFormat       = errors.New("invalid date format, expected format is YYYY-MM-DD")
)
