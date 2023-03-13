package kubernetes

import (
	"net/url"
	"os"
	"sync"
	"testing"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	"github.com/davecgh/go-spew/spew"
	"github.com/signalfx/signalfx-agent/pkg/core/common/kubernetes"
	"github.com/signalfx/signalfx-agent/pkg/core/services"
	"github.com/signalfx/signalfx-agent/pkg/observers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/signalfx/signalfx-agent/pkg/neotest/k8s/testhelpers/fakek8s"
)

var _ = Describe("Kubernetes Observer", func() {
	var config *Config
	var fakeK8s *FakeK8s
	var observer *Observer
	var lock sync.Mutex
	var endpoints map[services.ID]services.Endpoint

	BeforeEach(func() {
		config = &Config{}
		config.KubernetesAPI = &kubernetes.APIConfig{
			AuthType:   "none",
			SkipVerify: true,
		}
		config.DiscoverNodes = true

		fakeK8s = NewFakeK8s()
		fakeK8s.Start()
		K8sURL, _ := url.Parse(fakeK8s.URL())

		// The k8s golang library picks these up -- they are normally set
		// automatically by k8s in containers running in a real k8s env
		os.Setenv("KUBERNETES_SERVICE_HOST", K8sURL.Hostname())
		os.Setenv("KUBERNETES_SERVICE_PORT", K8sURL.Port())
	})

	startObserver := func() {
		lock.Lock()
		endpoints = make(map[services.ID]services.Endpoint)
		lock.Unlock()

		observer = &Observer{
			serviceCallbacks: &observers.ServiceCallbacks{
				Added: func(se services.Endpoint) {
					lock.Lock()
					endpoints[se.Core().ID] = se
					lock.Unlock()
				},
				Removed: func(se services.Endpoint) {
					lock.Lock()
					delete(endpoints, se.Core().ID)
					lock.Unlock()
				},
			},
			endpointsByUID: make(map[types.UID][]services.Endpoint),
		}

		err := observer.Configure(config)
		if err != nil {
			panic("K8s observer config failed")
		}
	}

	AfterEach(func() {
		fakeK8s.Close()
		observer.Shutdown()
	})

	It("Makes a port-less pod endpoint", func() {
		fakeK8s.SetInitialList([]runtime.Object{
			&v1.Pod{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Pod",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "test1",
					UID:  "abcdefghij",
				},
				Status: v1.PodStatus{
					Phase: v1.PodRunning,
					PodIP: "10.0.4.3",
					ContainerStatuses: []v1.ContainerStatus{
						{
							Name:         "container1",
							RestartCount: 5,
							State: v1.ContainerState{
								Running: &v1.ContainerStateRunning{},
							},
						},
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name: "container1",
						},
					},
				},
			},
		})

		startObserver()

		Eventually(func() int { lock.Lock(); defer lock.Unlock(); return len(endpoints) }).Should(Equal(1))
		lock.Lock()
		defer lock.Unlock()
		Expect(endpoints["test1-abcdefg-pod"].Core().Host).To(Equal("10.0.4.3"))
		Expect(endpoints["test1-abcdefg-pod"].Core().Port).To(Equal(uint16(0)))
		Expect(endpoints["test1-abcdefg-pod"].Core().DerivedFields()["pod_spec"].(*v1.PodSpec).Containers[0].Name).To(Equal("container1"))
	})

	It("Converts a pod to a set of endpoints", func() {
		fakeK8s.SetInitialList([]runtime.Object{
			&v1.Pod{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Pod",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "test1",
					UID:  "abcdefghij",
				},
				Status: v1.PodStatus{
					Phase: v1.PodRunning,
					PodIP: "10.0.4.3",
					ContainerStatuses: []v1.ContainerStatus{
						{
							Name:         "container1",
							RestartCount: 5,
							State: v1.ContainerState{
								Running: &v1.ContainerStateRunning{},
							},
						},
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name: "container1",
							Ports: []v1.ContainerPort{
								{
									Name:          "http",
									ContainerPort: 80,
								},
							},
						},
					},
				},
			},
		})

		startObserver()

		Eventually(func() int { lock.Lock(); defer lock.Unlock(); return len(endpoints) }).Should(Equal(2))
		lock.Lock()
		defer lock.Unlock()
		Expect(endpoints["test1-abcdefg-pod"].Core().Host).To(Equal("10.0.4.3"))
		Expect(endpoints["test1-abcdefg-pod"].Core().Port).To(Equal(uint16(0)))
		Expect(endpoints["test1-abcdefg-80"].Core().Host).To(Equal("10.0.4.3"))
		Expect(endpoints["test1-abcdefg-80"].Core().Port).To(Equal(uint16(80)))
	})

	It("Maps configuration from pod annotations", func() {
		fakeK8s.SetInitialList([]runtime.Object{
			&v1.Pod{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Pod",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test1",
					UID:       "abcdefghij",
					Namespace: "default",
					Annotations: map[string]string{
						"agent.signalfx.com/config.http.myVar":              "test123",
						"agent.signalfx.com/config.80.extraMetrics":         "true",
						"agent.signalfx.com/config.http.databases":          "[admin, db1]",
						"agent.signalfx.com/configFromEnv.http.username":    "USERNAME",
						"agent.signalfx.com/monitorType.http":               "mongo",
						"agent.signalfx.com/configFromSecret.http.password": "mongo/password",
						"agent.signalfx.com/config.https.myVar":             "abcde",
					},
				},
				Status: v1.PodStatus{
					Phase: v1.PodRunning,
					PodIP: "10.0.4.3",
					ContainerStatuses: []v1.ContainerStatus{
						{
							Name:         "container1",
							RestartCount: 5,
							State: v1.ContainerState{
								Running: &v1.ContainerStateRunning{},
							},
						},
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name: "container1",
							Env: []v1.EnvVar{
								{
									Name:  "USERNAME",
									Value: "bob123",
								},
							},
							Ports: []v1.ContainerPort{
								{
									Name:          "http",
									ContainerPort: 80,
								},
								{
									Name:          "https",
									ContainerPort: 443,
								},
							},
						},
					},
				},
			},
		})

		fakeK8s.CreateOrReplaceResource(&v1.Secret{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Secret",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "mongo",
				Namespace: "default",
			},
			Data: map[string][]byte{
				"password": []byte("s3cr3t"),
				"other":    []byte("other secret"),
			},
		})

		startObserver()

		Eventually(func() int { lock.Lock(); defer lock.Unlock(); return len(endpoints) }).Should(Equal(3))
		lock.Lock()
		defer lock.Unlock()
		Expect(endpoints["test1-abcdefg-443"].Core().MonitorType).To(Equal(""))
		Expect(endpoints["test1-abcdefg-443"].Core().Configuration["myVar"]).To(Equal("abcde"))
		Expect(endpoints["test1-abcdefg-80"].Core().MonitorType).To(Equal("mongo"))
		Expect(endpoints["test1-abcdefg-80"].Core().Configuration["myVar"]).To(Equal("test123"))
		Expect(endpoints["test1-abcdefg-80"].Core().Configuration["extraMetrics"]).To(Equal(true))
		Expect(endpoints["test1-abcdefg-80"].Core().Configuration["username"]).To(Equal("bob123"))
		Expect(endpoints["test1-abcdefg-80"].Core().Configuration["password"]).To(Equal("s3cr3t"))
		Expect(endpoints["test1-abcdefg-80"].Core().Configuration["databases"]).To(Equal([]interface{}{"admin", "db1"}))
	})

	It("Updates endpoints for changed and deleted pods", func() {
		fakeK8s.SetInitialList([]runtime.Object{
			&v1.Pod{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Pod",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test1",
					Namespace: "default",
					UID:       "abcdefghij",
				},
				Status: v1.PodStatus{
					Phase: v1.PodRunning,
					PodIP: "10.0.4.3",
					ContainerStatuses: []v1.ContainerStatus{
						{
							Name:         "container1",
							RestartCount: 5,
							State: v1.ContainerState{
								Running: &v1.ContainerStateRunning{},
							},
						},
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name: "container1",
							Ports: []v1.ContainerPort{
								{
									Name:          "http",
									ContainerPort: 80,
								},
							},
						},
					},
				},
			},
		})

		startObserver()

		Eventually(func() int { lock.Lock(); defer lock.Unlock(); return len(endpoints) }, 3*time.Second).Should(Equal(2))
		lock.Lock()
		Expect(endpoints["test1-abcdefg-pod"].Core().Host).To(Equal("10.0.4.3"))
		Expect(endpoints["test1-abcdefg-pod"].Core().Port).To(Equal(uint16(0)))
		Expect(endpoints["test1-abcdefg-80"].Core().Host).To(Equal("10.0.4.3"))
		Expect(endpoints["test1-abcdefg-80"].Core().Port).To(Equal(uint16(80)))
		lock.Unlock()

		fakeK8s.CreateOrReplaceResource(&v1.Pod{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Pod",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test1",
				Namespace: "default",
				UID:       "abcdefghij",
			},
			Status: v1.PodStatus{
				Phase: v1.PodRunning,
				PodIP: "10.0.4.5",
				ContainerStatuses: []v1.ContainerStatus{
					{
						Name:         "container1",
						RestartCount: 5,
						State: v1.ContainerState{
							Running: &v1.ContainerStateRunning{},
						},
					},
				},
			},
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					{
						Name: "container1",
						Ports: []v1.ContainerPort{
							{
								Name:          "http",
								ContainerPort: 80,
							},
							{
								Name:          "http-alt",
								ContainerPort: 8080,
							},
						},
					},
				},
			},
		})

		Eventually(func() int { lock.Lock(); defer lock.Unlock(); spew.Dump(endpoints); return len(endpoints) }, 3*time.Second).Should(Equal(3))
		lock.Lock()
		Expect(endpoints["test1-abcdefg-pod"].Core().Host).To(Equal("10.0.4.5"))
		Expect(endpoints["test1-abcdefg-pod"].Core().Port).To(Equal(uint16(0)))
		Expect(endpoints["test1-abcdefg-80"].Core().Host).To(Equal("10.0.4.5"))
		Expect(endpoints["test1-abcdefg-80"].Core().Port).To(Equal(uint16(80)))
		Expect(endpoints["test1-abcdefg-8080"].Core().Host).To(Equal("10.0.4.5"))
		Expect(endpoints["test1-abcdefg-8080"].Core().Port).To(Equal(uint16(8080)))
		lock.Unlock()

		fakeK8s.DeleteResourceByName("Pod", "default", "test1")
		Eventually(func() int { lock.Lock(); defer lock.Unlock(); return len(endpoints) }).Should(Equal(0))
	})

	It("Emits endpoints for nodes", func() {
		fakeK8s.SetInitialList([]runtime.Object{
			&v1.Node{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Node",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "node1",
					UID:  "node-1234",
				},
				Status: v1.NodeStatus{
					Addresses: []v1.NodeAddress{
						{
							Type:    v1.NodeExternalIP,
							Address: "1.2.3.4",
						},
						{
							Type:    v1.NodeInternalIP,
							Address: "10.5.4.3",
						},
						{
							Type:    v1.NodeInternalDNS,
							Address: "node1.internal",
						},
					},
				},
				Spec: v1.NodeSpec{},
			},
		})

		startObserver()

		Eventually(func() int { lock.Lock(); defer lock.Unlock(); return len(endpoints) }).Should(Equal(1))
		lock.Lock()
		Expect(endpoints["node-node1-node-12"].Core().Host).To(Equal("10.5.4.3"))
		Expect(endpoints["node-node1-node-12"].Core().DerivedFields()["node_addresses"].(map[v1.NodeAddressType]string)[v1.NodeInternalDNS]).To(Equal("node1.internal"))
		Expect(endpoints["node-node1-node-12"].Core().Target).To(Equal(services.TargetTypeKubernetesNode))
		lock.Unlock()

		fakeK8s.DeleteResourceByName("Node", "", "node1")
		Eventually(func() int { lock.Lock(); defer lock.Unlock(); return len(endpoints) }, 3*time.Second).Should(Equal(0))
	})

})

func TestKubernetes(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Kubernetes Observer Suite")
}
