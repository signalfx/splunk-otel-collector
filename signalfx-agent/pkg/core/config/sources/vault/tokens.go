// Copyright  Splunk, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package vault

import (
	"encoding/json"

	"github.com/hashicorp/vault/api"
	log "github.com/sirupsen/logrus"
)

func (v *vaultConfigSource) initTokenRenewalIfNeeded() error {
	if v.client.Token() == "" {
		// Blank tokens should never be allowed
		panic("Vault token must be set")
	}

	tokenAuth := v.client.Auth().Token()
	authSec, err := tokenAuth.LookupSelf()
	if err != nil {
		return err
	}

	if authSec.Auth == nil {
		authSec.Auth = &api.SecretAuth{}
	}

	// For some reason, the token lookup returns token data in the `Data` field
	// but the renewals of tokens return it in the proper `Auth` field.  This
	// seems like a Vault bug or at least a major inconsistency.
	authSec.Auth.Renewable, _ = authSec.Data["renewable"].(bool)

	if ttl, ok := authSec.Data["ttl"].(json.Number); ok {
		if ttlInt, err := ttl.Int64(); err != nil {
			authSec.Auth.LeaseDuration = int(ttlInt)
		}
	}

	authSec.Auth.ClientToken = v.client.Token()

	renewer, err := v.client.NewLifetimeWatcher(&api.LifetimeWatcherInput{
		Secret:        authSec,
		RenewBehavior: api.RenewBehaviorErrorOnErrors,
	})
	if err != nil {
		return err
	}

	v.tokenRenewer = renewer
	go renewer.Renew()

	go func() {
		for {
			select {
			case err := <-renewer.DoneCh():
				if err == api.ErrRenewerNotRenewable {
					log.Info("Vault token is not renewable, assuming valid indefinitely")
				} else {
					log.WithError(err).Error("Could not renew Vault token")
				}
				return
			case <-renewer.RenewCh():
				log.Info("Vault token renewed")
			}
		}
	}()
	return nil
}
