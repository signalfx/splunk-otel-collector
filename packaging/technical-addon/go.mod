module github.com/splunk/otel-technical-addon

go 1.23.0

require (
	github.com/google/go-cmp v0.7.0
	github.com/stretchr/testify v1.10.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
)

replace github.com/splunk/otel-technical-addon/internal/modularinput => ./internal/modularinput
