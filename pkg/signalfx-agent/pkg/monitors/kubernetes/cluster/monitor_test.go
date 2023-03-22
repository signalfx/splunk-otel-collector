package cluster

import (
	"fmt"
	"net/url"
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/signalfx/golib/v3/datapoint"
	log "github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/signalfx/signalfx-agent/pkg/core/common/kubernetes"
	"github.com/signalfx/signalfx-agent/pkg/monitors/types"
	"github.com/signalfx/signalfx-agent/pkg/neotest"
	. "github.com/signalfx/signalfx-agent/pkg/neotest/k8s/testhelpers/fakek8s"
)

var _ = Describe("Kubernetes plugin", func() {
	var config *Config
	var fakeK8s *FakeK8s
	var monitor *Monitor
	var output *neotest.TestOutput

	BeforeEach(func() {
		config = &Config{}
		config.IntervalSeconds = 1
		config.KubernetesAPI = &kubernetes.APIConfig{
			AuthType:   "none",
			SkipVerify: true,
		}

		fakeK8s = NewFakeK8s()
		fakeK8s.Start()
		K8sURL, _ := url.Parse(fakeK8s.URL())

		output = neotest.NewTestOutput()

		// The k8s go library picks these up -- they are set automatically in
		// containers running in a real k8s env
		os.Setenv("KUBERNETES_SERVICE_HOST", K8sURL.Hostname())
		os.Setenv("KUBERNETES_SERVICE_PORT", K8sURL.Port())

	})

	doSetup := func(alwaysClusterReporter bool, thisPodName string) {
		config.AlwaysClusterReporter = alwaysClusterReporter
		os.Setenv("MY_POD_NAME", thisPodName)

		os.Setenv("SFX_ACCESS_TOKEN", "deadbeef")

		monitor = &Monitor{}
		monitor.Output = output

		err := monitor.Configure(config)
		if err != nil {
			panic("K8s monitor config failed")
		}
	}

	AfterEach(func() {
		monitor.Shutdown()
		fakeK8s.Close()
	})

	// Making an int literal pointer requires a helper
	intp := func(n int32) *int32 { return &n }
	intValue := func(v datapoint.Value) int64 {
		return v.(datapoint.IntValue).Int()
	}

	waitForDatapoints := func(expected int) []*datapoint.Datapoint {
		dps := output.WaitForDPs(expected, 3)
		Expect(len(dps)).Should(BeNumerically(">=", expected))
		return dps
	}

	expectIntMetric := func(dps []*datapoint.Datapoint, uidField, objUid string, metricName string, metricValue int) {
		matched := false
		for _, dp := range dps {
			dims := dp.Dimensions
			if dp.Metric == metricName && dims[uidField] == objUid {
				Expect(intValue(dp.Value)).To(Equal(int64(metricValue)), fmt.Sprintf("%s %s", objUid, metricName))
				matched = true
			}
		}
		Expect(matched).To(Equal(true), fmt.Sprintf("%s %s %d", objUid, metricName, metricValue))
	}

	expectIntMetricMissing := func(dps []*datapoint.Datapoint, uidField, objUid string, metricName string) {
		matched := false
		for _, dp := range dps {
			dims := dp.Dimensions
			if dp.Metric == metricName && dims[uidField] == objUid {
				matched = true
			}
		}
		Expect(matched).To(Equal(false), fmt.Sprintf("%s %s", objUid, metricName))
	}

	It("Sends pod phase metrics", func() {
		log.SetLevel(log.DebugLevel)
		fakeK8s.SetInitialList([]runtime.Object{
			&v1.Pod{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Pod",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test1",
					UID:       "abcd",
					Namespace: "default",
					Labels: map[string]string{
						"env": "test",
					},
					OwnerReferences: []metav1.OwnerReference{
						{
							Kind: "DaemonSet",
							Name: "MySet",
						},
					},
				},
				Status: v1.PodStatus{
					Phase: v1.PodRunning,
					ContainerStatuses: []v1.ContainerStatus{
						{
							ContainerID:  "c1",
							Ready:        true,
							Name:         "container1",
							RestartCount: 5,
							State: v1.ContainerState{
								Running: &v1.ContainerStateRunning{},
							},
						},
						{
							ContainerID:  "",
							Ready:        true,
							Name:         "container2",
							RestartCount: 5,
							State: v1.ContainerState{
								Running: &v1.ContainerStateRunning{},
							},
						},
					},
				},
			},
		})

		doSetup(true, "")

		dps := waitForDatapoints(3)

		Expect(dps[0].Metric).To(Equal("kubernetes.pod_phase"))
		Expect(intValue(dps[0].Value)).To(Equal(int64(2)))
		Expect(dps[1].Metric).To(Equal("kubernetes.container_restart_count"))
		Expect(intValue(dps[1].Value)).To(Equal(int64(5)))
		Expect(dps[2].Metric).To(Equal("kubernetes.container_ready"))
		Expect(intValue(dps[2].Value)).To(Equal(int64(1)))

		dims := output.WaitForDimensions(2, 3)
		Expect(len(dims)).Should(Equal(2))

		Expect(dims).Should(ConsistOf(&types.Dimension{
			Name:  "kubernetes_pod_uid",
			Value: "abcd",
			Properties: map[string]string{
				"pod_creation_timestamp":   "0001-01-01T00:00:00Z",
				"kubernetes_workload":      "DaemonSet",
				"kubernetes_workload_name": "MySet",
				"daemonSet":                "MySet",
				"daemonSet_uid":            "",
				"env":                      "test",
			},
			Tags:              map[string]bool{},
			MergeIntoExisting: true,
		}, &types.Dimension{
			Name:  "container_id",
			Value: "c1",
			Properties: map[string]string{
				"container_status":        "running",
				"container_status_reason": "",
			},
			Tags:              nil,
			MergeIntoExisting: true,
		}))

		firstDim := dps[0].Dimensions
		Expect(firstDim["metric_source"]).To(Equal("kubernetes"))

		fakeK8s.CreateOrReplaceResource(&v1.Pod{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Pod",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pod2",
				UID:       "1234",
				Namespace: "default",
				Labels: map[string]string{
					"env": "prod",
				},
			},
			Status: v1.PodStatus{
				Phase: v1.PodFailed,
				ContainerStatuses: []v1.ContainerStatus{
					{
						ContainerID:  "c2",
						Name:         "container2",
						RestartCount: 0,
						State: v1.ContainerState{
							Running: &v1.ContainerStateRunning{},
						},
					},
				},
			},
		})

		_ = waitForDatapoints(6)
		dps = waitForDatapoints(6)
		expectIntMetric(dps, "kubernetes_pod_uid", "1234", "kubernetes.container_restart_count", 0)

		dims = output.WaitForDimensions(2, 3)
		Expect(len(dims)).Should(Equal(2))

		Expect(dims).Should(ConsistOf(&types.Dimension{
			Name:  "kubernetes_pod_uid",
			Value: "1234",
			Properties: map[string]string{
				"pod_creation_timestamp": "0001-01-01T00:00:00Z",
				"env":                    "prod",
			},
			Tags:              map[string]bool{},
			MergeIntoExisting: true,
		}, &types.Dimension{
			Name:  "container_id",
			Value: "c2",
			Properties: map[string]string{
				"container_status":        "running",
				"container_status_reason": "",
			},
			Tags:              nil,
			MergeIntoExisting: true,
		}))

		fakeK8s.CreateOrReplaceResource(&v1.Pod{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Pod",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pod2",
				UID:       "1234",
				Namespace: "default",
				Labels: map[string]string{
					"env": "qa",
				},
			},
			Status: v1.PodStatus{
				Phase: v1.PodFailed,
				ContainerStatuses: []v1.ContainerStatus{
					{
						ContainerID:  "c2",
						Name:         "container2",
						RestartCount: 2,
					},
				},
			},
		})

		_ = waitForDatapoints(6)
		dps = waitForDatapoints(6)
		expectIntMetric(dps, "kubernetes_pod_uid", "1234", "kubernetes.container_restart_count", 2)

		dims = output.WaitForDimensions(2, 3)

		Expect(dims).Should(ConsistOf(&types.Dimension{
			Name:  "kubernetes_pod_uid",
			Value: "1234",
			Properties: map[string]string{
				"pod_creation_timestamp": "0001-01-01T00:00:00Z",
				"env":                    "qa",
			},
			Tags:              map[string]bool{},
			MergeIntoExisting: true,
		}, &types.Dimension{
			Name:              "container_id",
			Value:             "c2",
			Properties:        map[string]string{},
			Tags:              nil,
			MergeIntoExisting: true,
		}))

		fakeK8s.CreateOrReplaceResource(&v1.Pod{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Pod",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pod2",
				UID:       "1234",
				Namespace: "default",
				Labels: map[string]string{
					"env": "qa",
				},
			},
			Status: v1.PodStatus{
				Phase: v1.PodFailed,
				ContainerStatuses: []v1.ContainerStatus{
					{
						ContainerID:  "c2",
						Name:         "container2",
						RestartCount: 3,
					},
				},
			},
		})

		_ = waitForDatapoints(6)
		dps = waitForDatapoints(6)
		expectIntMetric(dps, "kubernetes_pod_uid", "1234", "kubernetes.container_restart_count", 3)

		fakeK8s.DeleteResource(&v1.Pod{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Pod",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pod2",
				UID:       "1234",
				Namespace: "default",
			},
			Status: v1.PodStatus{
				Phase: v1.PodFailed,
				ContainerStatuses: []v1.ContainerStatus{
					{
						ContainerID:  "container_id",
						Name:         "container2",
						RestartCount: 2,
					},
				},
			},
		})

		// Throw away the next set of dps since they could still have the pod
		// metrics if sent before the update but after the previous assertion.
		_ = waitForDatapoints(6)
		dps = waitForDatapoints(6)

		expectIntMetricMissing(dps, "kubernetes_pod_uid", "1234", "kubernetes.container_restart_count")

	}, 5)

	It("Sends Deployment metrics", func() {
		fakeK8s.SetInitialList([]runtime.Object{
			&v1.Pod{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Pod",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test1",
					UID:       "1234",
					Namespace: "default",
				},
				Status: v1.PodStatus{
					Phase: v1.PodRunning,
				},
			},
			&appsv1.Deployment{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Deployment",
					APIVersion: "apps/v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "deploy1",
					UID:       "abcd",
					Namespace: "default",
				},
				Spec: appsv1.DeploymentSpec{
					Replicas: intp(int32(10)),
				},
				Status: appsv1.DeploymentStatus{
					AvailableReplicas: 5,
				},
			},
			&appsv1.Deployment{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Deployment",
					APIVersion: "apps/v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "deploy2",
					UID:       "efgh",
					Namespace: "default",
				},
				Spec: appsv1.DeploymentSpec{
					Replicas: intp(int32(1)),
				},
				Status: appsv1.DeploymentStatus{
					AvailableReplicas: 1,
					UpdatedReplicas:   1,
				},
			},
		})

		doSetup(true, "")

		dps := waitForDatapoints(8)

		By("Reporting on existing deployments")
		expectIntMetric(dps, "kubernetes_uid", "abcd", "kubernetes.deployment.desired", 10)
		expectIntMetric(dps, "kubernetes_uid", "abcd", "kubernetes.deployment.available", 5)
		expectIntMetric(dps, "kubernetes_uid", "efgh", "kubernetes.deployment.desired", 1)
		expectIntMetric(dps, "kubernetes_uid", "efgh", "kubernetes.deployment.available", 1)
		expectIntMetric(dps, "kubernetes_uid", "efgh", "kubernetes.deployment.updated", 1)

		fakeK8s.CreateOrReplaceResource(&appsv1.Deployment{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Deployment",
				APIVersion: "apps/v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "deploy2",
				UID:       "efgh",
				Namespace: "default",
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: intp(int32(1)),
			},
			Status: appsv1.DeploymentStatus{
				AvailableReplicas: 0,
			},
		})

		_ = waitForDatapoints(8)
		dps = waitForDatapoints(8)
		By("Responding to events pushed on the watch API")
		expectIntMetric(dps, "kubernetes_uid", "abcd", "kubernetes.deployment.desired", 10)
		expectIntMetric(dps, "kubernetes_uid", "abcd", "kubernetes.deployment.available", 5)
		expectIntMetric(dps, "kubernetes_uid", "efgh", "kubernetes.deployment.desired", 1)
		expectIntMetric(dps, "kubernetes_uid", "efgh", "kubernetes.deployment.available", 0)
		expectIntMetric(dps, "kubernetes_uid", "efgh", "kubernetes.deployment.updated", 0)
	})

})

func TestKubernetes(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Kubernetes Monitor Suite")
}
