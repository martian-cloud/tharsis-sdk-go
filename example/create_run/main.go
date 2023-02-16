// example/configuration_version/main.go contains an example for creating/uploading a configuration version.

// Package main contains create run examples
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	tharsis "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/config"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// ExampleCreateRun demonstrates CreateConfigurationVersion and UploadConfigurationVersion:
func ExampleCreateRun(workspacePath, directoryPath string) error {

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
	createdConfigurationVersion, err := client.ConfigurationVersion.CreateConfigurationVersion(ctx,
		&types.CreateConfigurationVersionInput{
			WorkspacePath: workspacePath,
			Speculative:   &isTrue,
		})
	if err != nil {
		return err
	}

	// Print:
	indented, err := json.MarshalIndent(createdConfigurationVersion, "", "  ")
	if err != nil {
		return err
	}
	fmt.Printf("\nCreated configuration version:\n%s\n\n", indented)

	// Second, upload the configuration version.
	err = client.ConfigurationVersion.UploadConfigurationVersion(ctx,
		&types.UploadConfigurationVersionInput{
			WorkspacePath:          workspacePath,
			ConfigurationVersionID: createdConfigurationVersion.Metadata.ID,
			DirectoryPath:          directoryPath,
		})
	if err != nil {
		return err
	}

	// Wait for the upload to complete:
	var updatedConfigurationVersion *types.ConfigurationVersion
	for {
		updatedConfigurationVersion, err = client.ConfigurationVersion.GetConfigurationVersion(ctx,
			&types.GetConfigurationVersionInput{ID: createdConfigurationVersion.Metadata.ID})
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

	// Print:
	updatedConfigurationVersion, err = client.ConfigurationVersion.GetConfigurationVersion(ctx,
		&types.GetConfigurationVersionInput{ID: createdConfigurationVersion.Metadata.ID})
	if err != nil {
		return err
	}
	indented, err = json.MarshalIndent(updatedConfigurationVersion, "", "  ")
	if err != nil {
		return err
	}
	fmt.Printf("Updated configuration version:\n%s\n\n", indented)

	// Third, create (and plan) the run.
	createdRun, err := client.Run.CreateRun(ctx,
		&types.CreateRunInput{
			WorkspacePath:          workspacePath,
			ConfigurationVersionID: &createdConfigurationVersion.Metadata.ID,
			IsDestroy:              false,
		})
	if err != nil {
		return err
	}

	// Print:
	indented, err = json.MarshalIndent(createdRun, "", "  ")
	if err != nil {
		return err
	}
	fmt.Printf("Created run:\n%s\n", indented)

	if createdRun.Plan == nil {
		return fmt.Errorf("run plan was nil")
	}
	planJobID := createdRun.Plan.CurrentJobID

	// Fourth, process the plan logs.
	logChannel, err := client.Job.SubscribeToJobLogs(ctx, &types.JobLogsSubscriptionInput{
		RunID:         createdRun.Metadata.ID,
		WorkspacePath: createdRun.WorkspacePath,
		JobID:         *planJobID,
	})
	if err != nil {
		return err
	}

	fmt.Println("Starting plan job logs:")
	for {
		logsEvent, ok := <-logChannel
		if !ok {
			break
		}

		if logsEvent.Error != nil {
			// Catch any incoming errors.
			return logsEvent.Error
		}

		_, err = os.Stdout.Write([]byte(logsEvent.Logs))
		if err != nil {
			return err
		}
	}
	fmt.Println("Finished plan job logs.")

	return nil
}

// The End.
