// Package job is responsible for displaying job logs.
package job

import (
	"context"
	"os"

	tharsis "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// DisplayLogs constants.
const (
	logLimit int32 = 2 * 1024 * 1024
)

// DisplayLogsInput is the input for displaying job logs.
type DisplayLogsInput struct {
	Client      *tharsis.Client
	RunID       string
	WorkspaceID string
	JobID       string
}

// DisplayLogs displays job logs.
func DisplayLogs(ctx context.Context, input *DisplayLogsInput) error {
	// Subscribe to run events so we know when to stop outputting logs.
	runEvents, err := input.Client.Run.SubscribeToWorkspaceRunEvents(ctx, &types.RunSubscriptionInput{
		RunID:         &input.RunID,
		WorkspacePath: input.WorkspaceID,
	})
	if err != nil {
		return err
	}

	// Subscribe to the job log events so we can fetch logs only when new ones are available.
	logEvents, err := input.Client.Job.SubscribeToJobLogEvents(ctx, &types.JobLogSubscriptionInput{
		JobID: input.JobID,
	})
	if err != nil {
		return err
	}

	var (
		currentOffset int32
		runCompleted  bool
	)

	for {
		// Retrieve the plan logs.
		output, err := input.Client.Job.GetJobLogs(ctx, &types.GetJobLogsInput{
			ID:          input.JobID,
			StartOffset: currentOffset,
			Limit:       logLimit,
		})
		if err != nil {
			return err
		}

		// Update the offset to the new log size.
		currentOffset += int32(len(output.Logs))

		if output.LogSize > 0 {
			// Write the logs to console.
			os.Stdout.WriteString(output.Logs)
		}

		if runCompleted {
			// Run has finished and all pending logs are displayed.
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-logEvents:
		case eventRun := <-runEvents:
			switch eventRun.Status {
			case types.RunApplied,
				types.RunCanceled,
				types.RunPlanned,
				types.RunPlannedAndFinished,
				types.RunErrored:
				runCompleted = true

				if currentOffset == output.LogSize {
					// All logs have been displayed.
					return nil
				}
			}
		}
	}
}
