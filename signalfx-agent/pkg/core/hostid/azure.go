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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/signalfx/signalfx-agent/pkg/utils/timeutil"
	log "github.com/sirupsen/logrus"
)

// AzureUniqueID constructs the unique ID of the underlying Azure VM.  If
// not running on Azure VM, returns the empty string.
// Details about Azure Instance Metadata ednpoint:
// https://docs.microsoft.com/en-us/azure/virtual-machines/windows/instance-metadata-service
func AzureUniqueID(cloudMetadataTimeout timeutil.Duration) string {
	c := http.Client{
		Timeout: cloudMetadataTimeout.AsDuration(),
	}
	req, err := http.NewRequest("GET", "http://169.254.169.254/metadata/instance?api-version=2018-10-01", nil)
	if err != nil {
		return ""
	}

	req.Header.Set("Metadata", "true")
	resp, err := c.Do(req)
	if err != nil {
		log.WithFields(log.Fields{
			"detail": err,
		}).Infof("No Azure metadata server detected, assuming not on Azure")
		return ""
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ""
	}

	type Info struct {
		SubscriptionID    string `json:"subscriptionId"`
		ResourceGroupName string `json:"resourceGroupName"`
		Name              string `json:"name"`
		VMScaleSetName    string `json:"vmScaleSetName"`
	}

	var compute struct {
		Doc Info `json:"compute"`
	}

	err = json.Unmarshal(body, &compute)
	if err != nil {
		return ""
	}

	if compute.Doc.SubscriptionID == "" || compute.Doc.ResourceGroupName == "" || compute.Doc.Name == "" {
		return ""
	}

	if compute.Doc.VMScaleSetName == "" {
		return strings.ToLower(fmt.Sprintf("%s/%s/microsoft.compute/virtualmachines/%s", compute.Doc.SubscriptionID, compute.Doc.ResourceGroupName, compute.Doc.Name))
	}

	instanceID := strings.TrimPrefix(compute.Doc.Name, compute.Doc.VMScaleSetName+"_")

	// names of VM's in VMScalesets seem to follow the of `<scale-set-name>_<instance-id>`
	// where scale-set-name is alphanumeric (and is the same as compute.vmScaleSetName
	// field from the metadata endpoint)
	if instanceID == "" {
		return ""
	}

	return strings.ToLower(fmt.Sprintf("%s/%s/microsoft.compute/virtualmachinescalesets/%s/virtualmachines/%s", compute.Doc.SubscriptionID, compute.Doc.ResourceGroupName, compute.Doc.VMScaleSetName, instanceID))

}
