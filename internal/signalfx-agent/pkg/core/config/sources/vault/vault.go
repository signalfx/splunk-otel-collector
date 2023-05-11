// Package vault contains the logic for using Vault as a remote config source
//
// How to use auth methods with Vault Go client: https://groups.google.com/forum/#!msg/vault-tool/cS7J2KbAwZg/7pu6PYSRAAAJ
package vault

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/vault/api"
	log "github.com/sirupsen/logrus"

	"github.com/signalfx/signalfx-agent/pkg/core/config/sources/vault/auth"
	"github.com/signalfx/signalfx-agent/pkg/core/config/types"
)

type vaultConfigSource struct {
	sync.Mutex
	// The Vault client
	client *api.Client
	// Secrets that have been read from Vault
	secretsByVaultPath                map[string]*api.Secret
	renewersByVaultPath               map[string]*api.LifetimeWatcher
	customWatchersByVaultPath         map[string]customWatcher
	nonRenewableVaultPathRefetchTimes map[string]time.Time
	// Used for unit testing
	nowProvider  func() time.Time
	conf         *Config
	tokenRenewer *api.LifetimeWatcher
	logger       log.FieldLogger
}

var _ types.Stoppable = &vaultConfigSource{}

// Config for the vault remote config
type Config struct {
	// The Vault Address.  Can also be provided by the standard Vault envvar
	// `VAULT_ADDR`.  This option takes priority over the envvar if provided.
	VaultAddr string `yaml:"vaultAddr"`
	// The Vault token, can also be provided by it the standard Vault envvar
	// `VAULT_TOKEN`.  This option takes priority over the envvar if provided.
	VaultToken string `yaml:"vaultToken" neverLog:"true"`
	// The polling interval for checking KV V2 secrets for a new version.  This
	// can be any string value that can be parsed by
	// https://golang.org/pkg/time/#ParseDuration.
	KVV2PollInterval time.Duration `yaml:"kvV2PollInterval" default:"60s"`
	// The authetication method to use, if any, to obtain the Vault token.  If
	// `vaultToken` is specified above, this option will have no effect.
	// Currently supported values are: `iam`.
	AuthMethod string `yaml:"authMethod"`
	// Further config options for the `iam` auth method.  These options are
	// identical to the [CLI helper tool options](https://github.com/hashicorp/vault/blob/v1.1.0/builtin/credential/aws/cli.go#L148)
	IAM auth.IAMConfig `yaml:"iam" default:"{}"`
	// Further config options for the `gcp` auth method. These options are
	// identical to the [CLI helper tool
	// options](https://github.com/hashicorp/vault-plugin-auth-gcp/blob/e1f6784b379d277038ca0661606aa8d23791e392/plugin/cli.go#L120).
	// You must provide a valid GCP IAM credential JSON either explicitly via
	// the `credentials` option (not recommended), or through any GCP
	// Application Default Credentials.
	GCP auth.GCPConfig `yaml:"gcp" default:"{}"`
}

// Validate the config
func (c *Config) Validate() error {
	// Defaults don't work for time.Duration fields
	if c.KVV2PollInterval == time.Duration(0) {
		c.KVV2PollInterval = 60 * time.Second
	}
	if c.VaultToken == "" && c.AuthMethod == "" {
		if os.Getenv("VAULT_TOKEN") == "" {
			return errors.New("vault token is required, either in the agent config, the envvar VAULT_TOKEN, or via an authMethod")
		}

		c.VaultToken = os.Getenv("VAULT_TOKEN")
	}

	if c.AuthMethod != "" {
		matched := false
		for _, method := range c.supportedAuthMethods() {
			if method.Name() == c.AuthMethod {
				matched = true
			}
		}
		if !matched {
			return errors.New("unsupported auth method: " + c.AuthMethod)
		}
	}

	return nil
}

func (c *Config) supportedAuthMethods() []auth.Method {
	return []auth.Method{
		&c.IAM,
		&c.GCP,
	}
}

var _ types.ConfigSourceConfig = &Config{}

// New creates a new Vault remote config source from the target config
func (c *Config) New() (types.ConfigSource, error) {
	return New(c)
}

// New creates a new vault ConfigSource
func New(conf *Config) (types.ConfigSource, error) {
	logger := log.WithFields(log.Fields{"remoteConfigSource": "vault"})

	logger.Info("Initializing new Vault remote config instance")

	c, err := api.NewClient(&api.Config{
		Address: conf.VaultAddr,
	})
	if err != nil {
		return nil, err
	}

	if conf.AuthMethod == "" {
		c.SetToken(conf.VaultToken)
	} else {
		for _, am := range conf.supportedAuthMethods() {
			if conf.AuthMethod == am.Name() {
				tokenSecret, err := am.GetToken(c)
				if err != nil {
					return nil, fmt.Errorf("could not get token from auth method %s: %v", conf.AuthMethod, err)
				}
				c.SetToken(tokenSecret.Auth.ClientToken)
			}
		}
	}

	vcs := &vaultConfigSource{
		client:                            c,
		secretsByVaultPath:                make(map[string]*api.Secret),
		renewersByVaultPath:               make(map[string]*api.LifetimeWatcher),
		customWatchersByVaultPath:         make(map[string]customWatcher),
		nonRenewableVaultPathRefetchTimes: make(map[string]time.Time),
		nowProvider:                       time.Now,
		conf:                              conf,
		logger:                            logger,
	}

	err = vcs.initTokenRenewalIfNeeded()
	if err != nil {
		return nil, err
	}

	return vcs, nil
}

func (v *vaultConfigSource) Name() string {
	return "vault"
}

func (v *vaultConfigSource) Get(path string) (map[string][]byte, uint64, error) {
	v.Lock()
	defer v.Unlock()

	vaultPath, key, err := splitConfigPath(path)
	if err != nil {
		return nil, 0, err
	}

	secret, ok := v.secretsByVaultPath[vaultPath]
	if !ok {
		v.logger.Debugf("Reading Vault secret at path: %s", vaultPath)

		secret, err = v.client.Logical().Read(vaultPath)
		if err != nil {
			return nil, 0, err
		}

		if secret == nil {
			return nil, 0, fmt.Errorf("no secret found at path %s", vaultPath)
		}

		switch {
		case secret.Renewable:
			renewer, err := v.client.NewLifetimeWatcher(&api.LifetimeWatcherInput{
				Secret:        secret,
				RenewBehavior: api.RenewBehaviorErrorOnErrors,
			})
			if err == nil {
				v.logger.Debugf("Setting up Vault renewer for secret at path %s", vaultPath)
				v.renewersByVaultPath[vaultPath] = renewer
				go renewer.Renew()
			} else {
				v.logger.Errorf("Could not set up renewal on Vault secret at path %s: %v", vaultPath, err)
			}
		case secret.LeaseDuration > 0:
			// We have a secret that isn't renewable but still expires.  We
			// need to just refetch it before it expires.  Set the refetch time
			// to half the lease duration.
			v.logger.Debugf("Secret at path %s cannot be renewed, refetching within %d seconds", vaultPath, secret.LeaseDuration)
			v.nonRenewableVaultPathRefetchTimes[vaultPath] = time.Now().Add(time.Duration(secret.LeaseDuration/2) * time.Second)
		default:
			// This secret is not renewable or on a lease.  If it has a
			// "metadata" field and has "/data/" in the vault path, then it is
			// probably a KV v2 secret.  In that case, we do a poll on the
			// secret's metadata to refresh it in the agent if a new version is
			// added to the secret.
			if md := secret.Data["metadata"]; md != nil && strings.Contains(vaultPath, "/data/") {
				watcher, err := newPollingKVV2Watcher(vaultPath, secret, v.client, v.conf.KVV2PollInterval)
				if err != nil {
					// This isn't really something that should cause the whole
					// secret to error out, it just won't get refetched.
					v.logger.WithError(err).Error("Secret looked like a KV V2 secret but has an unexpected form")
				} else {
					go watcher.Run()
					v.customWatchersByVaultPath[vaultPath] = watcher
				}
			}
		}
		v.secretsByVaultPath[vaultPath] = secret
	}

	for _, w := range secret.Warnings {
		v.logger.Warnf("Warning received for Vault secret at path %s: %s", vaultPath, w)
	}

	if val := traverseToKey(secret.Data, key); val != nil {
		v.logger.Debugf("Fetched secret at %s -> %s", vaultPath, key)

		return map[string][]byte{
			path: []byte(fmt.Sprintf("%#v", val)),
		}, 0, nil
	}

	return nil, 0, fmt.Errorf("no key %s found in Vault secret %s", key, vaultPath)
}

// Vault doesn't have a "watch" concept but we do have to renew tokens, so
// watch for errors doing that.
func (v *vaultConfigSource) WaitForChange(path string, version uint64, stop <-chan struct{}) error {
	vaultPath, _, err := splitConfigPath(path)
	if err != nil {
		return err
	}
	renewer := v.renewersByVaultPath[vaultPath]

	var watchErr error

	if renewer == nil {
		refetchTime := v.nonRenewableVaultPathRefetchTimes[vaultPath]
		if refetchTime.IsZero() {
			customWatcher := v.customWatchersByVaultPath[vaultPath]
			if customWatcher == nil {
				// There is nothing to do except wait for the whole thing to
				// stop
				<-stop
			} else {
			WATCHER:
				for {
					select {
					case <-stop:
						break WATCHER
					case <-customWatcher.ShouldRefetchCh():
						break WATCHER
					case err := <-customWatcher.ErrorCh():
						v.logger.WithError(err).WithFields(log.Fields{
							"vaultPath": vaultPath,
						}).Error("Error watching secret for change")
						// We don't really want to make it refetch the secret
						// if an error occurs in a watcher, just log it and
						// repeat this select.
						continue WATCHER
					}
				}
				customWatcher.Stop()
			}
		} else {
			timer := time.NewTimer(time.Until(refetchTime))
			defer timer.Stop()
			select {
			case <-stop:
				break
			case <-timer.C:
				break
			}
		}
	} else {
		select {
		// This will receive if there are an errors renewing a secret lease
		case watchErr = <-renewer.DoneCh():
			break
		case <-stop:
			renewer.Stop()
		}
	}

	v.Lock()
	// Wipe the secret from the cache so that it gets refetched
	delete(v.renewersByVaultPath, vaultPath)
	delete(v.secretsByVaultPath, vaultPath)
	delete(v.nonRenewableVaultPathRefetchTimes, vaultPath)
	delete(v.customWatchersByVaultPath, vaultPath)
	v.Unlock()

	v.logger.Debugf("Path changed or failed to renew: %s", vaultPath)

	return watchErr
}

func (v *vaultConfigSource) Stop() error {
	v.Lock()
	defer v.Unlock()

	if v.tokenRenewer != nil {
		v.tokenRenewer.Stop()
	}
	for p := range v.renewersByVaultPath {
		v.renewersByVaultPath[p].Stop()
	}

	for p := range v.customWatchersByVaultPath {
		v.customWatchersByVaultPath[p].Stop()
	}

	return nil
}
