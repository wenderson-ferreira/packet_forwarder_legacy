#!/bin/bash

if [[ -z "$1" ]] ; then
    # No binary specified
    echo "$0: No binary specified."
    exit 1
fi

binary="$1"
binary_name=`basename "$binary"`
binary_directory=`dirname "$binary"`
pushd "$binary_directory"
absolute_binary_directory="$(pwd)"
absolute_binary_path="$absolute_binary_directory/$binary_name"
popd

config="$2"

if [[ -z "$config" ]] ; then
    echo "$0: No configuration file to use specified."
    exit 1
fi

config_name=`basename "$config"`
config_directory=`dirname "$config"`
pushd "$config_directory"
absolute_config_directory="$(pwd)"
absolute_config_path="$absolute_config_directory/$config_name"
popd

echo "[Unit]
Description=TTN Packet Forwarder Service

[Install]
WantedBy=multi-user.target

[Service]
TimeoutStartSec=infinity
Type=simple
TimeoutSec=infinity
RestartSec=10
WorkingDirectory=$absolute_binary_directory
ExecStart=$absolute_binary_path start --config=\"$absolute_config_path\"
Restart=always
BusName=org.thethingsnetwork.ttn-pkt-fwd" > /etc/systemd/system/ttn-pkt-fwd.service

echo "$0: Installation of the systemd service complete."
