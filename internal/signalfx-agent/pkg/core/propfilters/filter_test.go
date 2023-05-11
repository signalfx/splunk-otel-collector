package propfilters

import (
	"testing"

	"github.com/signalfx/signalfx-agent/pkg/monitors/types"
	"github.com/stretchr/testify/assert"
)

func TestFilters(t *testing.T) {
	t.Run("Filter a single property name", func(t *testing.T) {
		f, _ := New([]string{"pod-template-hash"}, []string{"*"},
			[]string{"*"}, []string{"*"})

		properties := map[string]string{"pod-template-hash": "123", "replicaSet": "abc"}
		filteredProperties := f.FilterProperties(properties)

		expectedProperties := map[string]string{"replicaSet": "abc"}
		assert.Equal(t, filteredProperties, expectedProperties)
	})

	t.Run("Filter a regex property name", func(t *testing.T) {
		f, _ := New([]string{`/pod*/`}, []string{"*"},
			[]string{"*"}, []string{"*"})

		properties := map[string]string{"pod-template-hash": "123", "replicaSet": "abc"}
		filteredProperties := f.FilterProperties(properties)

		expectedProperties := map[string]string{"replicaSet": "abc"}
		assert.Equal(t, filteredProperties, expectedProperties)
	})

	t.Run("Filter a globbed property name", func(t *testing.T) {
		f, _ := New([]string{`pod*`}, []string{"*"},
			[]string{"*"}, []string{"*"})

		properties := map[string]string{"pod-template-hash": "123", "replicaSet": "abc"}
		filteredProperties := f.FilterProperties(properties)

		expectedProperties := map[string]string{"replicaSet": "abc"}
		assert.Equal(t, filteredProperties, expectedProperties)
	})

	t.Run("Filter a single property value", func(t *testing.T) {
		f, _ := New([]string{"*"}, []string{"123"},
			[]string{"*"}, []string{"*"})

		properties := map[string]string{"pod-template-hash": "123", "replicaSet": "abc"}
		filteredProperties := f.FilterProperties(properties)

		expectedProperties := map[string]string{"replicaSet": "abc"}
		assert.Equal(t, filteredProperties, expectedProperties)
	})

	t.Run("Filter a regex property value", func(t *testing.T) {
		f, _ := New([]string{"*"}, []string{`/12*/`},
			[]string{"*"}, []string{"*"})

		properties := map[string]string{"pod-template-hash": "123", "replicaSet": "abc"}
		filteredProperties := f.FilterProperties(properties)

		expectedProperties := map[string]string{"replicaSet": "abc"}
		assert.Equal(t, filteredProperties, expectedProperties)
	})

	t.Run("Filter a globbed property value", func(t *testing.T) {
		f, _ := New([]string{"*"}, []string{"12*"},
			[]string{"*"}, []string{"*"})

		properties := map[string]string{"pod-template-hash": "123", "replicaSet": "abc"}
		filteredProperties := f.FilterProperties(properties)

		expectedProperties := map[string]string{"replicaSet": "abc"}
		assert.Equal(t, filteredProperties, expectedProperties)
	})

	t.Run("Filter a property name and value", func(t *testing.T) {
		f, _ := New([]string{"pod-template-hash"}, []string{"abc"},
			[]string{"*"}, []string{"*"})

		properties := map[string]string{"pod-template-hash": "abc", "replicaSet": "abc"}
		filteredProperties := f.FilterProperties(properties)

		expectedProperties := map[string]string{"replicaSet": "abc"}
		assert.Equal(t, filteredProperties, expectedProperties)
	})

	t.Run("Filter a single dimension name", func(t *testing.T) {
		f, _ := New([]string{"*"}, []string{"*"},
			[]string{"kubernetes_pod_uid"}, []string{"*"})

		assert.True(t, f.MatchesDimension("kubernetes_pod_uid", "789"))
		assert.False(t, f.MatchesDimension("kubernetes_node_name", "789"))
	})

	t.Run("Filter a single dimension value", func(t *testing.T) {
		f, _ := New([]string{"*"}, []string{"*"},
			[]string{"*"}, []string{"789"})

		assert.True(t, f.MatchesDimension("kubernetes_pod_uid", "789"))
	})

	t.Run("Filter a regex dimension name", func(t *testing.T) {
		f, _ := New([]string{"*"}, []string{"*"},
			[]string{`/kubernetes_pod.*/`}, []string{"*"})

		assert.True(t, f.MatchesDimension("kubernetes_pod_uid", "789"))
		assert.False(t, f.MatchesDimension("kubernetes_node", "789"))
	})

	t.Run("Filter a regex dimension value", func(t *testing.T) {
		f, _ := New([]string{"*"}, []string{"*"},
			[]string{"*"}, []string{`/7.*/`})

		assert.True(t, f.MatchesDimension("kubernetes_pod_uid", "789"))
		assert.False(t, f.MatchesDimension("kubernetes_pod_uid", "456"))
	})

	t.Run("Filter a globbed dimension name", func(t *testing.T) {
		f, _ := New([]string{"*"}, []string{"*"},
			[]string{"kubernetes_pod*"}, []string{"*"})

		assert.True(t, f.MatchesDimension("kubernetes_pod_uid", "789"))
		assert.False(t, f.MatchesDimension("kubernetes_node", "789"))
	})

	t.Run("Filter a globbed dimension value", func(t *testing.T) {
		f, _ := New([]string{"*"}, []string{"*"},
			[]string{"*"}, []string{"7*"})

		assert.True(t, f.MatchesDimension("kubernetes_pod_uid", "789"))
		assert.False(t, f.MatchesDimension("kubernetes_pod_uid", "456"))
	})

	t.Run("Filter a dimension name and value", func(t *testing.T) {
		f, _ := New([]string{"*"}, []string{"*"},
			[]string{"kubernetes_pod_uid"}, []string{"789"})

		assert.True(t, f.MatchesDimension("kubernetes_pod_uid", "789"))
		assert.False(t, f.MatchesDimension("kubernetes_node", "789"))
		assert.False(t, f.MatchesDimension("kubernetes_pod_uid", "123"))
	})

	t.Run("Filter a dimension object given dimension name", func(t *testing.T) {
		f, _ := New([]string{"*"}, []string{"*"},
			[]string{"kubernetes_pod_uid"}, []string{"*"})

		properties := map[string]string{"pod-template-hash": "123", "replicaSet": "abc"}
		dim := &types.Dimension{
			Name:       "kubernetes_pod_uid",
			Value:      "789",
			Properties: properties,
			Tags:       nil,
		}

		filtered := f.FilterDimension(dim)
		assert.Nil(t, filtered)
	})

	t.Run("Filter a dimension object given property name", func(t *testing.T) {
		f, _ := New([]string{"pod-template-hash"}, []string{"*"},
			[]string{"*"}, []string{"*"})

		properties := map[string]string{"pod-template-hash": "123", "replicaSet": "abc"}
		dim := &types.Dimension{
			Name:       "kubernetes_pod_uid",
			Value:      "789",
			Properties: properties,
			Tags:       nil,
		}
		filteredDimension := f.FilterDimension(dim)

		expectedProperties := map[string]string{"replicaSet": "abc"}
		assert.Equal(t, filteredDimension.Properties, expectedProperties)
	})

	t.Run("Filter a dimension object given property value", func(t *testing.T) {
		f, _ := New([]string{"*"}, []string{"123"},
			[]string{"*"}, []string{"*"})

		properties := map[string]string{"pod-template-hash": "123", "replicaSet": "abc"}
		dim := &types.Dimension{
			Name:       "kubernetes_pod_uid",
			Value:      "789",
			Properties: properties,
			Tags:       nil,
		}

		filteredDimension := f.FilterDimension(dim)

		expectedProperties := map[string]string{"replicaSet": "abc"}
		assert.Equal(t, filteredDimension.Properties, expectedProperties)
	})

	t.Run("Filter a dimension object given property name and value", func(t *testing.T) {
		f, _ := New([]string{"pod-template-hash"}, []string{"123"},
			[]string{"*"}, []string{"*"})

		properties := map[string]string{"pod-template-hash": "123", "replicaSet": "abc", "service_uid": "123"}
		dim := &types.Dimension{
			Name:       "kubernetes_pod_uid",
			Value:      "789",
			Properties: properties,
			Tags:       nil,
		}
		filteredDimension := f.FilterDimension(dim)
		expectedProperties := map[string]string{"replicaSet": "abc", "service_uid": "123"}
		assert.Equal(t, filteredDimension.Properties, expectedProperties)
	})

	t.Run("Filter a dimension object given dimension value", func(t *testing.T) {
		f, _ := New([]string{"*"}, []string{"*"},
			[]string{"*"}, []string{"789"})

		properties := map[string]string{"pod-template-hash": "123", "replicaSet": "abc"}
		dim := &types.Dimension{
			Name:       "kubernetes_pod_uid",
			Value:      "789",
			Properties: properties,
			Tags:       nil,
		}

		filtered := f.FilterDimension(dim)
		assert.Nil(t, filtered)
	})

	t.Run("Filter a dimension object given dimension name and property name", func(t *testing.T) {
		f, _ := New([]string{"pod-template-hash"}, []string{"*"},
			[]string{"kubernetes_pod_uid"}, []string{"*"})

		properties := map[string]string{"pod-template-hash": "123", "replicaSet": "abc", "service_uid": "123"}
		nodeProperties := map[string]string{"pod-template-hash": "123", "replicaSet": "abc", "service_uid": "123"}
		dim := &types.Dimension{
			Name:       "kubernetes_pod_uid",
			Value:      "789",
			Properties: properties,
			Tags:       nil,
		}
		dimNode := &types.Dimension{
			Name:       "kubernetes_node",
			Value:      "minikube",
			Properties: nodeProperties,
			Tags:       nil,
		}
		filteredDimension := f.FilterDimension(dim)
		nodeFilteredDimension := f.FilterDimension(dimNode)
		expectedProperties := map[string]string{"replicaSet": "abc", "service_uid": "123"}
		assert.Equal(t, filteredDimension.Properties, expectedProperties)
		assert.Equal(t, nodeFilteredDimension.Properties, properties)
	})

	// negation tests
	t.Run("Filter a single negated property name", func(t *testing.T) {
		f, _ := New([]string{"!pod-template-hash"}, []string{"*"},
			[]string{"*"}, []string{"*"})

		properties := map[string]string{"pod-template-hash": "123", "replicaSet": "abc"}
		filteredProperties := f.FilterProperties(properties)

		expectedProperties := map[string]string{"pod-template-hash": "123"}
		assert.Equal(t, filteredProperties, expectedProperties)
	})

	t.Run("Filter a single negated property value", func(t *testing.T) {
		f, _ := New([]string{"*"}, []string{"!123"},
			[]string{"*"}, []string{"*"})

		properties := map[string]string{"pod-template-hash": "123", "replicaSet": "abc"}
		filteredProperties := f.FilterProperties(properties)

		expectedProperties := map[string]string{"pod-template-hash": "123"}
		assert.Equal(t, filteredProperties, expectedProperties)
	})

	t.Run("Match a negated dimension name", func(t *testing.T) {
		f, _ := New([]string{"*"}, []string{"*"},
			[]string{"!kubernetes_pod_uid"}, []string{"*"})

		dimensionName := "kubernetes_pod_uid"
		dimensionValue := "789"

		assert.False(t, f.MatchesDimension(dimensionName, dimensionValue))
		assert.True(t, f.MatchesDimension("kubernetes_node_name", dimensionValue))
	})

	t.Run("Match a negated dimension value", func(t *testing.T) {
		f, _ := New([]string{"*"}, []string{"*"},
			[]string{"*"}, []string{"!789"})

		dimensionName := "kubernetes_pod_uid"
		dimensionValue := "789"

		assert.False(t, f.MatchesDimension(dimensionName, dimensionValue))
		assert.True(t, f.MatchesDimension(dimensionName, "123"))
	})
}
