# Charmed Temporal Image

This repository contains a Temporal Server image with custom authentication via
Google IAM and authorization via OpenFGA. The image is designed for use with the
[Charmed Temporal K8s Operator](https://charmhub.io/temporal-k8s).

The main components that are handled in this repository are:

- **Temporal Server** - the main component of the Temporal architecture.
- **TCTL** - command line tool for interacting with the Temporal Server.

## Further reading

Since the repository contains two components, each of them is documented
individually. Please refer to the [Documentation Index](./docs/index.md) for
further reading.

## Try it out

To test the custom Temporal Server locally, we will be using
[microk8s](https://microk8s.io/docs/registry-built-in) as a local registry,
which will allow us to deploy our charm using this custom image.

1. Set up a
   [Google Cloud project](https://developers.google.com/workspace/guides/create-project).
   You will then need to set up an
   [OAuth 2.0 client ID](https://support.google.com/cloud/answer/6158849?hl=en).
   This will be used to acquire the credentials needed to set up authentication
   through the web UI and client libraries.

2. Set up the
   [charmed Temporal ecosystem](https://charmhub.io/temporal-k8s/docs/t-introduction).

3. Enable authentication on the Temporal Web UI charm as follows:

   ```bash
   juju config temporal-ui-k8s auth-enabled=true auth-client-id="<google_client_id>" auth-secret-id="<google_secret_id>"
   ```

4. Enable the microk8s registry:

   ```bash
   microk8s enable registry
   ```

5. Build the custom image in this repository and push it to the local microk8s
   registry:

   ```bash
   docker build . -t localhost:32000/temporal-auth
   docker images # make note of the image ID
   docker tag <IMAGE_ID> localhost:32000/temporal-auth
   docker push localhost:32000/temporal-auth
   ```

6. Attach the image as a resource to the server charm by running the following:

   ```bash
   juju refresh temporal-k8s --resource temporal-server-image=localhost:32000/temporal-auth
   ```

You should now be able to run sample workflows using the Temporal
[Python](https://github.com/canonical/temporal-lib-py) and
[Go](https://github.com/canonical/temporal-lib-go) client libraries against the
deployed Temporal server. Each library contains instructions on how to set the
necessary configuration variables for Google IAM authentication.
