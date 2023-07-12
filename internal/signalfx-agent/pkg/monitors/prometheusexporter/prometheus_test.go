package prometheusexporter

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testCase struct {
	name                  string
	scrapeFailureLogLevel string
	expectedError         error
}

func TestConfigValidate(t *testing.T) {
	testCases := []testCase{
		{
			name:                  "Valid log level: debug",
			scrapeFailureLogLevel: "debug",
			expectedError:         nil,
		},
		{
			name:                  "Valid log level: info",
			scrapeFailureLogLevel: "info",
			expectedError:         nil,
		},
		{
			name:                  "Valid log level: warn",
			scrapeFailureLogLevel: "warn",
			expectedError:         nil,
		},
		{
			name:                  "Valid log level: error",
			scrapeFailureLogLevel: "error",
			expectedError:         nil,
		},
		{
			name:                  "Invalid log level",
			scrapeFailureLogLevel: "badValue",
			expectedError:         errors.New("not a valid logrus Level: \"badValue\""),
		},
	}

	for _, tc := range testCases {
		tcc := tc
		t.Run(tc.name, func(t *testing.T) {
			config := &Config{
				ScrapeFailureLogLevel: tcc.scrapeFailureLogLevel,
			}

			err := config.Validate()

			if tcc.expectedError != nil {
				if err == nil {
					t.Errorf("Expected error '%s', but got nil", tcc.expectedError.Error())
				} else if err.Error() != tcc.expectedError.Error() {
					t.Errorf("Expected error '%s', but got '%s'", tcc.expectedError.Error(), err.Error())
					assert.EqualValues(t, "smartagentvalid", config.MonitorConfigCore().MonitorID)
				}
			} else if err != nil {
				t.Errorf("Expected no error, but got '%s'", err.Error())
			}
		})
	}
}
