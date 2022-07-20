package api

import (
	"fmt"

	"github.com/unweave/cli/errors"
	"github.com/unweave/cli/pkg/graphql"
)

// parseGqlError parses a GraphQL error and returns a known error type. It assumes that
// the main reason for the call failing is contained in the first error.
func parseGqlError(errs *graphql.Errors) error {
	if len(*errs) == 0 {
		return nil
	}

	msg := (*errs)[0].Message
	ext := (*errs)[0].Extensions
	code := ext["code"]

	switch code {
	case "FORBIDDEN":
		return fmt.Errorf("%w: %v", errors.HttpForbiddenError, msg)
	case "NOT_FOUND":
		return fmt.Errorf("%w: %v", errors.HttpNotFoundError, msg)
	case "UNAUTHORIZED":
		return fmt.Errorf("%w: %v", errors.HttpUnAuthorized, msg)
	}

	return errors.UnknownError
}
