package internal

import (
	"fmt"
	"strings"

	"github.com/hasura/go-graphql-client"
)

// These types and constants copied from the Tharsis API and modified as needed.

// GraphQLProblemType represents the type of problem
type GraphQLProblemType graphql.String

// Problem constants
const (
	Conflict   GraphQLProblemType = "CONFLICT"
	BadRequest GraphQLProblemType = "BAD_REQUEST"
	NotFound   GraphQLProblemType = "NOT_FOUND"
)

// GraphQLProblem is used to represent a user facing issue
type GraphQLProblem struct {
	Message graphql.String
	Type    GraphQLProblemType
	Field   []graphql.String
}

// ProblemsToError returns an error or nil.
func ProblemsToError(problems []GraphQLProblem) error {
	if len(problems) == 0 {
		return nil
	}
	var s []string
	for _, p := range problems {
		s = append(s, string(p.Message))
	}
	return fmt.Errorf(strings.Join(s, "; "))
}

// The End.
