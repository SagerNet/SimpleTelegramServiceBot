#!/usr/bin/env bash

set -e -o pipefail

sudo systemctl enable stsb
sudo systemctl start stsb
sudo journalctl -u stsb --output cat -f
