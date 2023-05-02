package config

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/signalfx/signalfx-agent/pkg/utils"
)

var _ = Describe("Config Loader", func() {
	var dir string
	var ctx context.Context
	var cancel func()

	mkFile := func(path string, content string) string {
		fullPath := filepath.Join(dir, path)

		err := os.MkdirAll(filepath.Dir(fullPath), 0755)
		Expect(err).ShouldNot(HaveOccurred())

		err = ioutil.WriteFile(fullPath, []byte(content), 0644)
		Expect(err).ShouldNot(HaveOccurred())

		return fullPath
	}

	outdent := utils.StripIndent

	BeforeEach(func() {
		var err error
		dir, err = ioutil.TempDir("", "loader-test")
		Expect(err).ShouldNot(HaveOccurred())

		ctx, cancel = context.WithCancel(context.Background())
	})

	AfterEach(func() {
		if dir != "" {
			os.RemoveAll(dir)
		}
		cancel()
	})

	It("Loads a basic config file", func() {
		path := mkFile("agent/agent.yaml", `signalFxAccessToken: abcd`)
		loads, err := LoadConfig(ctx, path)
		Expect(err).ShouldNot(HaveOccurred())

		var config *Config
		Eventually(loads).Should(Receive(&config))

		Expect(config.SignalFxAccessToken).To(Equal("abcd"))
	})

	It("Does basic validation checks on a config file", func() {
		path := mkFile("agent/agent.yaml", outdent(`
			signalFxAccessToken: abcd
			monitors: {}
		`))
		_, err := LoadConfig(ctx, path)
		Expect(err).To(HaveOccurred())
	})

	It("Errors on missing source path", func() {
		path := mkFile("agent/agent.yaml", outdent(fmt.Sprintf(`
			signalFxAccessToken: {"#from": '%s/does-not-exist'}
		`, dir)))

		_, err := LoadConfig(ctx, path)
		Expect(err).To(HaveOccurred())
	})

	It("Fills in dynamic values", func() {
		tokenPath := mkFile("agent/token", "abcd")
		path := mkFile("agent/agent.yaml", outdent(fmt.Sprintf(`
			signalFxAccessToken: {"#from": '%s'}
		`, tokenPath)))

		loads, err := LoadConfig(ctx, path)
		Expect(err).ShouldNot(HaveOccurred())

		var config *Config
		Eventually(loads).Should(Receive(&config))

		Expect(config.SignalFxAccessToken).To(Equal("abcd"))
	})

	It("Will merge seq into single seq", func() {
		mkFile("agent/conf/mon1.yaml", outdent(`
			- a
			- b
		`))

		mkFile("agent/conf/mon2.yaml", outdent(`
			- c
		`))

		path := mkFile("agent/agent.yaml", outdent(fmt.Sprintf(`
			signalFxAccessToken: abcd
			monitors:
			- type: collectd/mysql
			  databases: {"#from": '%s/agent/conf/*.yaml'}
		`, dir)))

		loads, err := LoadConfig(ctx, path)
		Expect(err).ShouldNot(HaveOccurred())

		var config *Config
		Eventually(loads).Should(Receive(&config))

		Expect(config.Monitors[0].OtherConfig["databases"]).Should(ConsistOf("a", "b", "c"))
	})

	It("Will error if merging maps and seqs", func() {
		mkFile("agent/conf/mon1.yaml", outdent(`
			- a
			- b
		`))

		mkFile("agent/conf/mon2.yaml", outdent(`
			env: dev
		`))

		path := mkFile("agent/agent.yaml", outdent(fmt.Sprintf(`
			signalFxAccessToken: abcd
			monitors:
			- type: collectd/mysql
			  databases: {"#from": '%s/agent/conf/*.yaml'}
		`, dir)))

		_, err := LoadConfig(ctx, path)
		Expect(err).To(HaveOccurred())
	})

	It("Will render seq into seq", func() {
		mkFile("agent/conf/databases.yaml", outdent(`
			[a, b, c]
		`))

		path := mkFile("agent/agent.yaml", outdent(fmt.Sprintf(`
			signalFxAccessToken: abcd
			monitors:
			- type: collectd/mysql
			  databases: {"#from": '%s/agent/conf/databases.yaml'}
		`, dir)))

		loads, err := LoadConfig(ctx, path)
		Expect(err).ShouldNot(HaveOccurred())

		var config *Config
		Eventually(loads).Should(Receive(&config))

		Expect(config.Monitors[0].OtherConfig["databases"]).Should(ConsistOf("a", "b", "c"))
	})

	It("Flattens seqs into seq", func() {
		mkFile("agent/conf/mon1.yaml", outdent(`
			- type: collectd/cpu
			  intervalSeconds: 5
			- type: collectd/vmem
		`))

		mkFile("agent/conf/mon2.yaml", outdent(`
			- type: collectd/mysql
			  host: 127.0.0.1
		`))

		path := mkFile("agent/agent.yaml", outdent(fmt.Sprintf(`
			signalFxAccessToken: abcd
			monitors:
			- type: fake
			- {"#from": '%s/agent/conf/*.yaml', flatten: true}
		`, dir)))

		loads, err := LoadConfig(ctx, path)
		Expect(err).ShouldNot(HaveOccurred())

		var config *Config
		Eventually(loads).Should(Receive(&config))

		Expect(len(config.Monitors)).To(Equal(4))
	})

	It("Will not flatten seqs into map", func() {
		mkFile("agent/conf/mon1.yaml", outdent(`
			- type: collectd/cpu
			  intervalSeconds: 5
			- type: collectd/vmem
		`))

		mkFile("agent/conf/mon2.yaml", outdent(`
			- type: collectd/mysql
			  host: 127.0.0.1
		`))

		path := mkFile("agent/agent.yaml", outdent(fmt.Sprintf(`
			signalFxAccessToken: abcd
			monitors:
			- type: collectd/mysql
			  _: {"#from": '%s/agent/conf/*.yaml', flatten: true}
		`, dir)))

		_, err := LoadConfig(ctx, path)
		Expect(err).To(HaveOccurred())
	})

	It("Flattens dynamic map into map", func() {
		mkFile("agent/conf/mon1.yaml", outdent(`
			port: 80
		`))

		mkFile("agent/conf/mon2.yaml", outdent(`
		  	host: 127.0.0.1
		`))

		path := mkFile("agent/agent.yaml", outdent(fmt.Sprintf(`
			signalFxAccessToken: abcd
			monitors:
			- type: collectd/mysql
			  _: {"#from": '%s/agent/conf/*.yaml', flatten: true}
		`, dir)))

		loads, err := LoadConfig(ctx, path)
		Expect(err).ShouldNot(HaveOccurred())

		var config *Config
		Eventually(loads).Should(Receive(&config))

		Expect(len(config.Monitors)).To(Equal(1))
		Expect(config.Monitors[0].OtherConfig["port"]).To(Equal(80))
	})

	It("Will flatten dynamic map from key that start's with _", func() {
		mkFile("agent/conf/mon1.yaml", outdent(`
			port: 80
			name: test
		`))

		mkFile("agent/conf/mon2.yaml", outdent(`
		  	host: 127.0.0.1
		`))

		path := mkFile("agent/agent.yaml", outdent(fmt.Sprintf(`
			signalFxAccessToken: abcd
			monitors:
			- type: collectd/mysql
			  _: {"#from": '%s/agent/conf/*.yaml', flatten: true}
		`, dir)))

		loads, err := LoadConfig(ctx, path)
		Expect(err).ShouldNot(HaveOccurred())

		var config *Config
		Eventually(loads).Should(Receive(&config))

		Expect(config.Monitors[0].OtherConfig["name"]).To(Equal("test"))
		Expect(config.Monitors[0].OtherConfig["port"]).To(Equal(80))
		Expect(config.Monitors[0].OtherConfig["host"]).To(Equal("127.0.0.1"))
	})

	It("Will not flatten dynamic map from key that doesn't start with _", func() {
		mkFile("agent/conf/mon1.yaml", outdent(`
			port: 80
			name: test
		`))

		mkFile("agent/conf/mon2.yaml", outdent(`
		  	host: 127.0.0.1
		`))

		path := mkFile("agent/agent.yaml", outdent(fmt.Sprintf(`
			signalFxAccessToken: abcd
			monitors:
			- type: collectd/mysql
			  password: {"#from": '%s/agent/conf/*.yaml', flatten: true}
		`, dir)))

		_, err := LoadConfig(ctx, path)
		Expect(err).To(HaveOccurred())
	})

	It("Watches dynamic value source for changes", func() {
		tokenPath := mkFile("agent/token", "abcd")
		path := mkFile("agent/agent.yaml", outdent(fmt.Sprintf(`
			signalFxAccessToken: {"#from": '%s'}
			configSources:
			  file:
			    pollRateSeconds: 1
		`, tokenPath)))

		loads, err := LoadConfig(ctx, path)
		Expect(err).ShouldNot(HaveOccurred())

		var config *Config
		Eventually(loads).Should(Receive(&config))

		mkFile("agent/token", "1234")
		Eventually(loads, 3).Should(Receive(&config))

		Expect(config.SignalFxAccessToken).To(Equal("1234"))
	})

	It("Recursively watches dynamic value source for changes", func() {
		passwordPath := mkFile("agent/password", "s3cr3t")

		monitorPath := mkFile("agent/token", outdent(fmt.Sprintf(`
			type: my-monitor
			password: {"#from": '%s'}
		`, passwordPath)))

		path := mkFile("agent/agent.yaml", outdent(fmt.Sprintf(`
			signalFxAccessToken: "abcd"
			configSources:
			  file:
			    pollRateSeconds: 1
			monitors:
			- {"#from": '%s'}
		`, monitorPath)))

		loads, err := LoadConfig(ctx, path)
		Expect(err).ShouldNot(HaveOccurred())

		var config *Config
		Eventually(loads).Should(Receive(&config))

		Expect(config.Monitors[0].OtherConfig["password"]).To(Equal("s3cr3t"))

		mkFile("agent/password", "sup3rs3cr3t")
		Eventually(loads, 2).Should(Receive(&config))

		Expect(config.Monitors[0].OtherConfig["password"]).To(Equal("sup3rs3cr3t"))
	})

	It("Recursively watches three levels deep for changes", func() {
		envPath := mkFile("agent/env", "dev")

		dimPath := mkFile("agent/dim", outdent(fmt.Sprintf(`
			author: bob
			env: {"#from": '%s'}
		`, envPath)))

		monitorPath := mkFile("agent/monitors", outdent(fmt.Sprintf(`
			type: my-monitor
			extraDimensions: {"#from": '%s'}
		`, dimPath)))

		path := mkFile("agent/agent.yaml", outdent(fmt.Sprintf(`
			signalFxAccessToken: "abcd"
			configSources:
			  file:
			    pollRateSeconds: 1
			monitors:
			- {"#from": '%s'}
		`, monitorPath)))

		loads, err := LoadConfig(ctx, path)
		Expect(err).ShouldNot(HaveOccurred())

		var config *Config
		Eventually(loads).Should(Receive(&config))

		Expect(config.Monitors[0].ExtraDimensions["env"]).To(Equal("dev"))

		mkFile("agent/env", "prod")
		Eventually(loads, 2).Should(Receive(&config))

		Expect(config.Monitors[0].ExtraDimensions["env"]).To(Equal("prod"))
	})

	It("Handles dynamic value reference loops without blowing out stack", func() {
		configPath := mkFile("agent/agent.yaml", "")

		dimPath := mkFile("agent/dims", outdent(fmt.Sprintf(`
			env: {"#from": "%s"}
		`, configPath)))

		monitorPath := mkFile("agent/monitors", outdent(fmt.Sprintf(`
			type: my-monitor
			extraDimensions: {"#from": "%s"}
		`, dimPath)))

		mkFile("agent/agent.yaml", outdent(fmt.Sprintf(`
			signalFxAccessToken: "abcd"
			configSources:
			  file:
			    pollRateSeconds: 1
			monitors:
			- {"#from": "%s", flatten: true}
		`, monitorPath)))

		_, err := LoadConfig(ctx, configPath)
		Expect(err).To(HaveOccurred())
	})

	It("Allows multiple references to the same source path", func() {
		tokenPath := mkFile("agent/token", "abcd")
		path := mkFile("agent/agent.yaml", outdent(fmt.Sprintf(`
			signalFxAccessToken: {"#from": '%s'}
			configSources:
			  file:
			    pollRateSeconds: 1
			monitors:
			- type: my-monitor
			  password: {"#from": '%s'}
			  other: {"#from": '%s'}
		`, tokenPath, tokenPath, tokenPath)))

		loads, err := LoadConfig(ctx, path)
		Expect(err).ShouldNot(HaveOccurred())

		var config *Config
		Eventually(loads).Should(Receive(&config))

		mkFile("agent/token", "1234")
		Eventually(loads, 2).Should(Receive(&config))

		Expect(config.SignalFxAccessToken).To(Equal("1234"))
		Expect(config.Monitors[0].OtherConfig["password"]).To(Equal(1234))
	})

	It("Allows optional values", func() {
		path := mkFile("agent/agent.yaml", outdent(fmt.Sprintf(`
			signalFxAccessToken: abcd
			monitors:
			- type: collectd/mysql
			  password: {"#from": '%s/agent/mysql-password.yaml', optional: true}
		`, dir)))

		loads, err := LoadConfig(ctx, path)
		Expect(err).ShouldNot(HaveOccurred())

		var config *Config
		Eventually(loads).Should(Receive(&config))

		Expect(config.Monitors[0].OtherConfig["password"]).To(BeNil())
	})

	It("Allows default values", func() {
		path := mkFile("agent/agent.yaml", outdent(fmt.Sprintf(`
			signalFxAccessToken: abcd
			monitors:
			- type: collectd/mysql
			  password: {"#from": '%s/agent/mysql-password.yaml', default: "s3cr3t"}
		`, dir)))

		loads, err := LoadConfig(ctx, path)
		Expect(err).ShouldNot(HaveOccurred())

		var config *Config
		Eventually(loads).Should(Receive(&config))

		Expect(config.Monitors[0].OtherConfig["password"]).To(Equal("s3cr3t"))
	})

	It("Will render raw seq into seq position", func() {
		mkFile("agent/conf/config.yaml", outdent(`
		    LoadPlugin "cpufreq"
		`))

		path := mkFile("agent/agent.yaml", outdent(fmt.Sprintf(`
			signalFxAccessToken: abcd
			monitors:
			- type: collectd/my-monitor
			  templates:
			  - {"#from": '%s/agent/conf/*.yaml', flatten: true, raw: true}
		`, dir)))

		loads, err := LoadConfig(ctx, path)
		Expect(err).ShouldNot(HaveOccurred())

		var config *Config
		Eventually(loads).Should(Receive(&config))

		Expect(config.Monitors[0].OtherConfig["templates"]).Should(ConsistOf(`LoadPlugin "cpufreq"`))
	})

	It("Allows JSONPath processing of source values", func() {
		path := mkFile("agent/agent.yaml", outdent(`
			signalFxAccessToken: {"#from": "env:ASDF_INFO", jsonPath: "$.token"}
		`))

		os.Setenv("ASDF_INFO", `{"token": "s3cr3t", "app_name": "my_app"}`)
		loads, err := LoadConfig(ctx, path)
		os.Unsetenv("ASDF_INFO")

		Expect(err).ShouldNot(HaveOccurred())

		var config *Config
		Eventually(loads).Should(Receive(&config))

		Expect(config.SignalFxAccessToken).Should(Equal(`s3cr3t`))
	})

})

func TestLoader(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Config Loader")
}
