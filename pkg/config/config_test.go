package config

import (
	"log"
	"os"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigLoad(t *testing.T) {

	clearEnvVars()

	testLogger := log.Default()
	testEndpoint := "http://some/test/endpoint"

	tests := []struct {
		expectedErr error
		expectedCfg *Config
		input       Config
		testName    string
	}{
		{
			testName: "empty",
			input:    Config{Logger: testLogger},
			expectedCfg: &Config{
				Logger:   log.Default(),
				Endpoint: testEndpoint,
			},
			expectedErr: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			actual, err := Load(WithLogger(testLogger), WithEndpoint(testEndpoint))
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
