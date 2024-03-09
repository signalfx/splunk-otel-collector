package measurements

import (
	"fmt"

	"github.com/mongodb/go-client-mongodb-atlas/mongodbatlas"
	log "github.com/sirupsen/logrus"
)

// Process is the MongoDB Process identified by the host and port on which the Process is running.
type Process struct {
	ID             string
	ProjectID      string
	Host           string
	ShardName      string
	ReplicaSetName string
	TypeName       string
	Port           int
}

// nextPage gets the next page for pagination request.
func nextPage(resp *mongodbatlas.Response, logger log.FieldLogger) (bool, int) {
	if resp == nil || len(resp.Links) == 0 || resp.IsLastPage() {
		return false, -1
	}

	currentPage, err := resp.CurrentPage()

	if err != nil {
		logger.WithError(err).Error("failed to get the next page")
		return false, -1
	}

	return true, currentPage + 1
}

func errorMsg(err error, resp *mongodbatlas.Response) (string, error) {
	if err != nil {
		return "request for getting %s failed (Atlas project: %s, host: %s, port: %d)", err
	}

	if resp == nil {
		return "response for getting %s returned empty (Atlas project: %s, host: %s, port: %d)", fmt.Errorf("empty response")
	}

	if err := mongodbatlas.CheckResponse(resp.Response); err != nil {
		return "response for getting %s returned error (Atlas project: %s, host: %s, port: %d)", err
	}

	return "", nil
}
