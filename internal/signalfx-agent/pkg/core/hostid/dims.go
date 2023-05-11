package hostid

import (
	"os"
	"sync"

	"github.com/signalfx/signalfx-agent/pkg/core/config"
	log "github.com/sirupsen/logrus"
)

// Dimensions returns a map of host-specific dimensions that are derived from
// the environment.
func Dimensions(conf *config.Config) map[string]string {
	log.Info("Fetching host id dimensions")

	var g dimGatherer

	g.GatherDim("host", func() string {
		if conf.Hostname != "" {
			return conf.Hostname
		}
		// Using the FQDN needs to default to true but the defaults lib that we
		// use can't distinguish between false and unspecified, so figure out
		// if the user specified it explicitly as false with this logic.
		return getHostname(conf.UseFullyQualifiedHost == nil || *conf.UseFullyQualifiedHost)
	})

	// The envvar exists primarily for testing but could be useful otherwise.
	// It remains undocumented for the time being though.
	if os.Getenv("SKIP_PLATFORM_HOST_DIMS") != "yes" {
		g.GatherDim("AWSUniqueId", func() string {
			return AWSUniqueID(conf.CloudMetadataTimeout)
		})
		g.GatherDim("gcp_id", func() string {
			return GoogleComputeID(conf.CloudMetadataTimeout)
		})
		g.GatherDim("kubernetes_node_uid", func() string {
			return KubernetesNodeUID(conf.Monitors)
		})
		g.GatherDim("azure_resource_id", func() string {
			return AzureUniqueID(conf.CloudMetadataTimeout)
		})
	}

	dims := g.WaitForDimensions()

	return dims
}

// Helper to fire off the dim lookups in parallel to minimize delay to agent
// start up.
type dimGatherer struct {
	lock sync.Mutex
	dims map[string]string
	wg   sync.WaitGroup
}

// GatherDim inserts the given dim key based on the output of the provider
// func.  If the output is blank, the dimension will not be inserted.
func (dg *dimGatherer) GatherDim(key string, provider func() string) {
	dg.wg.Add(1)
	go func() {
		res := provider()
		if res != "" {
			dg.lock.Lock()
			if dg.dims == nil {
				dg.dims = make(map[string]string)
			}

			dg.dims[key] = res
			dg.lock.Unlock()
		}
		dg.wg.Done()
	}()
}

func (dg *dimGatherer) WaitForDimensions() map[string]string {
	dg.wg.Wait()
	return dg.dims
}
