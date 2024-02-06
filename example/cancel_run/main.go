// Package main contains cancel run examples
package main

import (
	"context"
	"encoding/json"
	"fmt"

	tharsis "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/config"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// ExampleCancelRun demonstrates canceling the specified run.
func ExampleCancelRun(runID string, force bool) error {

	cfg, err := config.Load(config.WithEndpoint("http://localhost:8000"))
	if err != nil {
		return err
	}

	client, err := tharsis.NewClient(cfg)
	if err != nil {
		return err
	}

	// Second, cancel the run.
	ctx := context.Background()
	comment := "This is an example cancel operation on a run."
	canceledRun, err := client.Run.CancelRun(ctx,
		&types.CancelRunInput{
			RunID:   runID,
			Comment: &comment,
			Force:   &force,
		})
	if err != nil {
		return err
	}

	// Print:
	indented, err := json.MarshalIndent(canceledRun, "", "  ")
	if err != nil {
		return err
	}
	fmt.Printf("\nCanceled run:\n%s\n\n", indented)

	return nil
}
