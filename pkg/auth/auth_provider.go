// Package auth package
package auth

// TokenProvider is an interface with a GetToken method.
type TokenProvider interface {
	GetToken() (string, error)
}
