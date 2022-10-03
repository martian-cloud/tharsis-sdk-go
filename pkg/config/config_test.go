package config

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigLoad(t *testing.T) {

	clearEnvVars()

	tests := []struct {
		expectedErr error
		expectedCfg *Config
		input       Config
		testName    string
	}{
		{
			testName:    "empty",
			input:       Config{Logger: log.Default()},
			expectedCfg: nil,
			expectedErr: fmt.Errorf("unable to create a token provider:" +
				" to use a service account token, set environment variables" +
				" THARSIS_SERVICE_ACCOUNT_PATH and THARSIS_SERVICE_ACCOUNT_TOKEN;" +
				" to use a static token, set environment variable THARSIS_STATIC_TOKEN"),
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			actual, err := Load(WithLogger(log.Default()))
			assert.Equal(t, test.expectedErr, err)
			assert.True(t, reflect.DeepEqual(test.expectedCfg, actual))
		})
	}

}

// clearEnvVars clears relevant environment variables
func clearEnvVars() {
	os.Unsetenv("THARSIS_ENDPOINT")
	os.Unsetenv("THARSIS_STATIC_TOKEN")
	os.Unsetenv("THARSIS_SERVICE_ACCOUNT_PATH")
	os.Unsetenv("THARSIS_SERVICE_ACCOUNT_TOKEN")
}

// The End.
