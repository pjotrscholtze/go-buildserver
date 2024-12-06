#!/bin/sh
echo $CONFIG_PATH
mkdir -p /tmp/go-buildserver
eval `ssh-agent`
ssh-agent &

# With migrations
# /app/go-buildserver 1 $CONFIG_PATH

# Without migrations
/app/go-buildserver 0 $CONFIG_PATH
