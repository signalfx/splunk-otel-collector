package monitors

import (
	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/monitors/types"
)

// This code is somewhat convoluted, but basically it creates two types of mock
// monitors, static and dynamic.  It handles doing basic tracking of whether
// the instances have been configured and how, so that we don't have to pry
// into the internals of the manager.

type Config struct {
	config.MonitorConfig
	MyVar   string `yaml:"myVar"`
	MySlice []string
}

type DynamicConfig struct {
	config.MonitorConfig `acceptsEndpoints:"true"`

	Host string `yaml:"host" validate:"required"`
	Port uint16 `yaml:"port" validate:"required"`
	Name string `yaml:"name"`

	MyVar    string `yaml:"myVar"`
	Password string `yaml:"password"`
}

type MockMonitor interface {
	SetConfigHook(func(types.MonitorID, MockMonitor))
	AddShutdownHook(fn func())
	Type() string
	MyVar() string
	Password() string
}

type _MockMonitor struct {
	MType         string
	MMyVar        string
	MPassword     string
	shutdownHooks []func()
	configHook    func(types.MonitorID, MockMonitor)
}

func (m *_MockMonitor) Configure(conf *Config) error {
	m.MType = conf.Type
	m.MMyVar = conf.MyVar
	m.configHook(conf.MonitorID, m)
	return nil
}

func (m *_MockMonitor) Type() string {
	return m.MType
}

func (m *_MockMonitor) MyVar() string {
	return m.MMyVar
}

func (m *_MockMonitor) Password() string {
	return m.MPassword
}

func (m *_MockMonitor) SetConfigHook(fn func(types.MonitorID, MockMonitor)) {
	m.configHook = fn
}

func (m *_MockMonitor) AddShutdownHook(fn func()) {
	m.shutdownHooks = append(m.shutdownHooks, fn)
}

func (m *_MockMonitor) Shutdown() {
	for _, hook := range m.shutdownHooks {
		hook()
	}
}

type _MockServiceMonitor struct {
	_MockMonitor
}

func (m *_MockServiceMonitor) Configure(conf *DynamicConfig) error {
	m.MType = conf.Type
	m.MMyVar = conf.MyVar
	m.MPassword = conf.Password
	m.configHook(conf.MonitorID, m)
	return nil
}

type Static1 struct{ _MockMonitor }
type Static2 struct{ _MockMonitor }
type Dynamic1 struct{ _MockServiceMonitor }
type Dynamic2 struct{ _MockServiceMonitor }

func RegisterFakeMonitors() func() map[types.MonitorID]MockMonitor {
	instances := map[types.MonitorID]MockMonitor{}

	track := func(factory func() interface{}) func() interface{} {
		return func() interface{} {
			mon := factory().(MockMonitor)
			mon.SetConfigHook(func(id types.MonitorID, mon MockMonitor) {
				instances[id] = mon

				mon.AddShutdownHook(func() {
					delete(instances, id)
				})
			})

			return mon
		}
	}

	Register(&Metadata{MonitorType: "static1"}, track(func() interface{} { return &Static1{} }), &Config{})
	Register(&Metadata{MonitorType: "static2"}, track(func() interface{} { return &Static2{} }), &Config{})
	Register(&Metadata{MonitorType: "dynamic1"}, track(func() interface{} { return &Dynamic1{} }), &DynamicConfig{})
	Register(&Metadata{MonitorType: "dynamic2"}, track(func() interface{} { return &Dynamic2{} }), &DynamicConfig{})

	return func() map[types.MonitorID]MockMonitor {
		return instances
	}
}

func findMonitorsByType(monitors map[types.MonitorID]MockMonitor, _type string) []MockMonitor {
	mons := []MockMonitor{}
	for _, m := range monitors {
		if m.Type() == _type {
			mons = append(mons, m)
		}
	}
	return mons
}
