# TCTL

This snap can be used to run authorized tctl commands with pre-specified
environment variables for both staging and production environments.

By setting the environment variables `TEMPORAL_CLI_ADDRESS`, `GOOGLE_CLIENT_ID`
and `GOOGLE_CLIENT_SECRET` in `snap/snapcraft.yaml`, you can enable user login
through Google, which will generate a user-specific access token to be attached
with each request made to the Temporal server. For more information on how to
obtain Google credentials for your project, visit
[Google Cloud Platform Help](https://support.google.com/cloud/answer/6158849?hl=en#zippy=%2Cnative-applications%2Cdesktop-apps).

The snap can be installed as follows:

```bash
# Build snap
make all

# Install snap
sudo snap install tctl_next_amd64.snap --dangerous [--devmode]
```

Once installed, you can run tctl commands using `tctl.stg` for staging and
`tctl.prod` for production environments. More information can be found
[here](../docs/tctl/howtos/basic-operations.md).

Note: for vanilla-tctl which does not have an authorization plugin, visit
Temporal's [tctl repository](https://github.com/temporalio/tctl).
