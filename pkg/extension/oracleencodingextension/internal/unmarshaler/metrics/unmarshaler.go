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

// Package metrics implements unmarshaling of OCI (Oracle Cloud
// Infrastructure) Monitoring metrics, published in JSONL format, into
// OpenTelemetry metrics.
package metrics // import "github.com/signalfx/splunk-otel-collector/pkg/extension/oracleencodingextension/internal/unmarshaler/metrics"

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"

	"go.opentelemetry.io/collector/pdata/pmetric"
	conventions "go.opentelemetry.io/otel/semconv/v1.40.0"
	"go.uber.org/zap"
)

// oracleCloudNamespaceKey and oracleCloudResourceGroupKey are not yet part of
// the OTel semantic conventions oracle_cloud.* registry, so they are defined here
const (
	oracleCloudCompartmentIDKey = "oracle_cloud.compartment_id"
	oracleCloudNamespaceKey     = "oracle_cloud.namespace"
	oracleCloudResourceGroupKey = "oracle_cloud.resource_group"

	// oracleCloudRealmKey is the OTel semantic convention key for the OCI
	// realm the resource's tenancy belongs to (e.g. "oc1", "oc2").
	// See https://opentelemetry.io/docs/specs/semconv/registry/attributes/oracle-cloud/
	oracleCloudRealmKey = "oracle_cloud.realm"

	// dimensionResourceID is the OCI Monitoring dimension holding the OCID of
	// the resource emitting the metric. It is additionally promoted to the
	// Resource as cloud.resource_id.
	dimensionResourceID = "resourceId"

	// ocidPrefix is the prefix of the first segment of every OCID, e.g.
	// "ocid1" for the current OCID version.
	// See https://docs.oracle.com/en-us/iaas/Content/General/Concepts/identifiers.htm
	ocidPrefix = "ocid"

	// ocidRealmSegmentIndex is the index, after splitting an OCID on ".", of
	// the realm segment: ocid1.<resource-type>.<realm>.<region>[.<future-use>].<unique-id>
	ocidRealmSegmentIndex = 2
)

// ScopeName is the instrumentation scope name set on metrics produced by
// this unmarshaler.
const ScopeName = "github.com/signalfx/splunk-otel-collector/pkg/extension/oracleencodingextension"

// ociMetricRecord represents a single OCI Monitoring metric record, one of
// which is expected per line of JSONL input.
type ociMetricRecord struct {
	Namespace     string               `json:"namespace"`
	CompartmentID string               `json:"compartmentId"`
	ResourceGroup string               `json:"resourceGroup"`
	Name          string               `json:"name"`
	Dimensions    map[string]any       `json:"dimensions"`
	Metadata      ociMetricMetadata    `json:"metadata"`
	Datapoints    []ociMetricDatapoint `json:"datapoints"`
}

type ociMetricMetadata struct {
	Unit        string `json:"unit"`
	DisplayName string `json:"displayName"`
}

type ociMetricDatapoint struct {
	// Timestamp is reported by OCI Monitoring as epoch milliseconds.
	Timestamp int64   `json:"timestamp"`
	Value     float64 `json:"value"`
}

type resourceIdentity struct {
	compartmentID string
	namespace     string
	resourceGroup string
	resourceID    string
}

type metricIdentity struct {
	resource resourceIdentity
	name     string
}

// ResourceMetricsUnmarshaler unmarshals OCI Monitoring metrics, encoded as
// JSONL, into pmetric.Metrics.
type ResourceMetricsUnmarshaler struct {
	logger *zap.Logger
}

// NewResourceMetricsUnmarshaler creates a new ResourceMetricsUnmarshaler.
func NewResourceMetricsUnmarshaler(logger *zap.Logger) ResourceMetricsUnmarshaler {
	return ResourceMetricsUnmarshaler{logger: logger}
}

// UnmarshalMetrics reads a JSONL-encoded payload of OCI metric records, one
// per line, and converts it into an OpenTelemetry pmetric.Metrics object.
func (r ResourceMetricsUnmarshaler) UnmarshalMetrics(buf []byte) (pmetric.Metrics, error) {
	b := newMetricsBuilder(r.logger)

	if err := readJSONLLines(bytes.NewReader(buf), b.unmarshalRecord); err != nil {
		return pmetric.NewMetrics(), fmt.Errorf("failed to read JSONL input: %w", err)
	}

	return b.build(), nil
}

// readJSONLLines reads newline-delimited records from r, invoking handle with
// each non-blank, trimmed line. It is split out from UnmarshalMetrics so that
// read errors from the underlying reader (as opposed to bytes.Reader, which
// only ever yields io.EOF) can be exercised in tests.
func readJSONLLines(r io.Reader, handle func([]byte)) error {
	reader := bufio.NewReader(r)
	for {
		line, err := reader.ReadBytes('\n')
		trimmed := bytes.TrimSpace(line)
		if len(trimmed) > 0 {
			handle(trimmed)
		}
		if err != nil {
			if err != io.EOF {
				return err
			}
			return nil
		}
	}
}

// resourceAttributes maps an OCI metric record onto OpenTelemetry resource
// attributes, following the cloud.* semantic conventions plus the
// oracle_cloud.* namespace for fields with no generic equivalent.
func resourceAttributes(rec ociMetricRecord, resourceID string) map[string]string {
	attrs := map[string]string{
		string(conventions.CloudProviderKey): conventions.CloudProviderOracleCloud.Value.AsString(),
	}
	if rec.CompartmentID != "" {
		attrs[oracleCloudCompartmentIDKey] = rec.CompartmentID
	}
	if resourceID != "" {
		attrs[string(conventions.CloudResourceIDKey)] = resourceID
	}
	if rec.Namespace != "" {
		attrs[oracleCloudNamespaceKey] = rec.Namespace
	}
	if rec.ResourceGroup != "" {
		attrs[oracleCloudResourceGroupKey] = rec.ResourceGroup
	}
	if realm := extractRealm(rec.CompartmentID); realm != "" {
		attrs[oracleCloudRealmKey] = realm
	}
	return attrs
}

// extractRealm parses the realm segment out of an OCID, e.g. "oc1" out of
// "ocid1.compartment.oc1..exampleuniqueID".
// See https://docs.oracle.com/en-us/iaas/Content/General/Concepts/identifiers.htm
func extractRealm(ocid string) string {
	segments := strings.Split(ocid, ".")
	if len(segments) <= ocidRealmSegmentIndex || !strings.HasPrefix(segments[0], ocidPrefix) {
		return ""
	}
	return segments[ocidRealmSegmentIndex]
}

// extractResourceID reads the resourceId dimension out of an OCI metric
// record's dimensions, since it identifies the monitored resource itself and
// is promoted onto the Resource as cloud.resource_id. It is also kept in the
// returned dimensions so it remains available as a per-datapoint attribute.
// OCI Monitoring has been observed emitting this dimension key as either
// "resourceId" or "resourceID", so the key is matched case-insensitively.
func extractResourceID(dimensions map[string]any) (resourceID string) {
	if len(dimensions) == 0 {
		return ""
	}

	for k, v := range dimensions {
		if resourceID == "" && strings.EqualFold(k, dimensionResourceID) {
			resourceID, _ = v.(string)
			break
		}
	}
	return resourceID
}
