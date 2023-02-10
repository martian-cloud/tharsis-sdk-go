// Package internal contains internal functionality
package internal

import "github.com/hasura/go-graphql-client"

// StringPointerFromGraphQL converts a graphql.String pointer to a string pointer.
func StringPointerFromGraphQL(arg *graphql.String) *string {
	if arg == nil {
		return nil
	}
	result := string(*arg)
	return &result
}

// StringSliceFromGraphQL converts a slice of graphql.String to a slice of string.
func StringSliceFromGraphQL(arg []graphql.String) []string {
	result := []string{}

	for _, gs := range arg {
		result = append(result, string(gs))
	}

	return result
}
