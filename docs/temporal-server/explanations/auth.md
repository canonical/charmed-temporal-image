# Temporal Server Auth

As mentioned in the [architecture doc](./architecture.md), Temporal Server by
default does not offer any auth, just plugin mechanisms through which this can
be added.

We have added an OAuth-based authentication with
[Google Cloud](https://cloud.google.com), leveraging
[OpenFGA](https://openfga.dev/) for authorization.

## OAuth-based Authentication

Requests made to the Temporal Server when auth is enabled must include an
`Authorization` header with a valid Google OAuth access token. This can either
be done by:

- A login through the Temporal web UI, which can be enabled as outlined
  [here](../../../README.md) or setting up the charmed
  [Temporal k8s web UI](https://charmhub.io/temporal-ui-k8s).
- The use of a Google Cloud
  [service account](https://cloud.google.com/iam/docs/service-account-overview),
  which can be enabled using the
  [temporal-lib-go](https://github.com/canonical/temporal-lib-go) and
  [temporal-lib-py](https://github.com/canonical/temporal-lib-py) client
  libraries.

If the client manages to authenticate through Google OAuth, it will receive the
necessary access token back, which it must present to the Temporal Server on the
`Authorization` header of the request, which the Server will then use to perform
the Authorization, as explained below.

## Temporal Server Implementation

### Token Verification

On every request, the Temporal Server verifies the received access token against
Google OAuth servers to validate the client's identity. Once done, the custom
authorizer described below determines what namespaces the client is assigned
access to along with the respective role.

### Authorization Plugins

Inside the Temporal Server, we "inject" two plugins that are offered by the
Temporal Server, namely a `ClaimMapper` and an `Authorizer`. Further details
about the two interfaces can be found
[here](https://docs.temporal.io/server/security/#authorization).

#### ClaimMapper

The **ClaimMapper** that we provide is implemented by the **TokenClaimMapper**.
It expects to find an `Authorization` header with the contents in the format of
`Bearer <token>`, where the `<token>` is a valid Google OAuth access token.

It validates the token, and uses the `email` field in the token to query the
OpenFGA store for namespace access. This is first done by querying all the
groups that the user with the given email is a member of, and then querying for
each of these groups all the namespaces they are related to. For example, If
user `john` is a member of group `abc`, and group `abc` is related to namespace
`example` as a `writer`, then user `john` will be assigned `RoleWriter` on
namespace `example`. This is essentially how we achieve multi-tenancy, since
Temporal namespaces cannot exchange any information between them.

As a special config, we allow the specification of a set of groups that, if
users belong to any of them, they have full access to the entire System. These
are like super-admin groups. The config is called `adminGroups`, further on this
in the Config section below.

Temporal offers 4 different roles: `Admin`, `Writer`, `Reader` and `Worker`.
However, they do not have any inherent meaning. It is the `Authorizer` that
attaches meaning to them, as we shall see next.

#### Authorizer

The **Authorizer** receives the map of claims that the **ClaimMapper** builds
and issues an allow/deny decision based on it. It does not run any further
checks, but trusts that map completely (any actual auth verifications were done
by the ClaimMapper already).

The **Authorizer** we provide currently follows somewhat closely the default one
that is provided by the Temporal Server. We have a hardcoded list of read-only
APIs, for which we require the role of `Reader`. For anything else, we require
the role of `Writer`. If the request targets a particular namespace, we require
the role on that particular namespace, otherwise we require it on the special
`System`-wide component.

### Config

On top of Temporal Server's usual suite of configs, we've also added a new
section called `auth`, with the following structure:

```yaml
auth:
  enabled: { { .AUTH_ENABLED } }
  adminGroups: { { .ADMIN_GROUPS } }
  openAccessNamespaces: { { .OPEN_ACCESS_NAMESPACES } }
  googleClientID: { { .GOOGLE_CLIENT_ID } }
  ofga:
    apiScheme: { { .OFGA_API_SCHEME } }
    apiHost: { { .OFGA_API_HOST } }
    apiPort: { { .OFGA_API_PORT } }
    token: { { .OFGA_TOKEN } }
    storeID: { { .OFGA_STORE_ID } }
    authModelID: { { .OFGA_AUTH_MODEL_ID } }
```

Where:

- `enabled` is self explanatory.
- `adminGroups` are special groups that, if any user belongs to them, will have
  full admin rights in Temporal Server.
- `openAccessNamespaces` are special namespaces on which all users have write
  access.
- `googleClientID` is the client ID of the Google Cloud project used to handle
  authentication for the project. This is the same Google Cloud project which
  will be used to authenticate users through the Web UI as well as generate any
  service accounts.
- `ofga` contains all the parameters needed to communicate with an OpenFGA
  store, which must contain a valid authorization model.
