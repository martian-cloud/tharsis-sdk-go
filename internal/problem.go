package internal

import (
	"github.com/hasura/go-graphql-client"
)

// These types and constants copied from the Tharsis API and modified as needed.

// GraphQLProblemType represents the type of problem
type GraphQLProblemType graphql.String

// Problem constants
const (
	Conflict           GraphQLProblemType = "CONFLICT"
	BadRequest         GraphQLProblemType = "BAD_REQUEST"
	NotFound           GraphQLProblemType = "NOT_FOUND"
	Forbidden          GraphQLProblemType = "FORBIDDEN"
	ServiceUnavailable GraphQLProblemType = "SERVICE_UNAVAILABLE"
)

// GraphQLProblem is used to represent a user facing issue
type GraphQLProblem struct {
	Message graphql.String
	Type    GraphQLProblemType
	Field   []graphql.String
}
