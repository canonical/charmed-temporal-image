# Temporal Server

The
[Temporal Server](https://docs.temporal.io/concepts/what-is-the-temporal-server/)
is the "headquarters" or the "central brains" of the entire suite. It offers a
GRPC interface to the exterior, that is leveraged on one hand by the UI & TCTL
for management and on another hand by workers for actual workflow execution.

The original repository of the Temporal Server can be found
[here](https://github.com/temporalio/temporal).

The **Temporal Server** is further made up of 5 internal components:

- Frontend gateway: for rate limiting, routing, authorizing.
- Internal frontend gateway: to allow the internal `worker` service to connect
  without authorization.
- History subsystem: maintains data (mutable state, queues, and timers).
- Matching subsystem: hosts Task Queues for dispatching.
- Worker Service: for internal background Workflows.

All the 5 components are deployed with the help of the same build / Docker
image, just with different configuration parameters.

The Worker Service executes workflows through the Internal frontend gateway.

The Server requires a connection to a database (Cassandra, MySQL and Postgres
supported)

## Authorization

By default the Temporal Server doesn't offer any authorization, but it offers
the plugin mechanism to add a custom one.

We have added an OAuth-based authentication using Google Cloud and an
authorization mechanism that leverages [Google Cloud](https://cloud.google.com)
and [OpenFGA](https://openfga.dev/). This is further explained in the
[Temporal Auth doc](./auth.md).

## Encryption

The Temporal Server will in its operation log and save workflow and activity
related information which includes the parameters passed to activities and the
returned information. Although these are only accessible to people that hold
credentials to a particular namespace, one might want to add an extra layer of
security via encryption. As such, Temporal supports E2E encryption of all data
that is passed to the Temporal Server, such that only the workers will know the
exact data and the Server will only have encrypted versions of it. This works
transparently for the Temporal Server. Therefore, to leverage this
functionality, please refer to the
[temporal-lib-go](https://github.com/canonical/temporal-lib-go) and
[temporal-lib-py](https://github.com/canonical/temporal-lib-py) client libraries
where this is implemented.

## Code Layout

The Temporal Server is built from the
[temporal-server](../../../temporal-server/) subdirectory. It contains a custom
main function that starts the Temporal Server from a Go module / library as well
as setting up the auth plugins, which are explained in the
[Temporal Auth doc](./auth.md).

There is a [Dockerfile](../../../Dockerfile) for building the Docker image.
