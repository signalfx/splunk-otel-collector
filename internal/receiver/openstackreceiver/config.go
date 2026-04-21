// Copyright Splunk, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package openstackreceiver

import (
	"errors"
	"time"

	"go.opentelemetry.io/collector/config/configopaque"
	"go.opentelemetry.io/collector/scraper/scraperhelper"
)

// Config defines the configuration for the OpenStack receiver.
type Config struct {
	// AuthURL is the Keystone identity endpoint (required).
	// Example: "http://192.168.1.1/identity/v3"
	AuthURL string `mapstructure:"auth_url"`

	// Username to authenticate with Keystone (required).
	Username string `mapstructure:"username"`

	// Password to authenticate with Keystone (required).
	Password configopaque.String `mapstructure:"password"`

	// ProjectName is the name of the project to monitor (default: "admin").
	ProjectName string `mapstructure:"project_name"`

	// ProjectDomainID is the project domain ID (default: "default").
	ProjectDomainID string `mapstructure:"project_domain_id"`

	// UserDomainID is the user domain ID (default: "default").
	UserDomainID string `mapstructure:"user_domain_id"`

	// RegionName is the region for URL discovery.
	// Defaults to the first available region if left empty.
	RegionName string `mapstructure:"region_name"`

	scraperhelper.ControllerConfig `mapstructure:",squash"`

	// HTTPTimeout is the timeout for HTTP requests (default: 30s).
	HTTPTimeout time.Duration `mapstructure:"http_timeout"`

	// TLSInsecureSkipVerify skips TLS certificate verification when true.
	TLSInsecureSkipVerify bool `mapstructure:"tls_insecure_skip_verify"`

	// QueryHypervisorMetrics controls whether Nova hypervisor metrics are collected (default: true).
	QueryHypervisorMetrics bool `mapstructure:"query_hypervisor_metrics"`
}

func (c *Config) Validate() error {
	if c.AuthURL == "" {
		return errors.New(`"auth_url" is required`)
	}
	if c.Username == "" {
		return errors.New(`"username" is required`)
	}
	if c.Password == "" {
		return errors.New(`"password" is required`)
	}
	return nil
}
