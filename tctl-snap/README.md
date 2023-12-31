# TCTL

This snap can be used to run authorized tctl commands with pre-specified
environment variables for both staging and production environments.

By setting the snap variables, `google-client-id` and `google-client-secret` as
described below, you can enable user login through Google, which will generate a
user-specific access token to be attached with each request made to the Temporal
server. For more information on how to obtain Google credentials for your
project, visit
[Google Cloud Platform Help](https://support.google.com/cloud/answer/6158849?hl=en#zippy=%2Cnative-applications%2Cdesktop-apps).

The snap can be installed as follows:

```bash
# Build snap
make all

# Install snap
sudo snap install tctl_next_amd64.snap --dangerous
```

tctl commands can be run in the following pre-defined three environments:

- `dev`: This environment does not have the authorization plugin enabled and
  does not require login. It will send commands to the Temporal server with an
  empty authorization header. The following is an example command:

  ```bash
  tctl.dev namespace list
  ```

- `stg`: This is a staging environment with the authorization plugin enabled. It
  will require a Google client ID and Google client secret to be set as follows:

  ```bash
  sudo snap set tctl stg-google-client-id="<client_id>"
  sudo snap set tctl stg-google-client-secret="<client_secret>"
  ```

  tctl commands can then be run as follows:

  ```bash
  tctl.stg login

  # Include Temporal server address flag in the command
  tct.stg --address=localhost:7233 namespace list
  ```

- `prod`: This is a production environment with the authorization plugin
  enabled. It will require a Google client ID and Google client secret to be set
  as follows:

  ```bash
  sudo snap set tctl prod-google-client-id="<client_id>"
  sudo snap set tctl prod-google-client-secret="<client_secret>"
  ```

  tctl commands can then be run as follows:

  ```bash
  tctl.prod login

  # Include Temporal server address flag in the command
  tct.prod --address=localhost:7233 namespace list
  ```

  Some sample operations that can be run in any environment also include:

  ```bash
  # Register namespace
  tctl.prod namespace register <name>

  # Describe namespace
  tctl.prod namespace describe <name>

  # List workflows in namespace
  tctl.prod -n <name> --workflow-id <workflow-id>
  ```

  Other commands can be found [here](https://docs.temporal.io/tctl-v1).
