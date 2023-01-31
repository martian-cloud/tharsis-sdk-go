package tharsis

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/hashicorp/go-slug"
	"github.com/hasura/go-graphql-client"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/internal"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

const (
	// options for creating a temporary tarfile
	tarFlagWrite = os.O_CREATE | os.O_TRUNC | os.O_WRONLY
	tarMode      = 0600
)

// ConfigurationVersion implements functions related to Tharsis configuration versions.
type ConfigurationVersion interface {
	GetConfigurationVersion(ctx context.Context,
		input *types.GetConfigurationVersionInput) (*types.ConfigurationVersion, error)
	CreateConfigurationVersion(ctx context.Context,
		input *types.CreateConfigurationVersionInput) (*types.ConfigurationVersion, error)
	DownloadConfigurationVersion(ctx context.Context,
		input *types.GetConfigurationVersionInput, writer io.WriterAt) error
	UploadConfigurationVersion(ctx context.Context, input *types.UploadConfigurationVersionInput) error
}

type configurationVersion struct {
	client *Client
}

// NewConfigurationVersion returns a ConfigurationVersion object.
func NewConfigurationVersion(client *Client) ConfigurationVersion {
	return &configurationVersion{client: client}
}

// GetConfigurationVersion returns everything about the configuration version.
func (cv *configurationVersion) GetConfigurationVersion(ctx context.Context,
	input *types.GetConfigurationVersionInput) (*types.ConfigurationVersion, error) {
	var target struct {
		ConfigurationVersion *graphQLConfigurationVersion `graphql:"configurationVersion(id: $id)"`
	}
	variables := map[string]interface{}{"id": graphql.String(input.ID)}

	err := cv.client.graphqlClient.Query(ctx, true, &target, variables)
	if err != nil {
		return nil, err
	}
	if target.ConfigurationVersion == nil {
		return nil, newError(ErrNotFound, "configuration version with id %s not found", input.ID)
	}

	result := configurationVersionFromGraphQL(*target.ConfigurationVersion)
	return &result, nil
}

// CreateConfigurationVersion creates a new configuration version and returns its content.
// This call returns a ConfigurationVersion object, which contains an ID, aka created.Metadata.ID here.
// Said ID is then used by the UploadConfigurationVersion operation as the ConfigurationVersionID.
func (cv *configurationVersion) CreateConfigurationVersion(ctx context.Context,
	input *types.CreateConfigurationVersionInput) (*types.ConfigurationVersion, error) {

	var wrappedCreate struct {
		CreateConfigurationVersion struct {
			ConfigurationVersion graphQLConfigurationVersion
			Problems             []internal.GraphQLProblem
		} `graphql:"createConfigurationVersion(input: $input)"`
	}

	// Creating a new object requires the wrapped object above
	// but with all the contents in a struct in the variables.
	variables := map[string]interface{}{
		"input": *input,
	}

	err := cv.client.graphqlClient.Mutate(ctx, true, &wrappedCreate, variables)
	if err != nil {
		return nil, err
	}

	if err = errorFromGraphqlProblems(wrappedCreate.CreateConfigurationVersion.Problems); err != nil {
		return nil, err
	}

	created := configurationVersionFromGraphQL(wrappedCreate.CreateConfigurationVersion.ConfigurationVersion)
	return &created, nil
}

//////////////////////////////////////////////////////////////////////////////

// Related types and conversion functions:

// graphQLConfigurationVersion represents the insides of the query structure,
// everything in the configuration version object,
// and with graphql types.
type graphQLConfigurationVersion struct {
	Metadata    internal.GraphQLMetadata
	ID          graphql.String
	Status      graphql.String
	WorkspaceID graphql.String
	Speculative graphql.Boolean
}

// configurationVersionFromGraphQL converts a GraphQL ConfigurationVersion to an external ConfigurationVersion.
func configurationVersionFromGraphQL(g graphQLConfigurationVersion) types.ConfigurationVersion {
	result := types.ConfigurationVersion{
		Metadata:    internal.MetadataFromGraphQL(g.Metadata, g.ID),
		Status:      string(g.Status),
		Speculative: bool(g.Speculative),
		WorkspaceID: string(g.WorkspaceID),
	}
	return result
}

//////////////////////////////////////////////////////////////////////////////

// DownloadConfigurationVersion downloads a configuration version and returns the response.
func (cv configurationVersion) DownloadConfigurationVersion(ctx context.Context,
	input *types.GetConfigurationVersionInput, writer io.WriterAt) error {

	url := strings.Join([]string{cv.client.cfg.Endpoint, "v1", "configuration-versions", input.ID, "content"}, "/")
	resp, err := cv.do(ctx, http.MethodGet, url, nil, 0)
	if err != nil {
		return err
	}

	return copyFromResponseBody(resp, writer)
}

// UploadConfigurationVersion uploads a directory (the CLI's current working
// directory).  It packages it into a temporary tar.gz archive and then sends
// a reader to the API.
func (cv *configurationVersion) UploadConfigurationVersion(ctx context.Context,
	input *types.UploadConfigurationVersionInput) error {

	// Package the directory into a temporary tar.gz file.
	tarPath, err := cv.makeTarfile(input.DirectoryPath)
	if err != nil {
		return err
	}
	defer os.Remove(tarPath)

	// Get the length of the tar file.
	stat, err := os.Stat(tarPath)
	if err != nil {
		return err
	}
	tarLen := stat.Size()

	// Open a reader on the tar.gz file.
	tarRdr, err := os.Open(tarPath) // nosemgrep: gosec.G304-1
	if err != nil {
		return err
	}

	// Let the API do the rest of the work.
	return cv.uploadTarfile(ctx, input.WorkspacePath, input.ConfigurationVersionID, tarRdr, tarLen)
}

func (cv *configurationVersion) makeTarfile(dirPath string) (string, error) {

	// Check the directory path.
	stat, err := os.Stat(dirPath)
	if err != nil {
		return "", err
	}
	if !stat.IsDir() {
		return "", fmt.Errorf("not a directory: %s", dirPath)
	}

	// Create the temporary tar.gz file.
	tarFile, err := os.CreateTemp("", "uploadConfigurationVersion.tgz")
	if err != nil {
		return "", err
	}
	tarPath := tarFile.Name()

	// Open a writer to the temporary tar.gz file.
	tgzFileWriter, err := os.OpenFile(tarPath, tarFlagWrite, tarMode) // nosemgrep: gosec.G304-1
	if err != nil {
		return "", err
	}
	defer tgzFileWriter.Close() // executes last (of the deferred closings)

	_, err = slug.Pack(dirPath, tgzFileWriter, true)
	if err != nil {
		return "", err
	}

	return tarPath, nil
}

// uploadTarfile calls the Tharsis REST API to upload a configuration version.
func (cv *configurationVersion) uploadTarfile(ctx context.Context,
	workspacePath, configurationVersionID string, rdr io.Reader, leng int64) error {

	// TODO: When the API has been changed to accept workspace path
	// rather than only workspace ID, remove this workaround and
	// pass the workspace path to the API rather than the ID.
	workspaceID, err := cv.translateWorkspacePathToID(ctx, workspacePath)
	if err != nil {
		return err
	}

	url := strings.Join([]string{cv.client.cfg.Endpoint, "v1", "workspaces", workspaceID,
		"configuration-versions", configurationVersionID, "upload"}, "/")
	resp, err := cv.do(ctx, http.MethodPut, url, rdr, leng)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("upload tar file response code: %d", resp.StatusCode)
	}

	return nil
}

// do prepares, makes a request with appropriate headers and returns the response.
func (cv *configurationVersion) do(ctx context.Context,
	method string, url string, body io.Reader, length int64) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	// Get the authentication token.
	authToken, err := cv.client.cfg.TokenProvider.GetToken()
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+authToken)

	// Set appropriate request headers.
	if method == http.MethodPut {
		req.Header.Set("Content-Type", "application/octet-stream")
		req.Header.Set("Content-Length", fmt.Sprintf("%d", length))
	} else {
		req.Header.Set("Accept", "application/json")
	}

	// Make the request.
	resp, err := cv.client.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errorFromHTTPResponse(resp)
	}

	return resp, nil
}

// TODO: When the API has been changed to accept a workspace path directly rather
// than only a workspace ID, this method can be removed.
func (cv *configurationVersion) translateWorkspacePathToID(ctx context.Context,
	workspacePath string) (string, error) {
	workspace, err := cv.client.Workspaces.GetWorkspace(ctx,
		&types.GetWorkspaceInput{Path: &workspacePath})
	if err != nil {
		return "", err
	}

	return workspace.Metadata.ID, nil
}

func copyFromResponseBody(r *http.Response, v interface{}) error {
	if w, ok := v.(io.Writer); ok {
		_, err := io.Copy(w, r.Body)
		return err
	}
	return nil
}

// The End.
