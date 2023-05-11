package monitors

import (
	"strconv"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"

	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/core/meta"
	"github.com/signalfx/signalfx-agent/pkg/core/services"
	"github.com/signalfx/signalfx-agent/pkg/monitors/types"
)

// Used to make unique service ids
// nolint: gochecknoglobals
var serviceID = 0

// nolint: unparam
func newService(imageName string, publicPort int) services.Endpoint {
	serviceID++

	endpoint := services.NewEndpointCore(strconv.Itoa(serviceID), "", "test", nil)
	endpoint.Host = "example.com"
	endpoint.Port = uint16(publicPort)

	return &services.ContainerEndpoint{
		EndpointCore:  *endpoint,
		AltPort:       0,
		Container:     services.Container{Image: imageName},
		Orchestration: services.Orchestration{},
	}
}

var _ = Describe("Monitor Manager", func() {
	var manager *MonitorManager
	var getMonitors func() map[types.MonitorID]MockMonitor

	BeforeEach(func() {
		DeregisterAll()

		getMonitors = RegisterFakeMonitors()

		manager = NewMonitorManager(&meta.AgentMeta{})
	})

	It("Starts up static monitors immediately", func() {
		manager.Configure([]config.MonitorConfig{
			{
				Type: "static1",
			},
			{
				Type:          "dynamic1",
				DiscoveryRule: `container_image =~ "my-service"`,
			},
		}, &config.CollectdConfig{}, 10)

		Expect(len(getMonitors())).To(Equal(1))
		for _, mon := range getMonitors() {
			Expect(mon.Type()).To(Equal("static1"))
		}
	})

	It("Shuts down static monitors when removed from config", func() {
		manager.Configure([]config.MonitorConfig{
			{
				Type: "static1",
			},
			{
				Type:          "dynamic1",
				DiscoveryRule: `container_image =~ "my-service"`,
			},
		}, &config.CollectdConfig{}, 10)

		Expect(len(getMonitors())).To(Equal(1))

		manager.Configure([]config.MonitorConfig{
			{
				Type:          "dynamic1",
				DiscoveryRule: `container_image =~ "my-service"`,
			},
		}, &config.CollectdConfig{}, 10)

		Expect(len(getMonitors())).To(Equal(0))
	})

	It("Starts up dynamic monitors upon service discovery", func() {
		manager.Configure([]config.MonitorConfig{
			{
				Type: "static1",
			},
			{
				Type:          "dynamic1",
				DiscoveryRule: `container_image =~ "my-service"`,
			},
		}, &config.CollectdConfig{}, 10)

		Expect(len(getMonitors())).To(Equal(1))

		manager.EndpointAdded(newService("my-service", 5000))

		Expect(len(getMonitors())).To(Equal(2))

		mons := findMonitorsByType(getMonitors(), "dynamic1")
		Expect(len(mons)).To(Equal(1))
	})

	It("Shuts down dynamic monitors upon service removed", func() {
		manager.Configure([]config.MonitorConfig{
			{
				Type: "static1",
			},
			{
				Type:          "dynamic1",
				DiscoveryRule: `container_image =~ "my-service"`,
			},
		}, &config.CollectdConfig{}, 10)

		service := newService("my-service", 5000)
		manager.EndpointAdded(service)

		mons := findMonitorsByType(getMonitors(), "dynamic1")
		Expect(len(mons)).To(Equal(1))

		shutdownCallCount := 0
		for i := range mons {
			mons[i].AddShutdownHook(func() {
				shutdownCallCount++
			})
		}

		manager.EndpointRemoved(service)

		mons = findMonitorsByType(getMonitors(), "dynamic1")
		Expect(len(mons)).To(Equal(0))
		Expect(shutdownCallCount).To(Equal(1))
	})

	It("Starts and stops multiple monitors for multiple endpoints of same monitor type", func() {
		manager.Configure([]config.MonitorConfig{
			{
				Type: "static1",
			},
			{
				Type:          "dynamic1",
				DiscoveryRule: `container_image =~ "my-service"`,
			},
		}, &config.CollectdConfig{}, 10)

		service := newService("my-service", 5000)
		service2 := newService("my-service", 5001)
		manager.EndpointAdded(service)
		manager.EndpointAdded(service2)

		mons := findMonitorsByType(getMonitors(), "dynamic1")
		Expect(len(mons)).To(Equal(2))

		shutdownCallCount := 0
		for i := range mons {
			mons[i].AddShutdownHook(func() {
				shutdownCallCount++
			})
		}

		manager.EndpointRemoved(service)

		mons = findMonitorsByType(getMonitors(), "dynamic1")
		Expect(len(mons)).To(Equal(1))
		Expect(shutdownCallCount).To(Equal(1))

		manager.EndpointRemoved(service2)

		mons = findMonitorsByType(getMonitors(), "dynamic1")
		Expect(len(mons)).To(Equal(0))
		Expect(shutdownCallCount).To(Equal(2))
	})

	It("Re-monitors service if monitor is removed temporarily", func() {
		log.SetLevel(log.DebugLevel)
		goodConfig := []config.MonitorConfig{
			{
				Type: "static1",
			},
			{
				Type:          "dynamic1",
				DiscoveryRule: `container_image =~ "my-service"`,
			},
		}
		manager.Configure(goodConfig, &config.CollectdConfig{}, 10)

		manager.EndpointAdded(newService("my-service", 5000))

		mons := findMonitorsByType(getMonitors(), "dynamic1")
		Expect(len(mons)).To(Equal(1))

		manager.Configure([]config.MonitorConfig{
			{
				Type: "static1",
			},
			{
				Type:          "dynamic1",
				DiscoveryRule: `container_image =~ "my-service"`,
				OtherConfig:   map[string]interface{}{"invalid": true},
			},
		}, &config.CollectdConfig{}, 10)

		mons = findMonitorsByType(getMonitors(), "dynamic1")
		Expect(len(mons)).To(Equal(0))

		manager.Configure(goodConfig, &config.CollectdConfig{}, 10)

		mons = findMonitorsByType(getMonitors(), "dynamic1")
		Expect(len(mons)).To(Equal(1))
	})

	It("Starts monitoring previously discovered service if new monitor config matches", func() {
		manager.Configure([]config.MonitorConfig{
			{
				Type:          "dynamic1",
				DiscoveryRule: `container_image =~ "their-service"`,
			},
		}, &config.CollectdConfig{}, 10)

		manager.EndpointAdded(newService("my-service", 5000))

		mons := findMonitorsByType(getMonitors(), "dynamic1")
		Expect(len(mons)).To(Equal(0))

		manager.Configure([]config.MonitorConfig{
			{
				Type:          "dynamic1",
				DiscoveryRule: `container_image =~ "my-service"`,
			},
		}, &config.CollectdConfig{}, 10)

		mons = findMonitorsByType(getMonitors(), "dynamic1")
		Expect(len(mons)).To(Equal(1))
	})

	It("Stops monitoring service if new monitor config no longer matches", func() {
		manager.Configure([]config.MonitorConfig{
			{
				Type:          "dynamic1",
				DiscoveryRule: `container_image =~ "my-service"`,
			},
		}, &config.CollectdConfig{}, 10)

		manager.EndpointAdded(newService("my-service", 5000))

		mons := findMonitorsByType(getMonitors(), "dynamic1")
		Expect(len(mons)).To(Equal(1))

		manager.Configure([]config.MonitorConfig{
			{
				Type:          "dynamic1",
				DiscoveryRule: `container_image =~ "their-service"`,
			},
		}, &config.CollectdConfig{}, 10)

		mons = findMonitorsByType(getMonitors(), "dynamic1")
		Expect(len(mons)).To(Equal(0))
	})

	It("Monitors the same service on multiple monitors", func() {
		manager.Configure([]config.MonitorConfig{
			{
				Type: "static1",
			},
			{
				Type:          "dynamic1",
				DiscoveryRule: `container_image =~ "my-service"`,
			},
		}, &config.CollectdConfig{}, 10)

		manager.EndpointAdded(newService("my-service", 5000))

		mons := findMonitorsByType(getMonitors(), "dynamic1")
		Expect(len(mons)).To(Equal(1))

		manager.Configure([]config.MonitorConfig{
			{
				Type: "static1",
			},
			{
				Type:          "dynamic1",
				DiscoveryRule: `container_image =~ "my-service"`,
			},
			{
				Type:          "dynamic2",
				DiscoveryRule: `container_image =~ "my-service"`,
			},
		}, &config.CollectdConfig{}, 10)

		mons = findMonitorsByType(getMonitors(), "dynamic1")
		Expect(len(mons)).To(Equal(1))

		mons = findMonitorsByType(getMonitors(), "dynamic2")
		Expect(len(mons)).To(Equal(1))

		// Test restarting and making sure it sticks
		manager.Configure([]config.MonitorConfig{
			{
				Type:          "dynamic1",
				DiscoveryRule: `container_image =~ "my-service"`,
			},
			{
				Type:          "dynamic2",
				DiscoveryRule: `container_image =~ "my-service"`,
			},
		}, &config.CollectdConfig{}, 10)

		mons = findMonitorsByType(getMonitors(), "dynamic1")
		Expect(len(mons)).To(Equal(1))

		mons = findMonitorsByType(getMonitors(), "dynamic2")
		Expect(len(mons)).To(Equal(1))
	})

	It("Validates required fields", func() {
		manager.Configure([]config.MonitorConfig{
			{
				Type: "static1",
			},
			{
				Type: "dynamic2",
				OtherConfig: map[string]interface{}{
					"host": "example.com",
					// Port is missing but required
				},
			},
		}, &config.CollectdConfig{}, 10)

		mons := findMonitorsByType(getMonitors(), "dynamic2")
		Expect(len(mons)).To(Equal(0))

		manager.Configure([]config.MonitorConfig{
			{
				Type: "static1",
			},
			{
				Type: "dynamic2",
				OtherConfig: map[string]interface{}{
					"host": "example.com",
					"port": 80,
				},
			},
		}, &config.CollectdConfig{}, 10)

		mons = findMonitorsByType(getMonitors(), "dynamic2")
		Expect(len(mons)).To(Equal(1))
	})

	It("Monitors self-configured endpoints", func() {
		manager.Configure([]config.MonitorConfig{
			{
				Type: "static1",
			},
		}, &config.CollectdConfig{}, 10)

		endpoint := newService("my-service", 5000)
		endpoint.Core().MonitorType = "dynamic1"
		endpoint.Core().Configuration = map[string]interface{}{
			"myVar": "testing",
		}

		manager.EndpointAdded(endpoint)

		mons := findMonitorsByType(getMonitors(), "dynamic1")
		Expect(len(mons)).To(Equal(1))
		Expect(mons[0].MyVar()).To(Equal("testing"))
	})

	It("Merges self-configured endpoint config with agent config", func() {
		manager.Configure([]config.MonitorConfig{
			{
				Type:          "dynamic1",
				DiscoveryRule: `port == 5000 && container_image =~ "my-service"`,
				OtherConfig: map[string]interface{}{
					"password": "s3cr3t",
				},
			},
		}, &config.CollectdConfig{}, 10)

		endpoint := newService("my-service", 5000)
		endpoint.Core().Configuration = map[string]interface{}{
			"myVar": "testing",
		}

		manager.EndpointAdded(endpoint)

		mons := findMonitorsByType(getMonitors(), "dynamic1")
		Expect(len(mons)).To(Equal(1))
		Expect(mons[0].MyVar()).To(Equal("testing"))
		Expect(mons[0].Password()).To(Equal("s3cr3t"))
	})

	It("Does not double monitor self-configured endpoint", func() {
		manager.Configure([]config.MonitorConfig{
			{
				Type: "static1",
			},
		}, &config.CollectdConfig{}, 10)

		endpoint := newService("my-service", 5000)
		endpoint.Core().MonitorType = "dynamic1"
		endpoint.Core().Configuration = map[string]interface{}{
			"myVar": "testing",
		}

		manager.EndpointAdded(endpoint)

		manager.Configure([]config.MonitorConfig{
			{
				Type: "static1",
			},
			{
				Type:          "dynamic1",
				DiscoveryRule: `container_image =~ "my-service"`,
			},
		}, &config.CollectdConfig{}, 10)

		mons := findMonitorsByType(getMonitors(), "dynamic1")
		Expect(len(mons)).To(Equal(1))
		Expect(mons[0].MyVar()).To(Equal("testing"))

		manager.EndpointRemoved(endpoint)

		endpoint.Core().MonitorType = ""
		manager.EndpointAdded(endpoint)

		mons = findMonitorsByType(getMonitors(), "dynamic1")
		Expect(len(mons)).To(Equal(1))
	})

	It("Does not stop monitoring self-configured endpoint on reconfig", func() {
		manager.Configure([]config.MonitorConfig{
			{
				Type: "static1",
			},
		}, &config.CollectdConfig{}, 10)

		endpoint := newService("my-service", 5000)
		endpoint.Core().MonitorType = "dynamic1"
		endpoint.Core().Configuration = map[string]interface{}{
			"myVar": "testing",
		}

		manager.EndpointAdded(endpoint)

		manager.Configure([]config.MonitorConfig{
			{
				Type: "static1",
			},
			{
				Type:          "dynamic2",
				DiscoveryRule: `container_image =~ "not-my-service"`,
			},
		}, &config.CollectdConfig{}, 10)

		mons := findMonitorsByType(getMonitors(), "dynamic1")
		Expect(len(mons)).To(Equal(1))
		Expect(mons[0].MyVar()).To(Equal("testing"))
	})

	It("Stops monitoring self-configured endpoint when removed", func() {
		manager.Configure([]config.MonitorConfig{
			{
				Type: "static1",
			},
		}, &config.CollectdConfig{}, 10)

		endpoint := newService("my-service", 5000)
		endpoint.Core().MonitorType = "dynamic1"
		endpoint.Core().Configuration = map[string]interface{}{
			"myVar": "testing",
		}

		manager.EndpointAdded(endpoint)

		staticMons := findMonitorsByType(getMonitors(), "static1")
		Expect(len(staticMons)).To(Equal(1))
		mons := findMonitorsByType(getMonitors(), "dynamic1")
		Expect(len(mons)).To(Equal(1))

		manager.EndpointRemoved(endpoint)

		staticMons = findMonitorsByType(getMonitors(), "static1")
		Expect(len(staticMons)).To(Equal(1))

		mons = findMonitorsByType(getMonitors(), "dynamic1")
		Expect(len(mons)).To(Equal(0))
	})
})

func TestMonitors(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Monitor Suite")
}
