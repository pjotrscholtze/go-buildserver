#!/bin/sh
echo $CONFIG_PATH
mkdir -p /tmp/go-buildserver
eval `ssh-agent`
ssh-agent &

/app/go-buildserver 0 $CONFIG_PATH
