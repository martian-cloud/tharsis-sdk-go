package tharsis

import (
	"context"
	"fmt"

	"github.com/hasura/go-graphql-client"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/internal"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// TerraformCLIVersion implements the function related to Terraform CLI Versions.
type TerraformCLIVersion interface {
	CreateTerraformCLIDownloadURL(ctx context.Context, input *types.CreateTerraformCLIDownloadURLInput) (string, error)
}

type terraformCLIVersion struct {
	client *Client
}

// NewTerraformCLIVersion returns a TerraformCLIVersion object.
func NewTerraformCLIVersion(client *Client) TerraformCLIVersion {
	return &terraformCLIVersion{client: client}
}

// CreateTerraformCLIDownloadURL returns a URL where Terraform CLI can be downloaded from.
func (t *terraformCLIVersion) CreateTerraformCLIDownloadURL(ctx context.Context,
	input *types.CreateTerraformCLIDownloadURLInput) (string, error) {
	var wrappedCreate struct {
		CreateTerraformCLIDownloadURL struct {
			DownloadURL graphql.String
			Problems    []internal.GraphQLProblem
		} `graphql:"createTerraformCLIDownloadURL(input: $input)"`
	}

	variables := map[string]interface{}{
		"input": *input,
	}

	// Execute mutation request.
	err := t.client.graphqlClient.Mutate(ctx, &wrappedCreate, variables)
	if err != nil {
		return "", err
	}

	err = internal.ProblemsToError(wrappedCreate.CreateTerraformCLIDownloadURL.Problems)
	if err != nil {
		return "", fmt.Errorf("problems assigning managed identity to workspace: %v", err)
	}

	return string(wrappedCreate.CreateTerraformCLIDownloadURL.DownloadURL), nil
}
