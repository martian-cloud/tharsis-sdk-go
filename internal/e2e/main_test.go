//go:build integration
// +build integration

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/aws/smithy-go/ptr"
	tharsis "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/config"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// The lone TestMain function here sets up for, launches, and cleans up for
// all Tharsis SDK integration tests.  All tests use the special group defined
// next.
//
// The top-level group cannot be deleted or created from the service account.

const (
	topGroupName = "tharsis-sdk-e2e-tests"
)

func TestMain(m *testing.M) {
	log.Println("Starting TestMain:")
	ctx := context.Background()
	client, err := createClient()
	if err != nil {
		log.Printf("failed to create a Tharsis SDK client: %v\n", err)
		os.Exit(1)
	}
	toTeardown, err := setup(ctx, client)
	if err != nil {
		log.Printf("failed to set up for integration tests: %v\n", err)
		os.Exit(1)
	}
	resultCode := m.Run()
	// Attempt to tear down even if some tests failed.
	err = toTeardown(client)
	if err != nil {
		log.Printf("failed to tear down from integration tests: %v\n", err)
		os.Exit(1)
	}

	log.Printf("Finished TestMain: resultCode: %d\n", resultCode)
	os.Exit(resultCode)
}

// createClient creates the client and the prerequisites.
// It can be called by functions in this file and by test functions.
func createClient() (*tharsis.Client, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	client, err := tharsis.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	return client, nil
}

// setup sets things up before the tests are run.
// It returns a function that can tear down things afterward.
//
// Experiments showed that trying to handle flags here does not work.
//
func setup(ctx context.Context, client *tharsis.Client) (func(client *tharsis.Client) error, error) {
	log.Println("Setting up Tharsis SDK integration tests...")

	// It it becomes necessary to build a binary for these tests, build it here.
	// See Terraform's internal/cloud/e2e tests.

	// Make sure the top-level group already exists.
	topGroup, err := client.Group.GetGroup(ctx, &types.GetGroupInput{Path: ptr.String(topGroupName)})
	if err != nil {
		return nil, err
	}
	if topGroup == nil {
		return nil, fmt.Errorf("in early setup, top-level group must already exist")
	}

	log.Println("Finished setting up Tharsis SDK integration tests.")
	return func(client *tharsis.Client) error {
		return teardown(ctx, client)
	}, nil
}

// teardown tears down the things that the setup function set up.
func teardown(ctx context.Context, client *tharsis.Client) error {
	log.Println("Tearing down from Tharsis SDK integration tests...")

	log.Println("Finished tearing down from Tharsis SDK integration tests.")
	return nil
}

// The End.
