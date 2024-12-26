#!/usr/bin/env bash

sudo systemctl stop stsb
sudo rm -rf /var/lib/stsb
sudo rm -rf /usr/local/bin/stsb
sudo rm -rf /usr/local/etc/stsb
sudo rm -rf /etc/systemd/system/stsb.service
sudo systemctl daemon-reload
