package errors

import (
	goErr "errors"
)

var UnknownError = goErr.New("unknown error")
var NotLoggedInError = goErr.New("not logged in")
var HttpUnAuthorized = goErr.New("unauthorized")
var HttpForbiddenError = goErr.New("access denied")
var HttpNotFoundError = goErr.New("not found")
