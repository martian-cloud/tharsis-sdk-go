// example/configuration_version/main.go contains an example for creating/uploading a configuration version.

// Package main contains configuration version examples
package main

import (
	"context"
	"encoding/json"
	"fmt"

	tharsis "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/config"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// ExampleCreateUploadConfigurationVersion demonstrates CreateConfigurationVersion and UploadConfigurationVersion:
func ExampleCreateUploadConfigurationVersion(workspacePath, directoryPath string) error {

	cfg, err := config.Load(config.WithEndpoint("http://localhost:8000"))
	if err != nil {
		return err
	}

	client, err := tharsis.NewClient(cfg)
	if err != nil {
		return err
	}

	// First, create the configuration version.
	ctx := context.Background()
	isTrue := true
	created, err := client.ConfigurationVersion.CreateConfigurationVersion(ctx,
		&types.CreateConfigurationVersionInput{
			WorkspacePath: workspacePath,
			Speculative:   &isTrue,
		})
	if err != nil {
		return err
	}

	// Print:
	indented, err := json.MarshalIndent(created, "", "  ")
	if err != nil {
		return err
	}
	fmt.Printf("Created configuration version:\n%s\n", indented)

	// Second, upload the configuration version.
	err = client.ConfigurationVersion.UploadConfigurationVersion(ctx,
		&types.UploadConfigurationVersionInput{
			WorkspacePath:          workspacePath,
			ConfigurationVersionID: created.Metadata.ID,
			DirectoryPath:          directoryPath,
		})
	if err != nil {
		return err
	}

	// Wait for the upload to complete:
	var updatedConfigurationVersion *types.ConfigurationVersion
	for {
		updatedConfigurationVersion, err = client.ConfigurationVersion.GetConfigurationVersion(ctx,
			&types.GetConfigurationVersionInput{ID: created.Metadata.ID})
		if err != nil {
			return err
		}
		if updatedConfigurationVersion.Status != "pending" {
			break
		}
	}
	if updatedConfigurationVersion.Status != "uploaded" {
		return fmt.Errorf("upload failed; status is %s", updatedConfigurationVersion.Status)
	}

	return nil
}
