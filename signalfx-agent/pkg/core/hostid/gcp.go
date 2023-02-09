// Copyright  Splunk, Inc.
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

package hostid

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/signalfx/signalfx-agent/pkg/utils/timeutil"
	log "github.com/sirupsen/logrus"
)

// GoogleComputeID generates a unique id for the compute instance that the
// agent is running on.  It returns a blank string if we are not running on GCP
// or there was an error getting the metadata.
func GoogleComputeID(cloudMetadataTimeout timeutil.Duration) string {
	projectID := getMetadata("project/project-id", cloudMetadataTimeout)
	if projectID == "" {
		return ""
	}
	instanceID := getMetadata("instance/id", cloudMetadataTimeout)
	if instanceID == "" {
		return ""
	}

	return fmt.Sprintf("%s_%s", projectID, instanceID)
}

func getMetadata(path string, cloudMetadataTimeout timeutil.Duration) string {
	url := fmt.Sprintf("http://metadata.google.internal/computeMetadata/v1/%s", path)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		// This would only be due to a programming bug
		panic("Could not construct request for GCP ID")
	}

	req.Header.Add("Metadata-Flavor", "Google")

	c := http.Client{
		Timeout: cloudMetadataTimeout.AsDuration(),
	}

	resp, err := c.Do(req)
	if err != nil {
		log.WithFields(log.Fields{
			"detail": err,
		}).Infof("No GCP metadata server detected at %s , assuming not on GCP", url)
		return ""
	}
	defer resp.Body.Close()

	if resp.Header.Get("Metadata-Flavor") != "Google" {
		return ""
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ""
	}

	return string(data)
}
