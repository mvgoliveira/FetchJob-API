package errors

import "errors"

var (
	ErrWhileQueryingOnDatabase = errors.New("error while querying the database")
	ErrNoDataFoundOnDatabase   = errors.New("no data found in the database")
	ErrUnmarshallingData       = errors.New("could not unmarshall data")
	ErrCouldNotSaveToDatabase  = errors.New("could not save to database")
	ErrUserNotAuthorized       = errors.New("user does not have access to the requested operation")
)
