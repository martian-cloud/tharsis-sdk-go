// example/configuration_version/main.go contains an example for creating/uploading a configuration version
// and from a remote module source.

// Package main contains examples on how to apply a run
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

// ExampleApplyRun demonstrates CreateConfigurationVersion and UploadConfigurationVersion
// followed by creating, planning, and applying the run:
func ExampleApplyRun(workspacePath, directoryPath string) error {

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
	isFalse := false
	createdConfigurationVersion, err := client.ConfigurationVersion.CreateConfigurationVersion(ctx,
		&types.CreateConfigurationVersionInput{
			WorkspacePath: workspacePath,
			Speculative:   &isFalse,
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
		return fmt.Errorf("Upload failed; status is %s", updatedConfigurationVersion.Status)
	}

	// Print:
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
	logChannel, err := client.Job.GetJobLogs(ctx, &types.GetJobLogsInput{
		ID:          *planJobID,
		StartOffset: 0,
		Limit:       5 * 1024 * 1024,
	})
	if err != nil {
		return err
	}

	fmt.Println("Starting plan job logs:")
	for {
		logs, ok := <-logChannel
		if !ok {
			break
		}
		_, err = os.Stdout.Write([]byte(logs))
		if err != nil {
			return err
		}
	}
	fmt.Println("Finished plan job logs.")

	// Fifth, apply the run.
	applyComment := "This is a run apply comment."
	appliedRun, err := client.Run.ApplyRun(ctx,
		&types.ApplyRunInput{
			RunID:   createdRun.Metadata.ID,
			Comment: &applyComment,
		})
	if err != nil {
		return err
	}

	// Print:
	indented, err = json.MarshalIndent(appliedRun, "", "  ")
	if err != nil {
		return err
	}
	fmt.Printf("\nApplied run:\n%s\n", indented)

	if appliedRun.Apply == nil {
		return fmt.Errorf("run apply was nil")
	}
	applyJobID := appliedRun.Apply.CurrentJobID

	// Sixth, process the apply logs.
	logChannel, err = client.Job.GetJobLogs(ctx, &types.GetJobLogsInput{
		ID:          *applyJobID,
		StartOffset: 0,
		Limit:       5 * 1024 * 1024,
	})
	if err != nil {
		return err
	}

	fmt.Println("Starting apply job logs:")
	for {
		logs, ok := <-logChannel
		if !ok {
			break
		}
		_, err = os.Stdout.Write([]byte(logs))
		if err != nil {
			return err
		}
	}
	fmt.Println("Finished apply job logs.")

	return nil
}

// ExampleApplyModule demonstrates the plan and apply steps for a remote module:
func ExampleApplyModule(workspacePath, moduleSource, moduleVersion string) error {
	ctx := context.Background()

	cfg, err := config.Load(config.WithEndpoint("http://localhost:8000"))
	if err != nil {
		return err
	}

	client, err := tharsis.NewClient(cfg)
	if err != nil {
		return err
	}

	// Third, create (and plan) the run.
	createdRun, err := client.Run.CreateRun(ctx,
		&types.CreateRunInput{
			WorkspacePath: workspacePath,
			IsDestroy:     false,
			ModuleSource:  &moduleSource,
			ModuleVersion: &moduleVersion,
		})
	if err != nil {
		return err
	}

	// Print:
	indented, err := json.MarshalIndent(createdRun, "", "  ")
	if err != nil {
		return err
	}
	fmt.Printf("Created run:\n%s\n", indented)

	if createdRun.Plan == nil {
		return fmt.Errorf("run plan was nil")
	}
	planJobID := createdRun.Plan.CurrentJobID

	// Fourth, process the plan logs.
	logChannel, err := client.Job.GetJobLogs(ctx, &types.GetJobLogsInput{
		ID:          *planJobID,
		StartOffset: 0,
		Limit:       5 * 1024 * 1024,
	})
	if err != nil {
		return err
	}

	fmt.Println("Starting plan job logs:")
	for {
		logs, ok := <-logChannel
		if !ok {
			break
		}
		_, err = os.Stdout.Write([]byte(logs))
		if err != nil {
			return err
		}
	}
	fmt.Println("Finished plan job logs.")

	// Fifth, apply the run.
	applyComment := "This is a run apply comment."
	appliedRun, err := client.Run.ApplyRun(ctx,
		&types.ApplyRunInput{
			RunID:   createdRun.Metadata.ID,
			Comment: &applyComment,
		})
	if err != nil {
		return err
	}

	// Print:
	indented, err = json.MarshalIndent(appliedRun, "", "  ")
	if err != nil {
		return err
	}
	fmt.Printf("\nApplied run:\n%s\n", indented)

	if appliedRun.Apply == nil {
		return fmt.Errorf("run apply was nil")
	}
	applyJobID := appliedRun.Apply.CurrentJobID

	// Sixth, process the apply logs.
	logChannel, err = client.Job.GetJobLogs(ctx, &types.GetJobLogsInput{
		ID:          *applyJobID,
		StartOffset: 0,
		Limit:       5 * 1024 * 1024,
	})
	if err != nil {
		return err
	}

	fmt.Println("Starting apply job logs:")
	for {
		logs, ok := <-logChannel
		if !ok {
			break
		}
		_, err = os.Stdout.Write([]byte(logs))
		if err != nil {
			return err
		}
	}
	fmt.Println("Finished apply job logs.")

	return nil
}

// The End.
