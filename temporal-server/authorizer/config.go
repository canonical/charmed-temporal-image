package authorizer

import (
	"fmt"

	"go.temporal.io/server/common/config"
)

// ConfigWithAuth is a wrapper over Temporal Server's config.Config which also
// adds custom Auth fields.
type ConfigWithAuth struct {
	*config.Config `yaml:",inline"`

	Auth Auth `yaml:"auth"`
}

// Auth contains all the configuration and secrets required to perform
// authentication and authorization functionalities in the Temporal Server
// service.
type Auth struct {
	Enabled              bool                `yaml:"enabled"`
	OFGA                 AuthorizationConfig `yaml:"ofga"`
	AdminGroup           string              `yaml:"adminGroup"`
	OpenAccessNamespaces string              `yaml:"openAccessNamespaces"`
	GoogleClientID       string              `yaml:"googleClientID"`
}

// AuthorizationConfig holds the configuration required for communicating with
// the authorization service.
type AuthorizationConfig struct {
	// APIScheme is either "http" or "https".
	APIScheme string `yaml:"apiScheme"`
	// APIHost is the host URL of the authorization service minus the scheme.
	APIHost string `yaml:"apiHost"`
	// APIPort is the port on the host machine where the authorization service is
	// running.
	APIPort string `yaml:"apiPort"`
	// BearerToken is the token attached to the authorization header for requests
	// made to the OpenFGA store.
	BearerToken string `yaml:"token"`
	// StoreID is the ID of the auth store defined in the authorization service
	// that must be used for auth checks.
	StoreID string `yaml:"storeID"`
	// AuthModelID is the ID of the defined authorization model that is
	// currently being used for authorization checks.
	AuthModelID string `yaml:"authModelID"`
}

// LoadConfigWithAuth loads a config yaml from the given directory. The expected
// structure of the file is the one respresented by ConfigWithAuth.
func LoadConfigWithAuth(env string, configDir string, zone string) (*ConfigWithAuth, error) {
	cfg := ConfigWithAuth{}
	err := config.Load(env, configDir, zone, &cfg)
	if err != nil {
		return nil, fmt.Errorf("config file corrupted: %w", err)
	}
	return &cfg, nil
}
