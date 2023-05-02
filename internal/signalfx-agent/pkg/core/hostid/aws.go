package hostid

import (
	"fmt"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	log "github.com/sirupsen/logrus"

	"github.com/signalfx/signalfx-agent/pkg/utils/timeutil"
)

// AWSUniqueID constructs the unique EC2 instance of the underlying host.  If
// not running on EC2, returns the empty string.
func AWSUniqueID(cloudMetadataTimeout timeutil.Duration) string {
	// Pass in the HTTP client so cloudMetadataTimeout from the config
	// is respected.
	c := &http.Client{
		Timeout: cloudMetadataTimeout.AsDuration(),
	}

	sess, err := session.NewSession(aws.NewConfig().WithHTTPClient(c))
	if err != nil {
		log.WithFields(log.Fields{
			"detail": err,
		}).Info("Failed to create new session for AWS metadata collection")
		return ""
	}

	client := ec2metadata.New(sess)

	doc, err := client.GetInstanceIdentityDocument()
	if err != nil {
		log.WithFields(log.Fields{
			"detail": err,
		}).Info("No AWS metadata server detected, assuming not on EC2")
		return ""
	}

	if doc.AccountID == "" || doc.InstanceID == "" || doc.Region == "" {
		log.Errorf("One (or more) required field is empty. AccountID: %s ; InstanceID: %s ; Region: %s", doc.AccountID, doc.InstanceID, doc.Region)
		return ""
	}

	return fmt.Sprintf("%s_%s_%s", doc.InstanceID, doc.Region, doc.AccountID)
}
