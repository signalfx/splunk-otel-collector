package cloudfoundry

import (
	"github.com/cloudfoundry-incubator/uaago"
)

func getUAAToken(uaaURL, username, password string, skipVerification bool) (string, error) {
	uaaClient, err := uaago.NewClient(uaaURL)
	if err != nil {
		return "", err
	}

	token, _, err := uaaClient.GetAuthTokenWithExpiresIn(username, password, skipVerification)
	return token, err
}
