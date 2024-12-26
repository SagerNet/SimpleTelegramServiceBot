#!/usr/bin/env bash

set -e -o pipefail

if [ -d /usr/local/go ]; then
  export PATH="$PATH:/usr/local/go/bin"
fi

DIR=$(dirname "$0")
PROJECT=$DIR/../..

pushd $PROJECT
go install -v -trimpath -ldflags "-s -w -buildid=" .
popd

sudo systemctl stop stsb
sudo cp $(go env GOPATH)/bin/stsb /usr/local/bin/
sudo systemctl start stsb
