module github.com/splunk/splunk_otel_dotnet_deployer

go 1.23.10

require github.com/stretchr/testify v1.10.0

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/splunk/splunk_otel_dotnet_deployer/internal/modularinput => ./internal/modularinput
