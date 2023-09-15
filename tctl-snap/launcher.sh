#! /bin/bash

mkdir -p $SNAP_USER_DATA/.config/temporalio
if [[ ! -f $SNAP_USER_DATA/.config/temporalio/tctl.yaml ]]
then
  echo $'active: local\naliases: {}\nversion: 2\n' > $SNAP_USER_DATA/.config/temporalio/tctl.yaml
fi

exec "$@"
