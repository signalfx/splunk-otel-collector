// This file contains all of the logic for setting up/initializing the scraper which will pull
// from the REST API. The scraper object will wrap the client and handle the request dispatching,
// converting the json response to a metric, and emitting this metric.

package githubmetricsreceiver

type githubMetricsScraper struct {
    client githubmetricsClinet
}
