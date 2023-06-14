package prometheusexporter

import (
	"errors"
	"testing"
)

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name               string
		scrapeFailureLevel string
		expectedError      error
	}{
		{
			name:               "Valid log level: debug",
			scrapeFailureLevel: "debug",
			expectedError:      nil,
		},
		{
			name:               "Valid log level: info",
			scrapeFailureLevel: "info",
			expectedError:      nil,
		},
		{
			name:               "Valid log level: warn",
			scrapeFailureLevel: "warn",
			expectedError:      nil,
		},
		{
			name:               "Valid log level: error",
			scrapeFailureLevel: "error",
			expectedError:      nil,
		},
		{
			name:               "Invalid log level",
			scrapeFailureLevel: "badValue",
			expectedError:      errors.New("not a valid logrus Level: \"badValue\""),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			config := &Config{
				ScrapeFailureLogLevel: test.scrapeFailureLevel,
			}

			err := config.Validate()

			if test.expectedError != nil {
				if err == nil {
					t.Errorf("Expected error '%s', but got nil", test.expectedError.Error())
				} else if err.Error() != test.expectedError.Error() {
					t.Errorf("Expected error '%s', but got '%s'", test.expectedError.Error(), err.Error())
				}
			} else if err != nil {
				t.Errorf("Expected no error, but got '%s'", err.Error())
			}
		})
	}
}
