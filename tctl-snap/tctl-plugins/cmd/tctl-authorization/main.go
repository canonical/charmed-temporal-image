package main

import (
	"context"
	"fmt"
	"os"

	"github.com/canonical/charmed-temporal-image/tctl-plugins/cmd"
	"github.com/hashicorp/go-plugin"
	cliplugin "github.com/temporalio/tctl/cli/plugin"
)

type provider struct{}

func (p provider) GetHeaders(ctx context.Context) (map[string]string, error) {
	env := os.Getenv("TCTL_ENVIRONMENT")
	if env == "dev" {
		return map[string]string{}, nil
	}

	clientID, err := cmd.ClientID()
	if err != nil {
		return map[string]string{}, err
	}

	if clientID == "" {
		fmt.Fprintf(os.Stderr, "no google-client-id found for %v environment. use 'sudo snap set tctl %v-google-client-id=\"<client_id>\"'.\n", env, env)
		return map[string]string{}, nil
	}

	clientSecret, err := cmd.ClientSecret()
	if err != nil {
		return map[string]string{}, err
	}

	if clientSecret == "" {
		fmt.Fprintf(os.Stderr, "no google-client-secret found for %v environment. use 'sudo snap set tctl %v-google-client-secret=\"<client_secret>\"'.\n", env, env)
		return map[string]string{}, nil
	}

	token, err := cmd.FetchValidToken(clientID, clientSecret)
	if err != nil {
		return map[string]string{}, err
	}

	return map[string]string{
		"Authorization": "Bearer " + token,
	}, nil
}

func main() {
	var pluginMap = map[string]plugin.Plugin{
		cliplugin.HeadersProviderPluginType: &cliplugin.HeadersProviderPlugin{
			Impl: &provider{},
		},
	}
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: cliplugin.PluginHandshakeConfig,
		Plugins:         pluginMap,
	})
}
