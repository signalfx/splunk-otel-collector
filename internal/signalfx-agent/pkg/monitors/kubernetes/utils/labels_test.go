package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPropsAndTagsFromLabels(t *testing.T) {
	tests := []struct {
		name            string
		labels          map[string]string
		sendUnsanitized bool
		expectedProps   map[string]string
		expectedTags    map[string]bool
	}{
		{
			name: "sanitizes properties by default",
			labels: map[string]string{
				"app.kubernetes.io/name":    "myapp",
				"app.kubernetes.io/version": "1.0.0",
				"kubernetes.io/cluster":     "prod",
			},
			sendUnsanitized: false,
			expectedProps: map[string]string{
				"app_kubernetes_io_name":    "myapp",
				"app_kubernetes_io_version": "1.0.0",
				"kubernetes_io_cluster":     "prod",
			},
			expectedTags: map[string]bool{},
		},
		{
			name: "sends both sanitized and unsanitized when enabled",
			labels: map[string]string{
				"app.kubernetes.io/name":    "myapp",
				"app.kubernetes.io/version": "1.0.0",
			},
			sendUnsanitized: true,
			expectedProps: map[string]string{
				"app_kubernetes_io_name":    "myapp",
				"app_kubernetes_io_version": "1.0.0",
				"app.kubernetes.io/name":    "myapp",
				"app.kubernetes.io/version": "1.0.0",
			},
			expectedTags: map[string]bool{},
		},
		{
			name: "handles forward slashes",
			labels: map[string]string{
				"example.com/role": "frontend",
			},
			sendUnsanitized: true,
			expectedProps: map[string]string{
				"example_com_role": "frontend",
				"example.com/role": "frontend",
			},
			expectedTags: map[string]bool{},
		},
		{
			name: "handles empty value labels as tags with sanitization",
			labels: map[string]string{
				"tag.with.dots": "",
				"regular-tag":   "",
			},
			sendUnsanitized: false,
			expectedProps:   map[string]string{},
			expectedTags: map[string]bool{
				"tag_with_dots": true,
				"regular-tag":   true,
			},
		},
		{
			name: "handles empty value labels as tags with both sanitized and unsanitized",
			labels: map[string]string{
				"tag.with.dots": "",
			},
			sendUnsanitized: true,
			expectedProps:   map[string]string{},
			expectedTags: map[string]bool{
				"tag_with_dots": true,
				"tag.with.dots": true,
			},
		},
		{
			name: "does not duplicate when no special characters",
			labels: map[string]string{
				"simple":      "value",
				"another-one": "value2",
			},
			sendUnsanitized: true,
			expectedProps: map[string]string{
				"simple":      "value",
				"another-one": "value2",
			},
			expectedTags: map[string]bool{},
		},
		{
			name: "handles mixed dots and slashes",
			labels: map[string]string{
				"app.kubernetes.io/component": "database",
				"company.com/team":            "platform",
			},
			sendUnsanitized: false,
			expectedProps: map[string]string{
				"app_kubernetes_io_component": "database",
				"company_com_team":            "platform",
			},
			expectedTags: map[string]bool{},
		},
		{
			name: "handles mixed dots and slashes with unsanitized",
			labels: map[string]string{
				"app.kubernetes.io/component": "database",
			},
			sendUnsanitized: true,
			expectedProps: map[string]string{
				"app_kubernetes_io_component": "database",
				"app.kubernetes.io/component": "database",
			},
			expectedTags: map[string]bool{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			props, tags := PropsAndTagsFromLabels(tt.labels, tt.sendUnsanitized)
			assert.Equal(t, tt.expectedProps, props)
			assert.Equal(t, tt.expectedTags, tags)
		})
	}
}
