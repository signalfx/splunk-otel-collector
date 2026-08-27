package promqlreceiver

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/prometheus/prometheus/config"
)

func TestQueryPrometheusAPI(t *testing.T) {
	cfg := &config.Config{
		GlobalConfig: config.GlobalConfig{},
		ScrapeConfigs: []*config.ScrapeConfig{
			{
				JobName: "my-dynamic-job",
			},
		},
	}

	srv := server.
		fmt.Println("Starting Prometheus programmatically...")
	if err := srv.Run(context.Background()); err != nil {
		fmt.Printf("Error running server: %v\n", err)
		os.Exit(1)
	}
}
