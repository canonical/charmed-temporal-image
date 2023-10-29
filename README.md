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
which will allow us to run the docker container locally as well deploy our charm
using this custom image.

1. Set the necessary authorization environment variables in the
   [`development.yaml`](./temporal-server/config/development.yaml) file. This
   will require:

   - Setting up an [OpenFGA store](https://charmhub.io/openfga-k8s).
   - Setting up a
     [Google Cloud project](https://developers.google.com/workspace/guides/create-project).

2. Clone the main [Temporal repository](https://github.com/temporalio/temporal)
   and add the necessary environment variables to the
   [`temporal-ui` application](https://github.com/temporalio/temporal/blob/main/develop/docker-compose/docker-compose.yml)
   as shown below.

```bash
git clone org-56493103@github.com:temporalio/temporal.git

 environment:
   - TEMPORAL_ADDRESS=localhost:7233
   - TEMPORAL_CORS_ORIGINS=http://localhost:3000
   - TEMPORAL_AUTH_ENABLED=true
   - TEMPORAL_AUTH_PROVIDER_URL=https://accounts.google.com
   - TEMPORAL_AUTH_CLIENT_ID=<GOOGLE_CLIENT_ID>
   - TEMPORAL_AUTH_CLIENT_SECRET=<GOOGLE_SECRET_ID>
   - TEMPORAL_AUTH_CALLBACK_URL=http://localhost:8080/auth/sso/callback # This must be included as a callback URL to your Google IAM project.
   - TEMPORAL_AUTH_SCOPES=openid,profile,email

```

You can then start the necessary dependencies:

```bash
make start-dependencies
make install-schema
```

Note: this step is only necessary if you are testing the image outside of the
charmed Temporal ecosystem. If testing the image using the Temporal server
charm, you must set up the necessary charms according to the
[Temporal charms documentation](https://charmhub.io/temporal-k8s/docs/t-introduction)
before proceeding.

3. Enable the microk8s registry:

```bash
microk8s enable registry
```

4. Build the custom image in this repository and push it to the local microk8s
   registry:

```bash
docker build . -t localhost:32000/temporal-auth
docker images # make note of the image ID
docker tag <IMAGE_ID> localhost:32000/temporal-auth
docker push localhost:32000/temporal-auth
```

5. To run your container, use the following command:

```bash
docker run -d -p 7233:7233 --name temporal-auth localhost:32000/temporal-auth
```

You should now be able to run sample workflows using the Temporal
[Python](https://github.com/canonical/temporal-lib-py) and
[Go](https://github.com/canonical/temporal-lib) client libraries. Each library
contains instructions on how to set the necessary configuration variables for
Google IAM authentication.

6. (Optional) To test this custom image on the
   [Temporal server charm](https://github.com/canonical/temporal-k8s-operator),
   you can set up the
   [charmed Temporal ecosystem](https://charmhub.io/temporal-k8s/docs/t-introduction)
   and attaching the image as a resource by running the following:

```bash
git clone git@github.com:canonical/temporal-k8s-operator.git
charmcraft pack && juju deploy ./temporal-k8s_ubuntu-22.04-amd64.charm --resource temporal-server-image=localhost:32000/temporal-auth
```
