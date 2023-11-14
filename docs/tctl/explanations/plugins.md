# TCTL & Plugins

[TCTL](https://docs.temporal.io/tctl/) is a command-line tool that can be used
to interact with the Temporal Server in a multitude of ways.

We are also building our own flavor of TCTL with a couple of plugins to
facilitate interacting with our Temporal Server with custom auth. For that, we
are adding 2 extra plugins:

- tctl-login
- tctl-authorization

## tctl-login

The tctl-login plugin is a simple binary that starts a web browser to perform
Google authentication using the provided Google client ID and client secret,
then fetches Google OAuth access and refresh tokens, which it then stores in the
local filesystem under `/home/<user>/snap/tctl/current/`.

The binary that is built will have the name `tctl-login`. By default, the `tctl`
tool will look in the path for executables named `tctl-<name>` and will call
them when one calls `tctl <name>`. As such, there's no special configuration
that is required to make `tctl` work with this plugin.

## tctl-authorization

The tctl-authorization plugin is a more interesting binary, which uses the
https://github.com/temporalio/tctl/cli/plugin mechanism, which underneath uses
the https://github.com/hashicorp/go-plugin mechanism. This essentially sets up a
small GRPC server in the binary, and then the actual TCTL command performs a
GRPC call to this plugin to perform whatever operation is needed.

In our case, the procedure that is called is a method which retrieves the access
token from the local filesystem that was created by `tctl-login` and sets it as
an `Authorization` header to the request towards Temporal Server.

This plugin must be provided to the `tctl` tool by configuring the following env
variable: `TEMPORAL_CLI_PLUGIN_HEADERS_PROVIDER=tctl-authorization`, which is
done automatically through the snap configuration. Below are some configurations
that must be manually set by the user to use the tctl snap depending on the
environment:

- `stg-google-client-id`
- `stg-google-client-secret`
- `prod-google-client-id`
- `prod-google-client-secret`

More information can be found [here](../../../tctl-snap/README.md).

## Packaging

Because of the custom functionality, there are a couple of config values which
need to be set depending on the environment one wishes to run tctl commands
against. We have packaged the TCTL binary in a
[Snap package](../../../tctl-snap/) that reduces a bit of the overhead that is
required to use it.
