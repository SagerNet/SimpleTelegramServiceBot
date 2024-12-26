#!/usr/bin/env bash

set -e -o pipefail

if [ -d /usr/local/go ]; then
  export PATH="$PATH:/usr/local/go/bin"
fi

DIR=$(dirname "$0")
PROJECT=$DIR/../..

pushd $PROJECT
go install -v -trimpath -ldflags "-s -w -buildid=" ./
popd

sudo cp $(go env GOPATH)/bin/stsb /usr/local/bin/
sudo mkdir -p /usr/local/etc/stsb
sudo cp $PROJECT/release/config/config.json /usr/local/etc/stsb/config.json
sudo cp $DIR/stsb.service /etc/systemd/system
sudo systemctl daemon-reload
