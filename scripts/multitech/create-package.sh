#!/bin/bash
#
# Parts of the script based on installer.sh by Jac Kersing <j.kersing@the-box.com>
#
# What this script does:
# - Creates a configuration file
# - Generates a script and a readme file
# - Packages everything with the kerlink build for DOTA install
# Possible improvements:
# - Custom certificate address
# - TTN-hosted static url for produsb.sh and the Let's Encrypt cert

if [[ -z "$(which tar)" ]] ; then
    echo "$0: tar required to run this script."
    exit 1
fi

multitech_installer_file=$(echo "$0" | grep -o '^.*\/')/multitech-installer.sh
if [[ ! -f "$multitech_installer_file" ]] ; then
    echo "$0: Can't find multitech-installer.sh at $multitech_installer_file, please check and restart this script."
    exit 1
fi

usage_str="usage: create-kerlink-package.sh [path to the Multitech build] ([Gateway ID] [Gateway key])

example: ./create-kerlink-package.sh packet-forwarder-multitech"

# Getting path to the kerlink binary
INITIAL_BINARY_PATH="$1"
if [[ -z "$INITIAL_BINARY_PATH" ]] ; then
    echo "$0: $usage_str"
    exit 1
fi

if [[ $(which openssl) =~ "not found" ]] ; then
    random_string="-$(< /dev/urandom tr -dc _A-Z-a-z-0-9 | head -c${1:-32};echo;)"
else
    random_string="-$(openssl rand -base64 15)"
    random_string="${random_string//\/}"
fi

gatewayID="$2"
gatewayKey="$3"
if [[ ! -z "$4" ]] ; then
    OUTPUT="/dev/null"
else
    OUTPUT="/dev/stdout"
fi

echo "$0: Creating file tree" &> "$OUTPUT"

WORKDIR="/tmp/packet-forwarder-multitech-$random_string"
BASE="/usr/bin"
PKTFWD_DESTDIR="$WORKDIR$BASE"
BINARY_NAME="ttn-pkt-fwd"

mkdir -p "$PKTFWD_DESTDIR" # /usr/bin
mkdir -p "$WORKDIR/etc/init.d" # /etc/init.d
mkdir -p "$WORKDIR/usr/cfg"
local_config_file="/usr/cfg/config.yml"
config_file="$WORKDIR$local_config_file"
touch "$config_file"
cp "$INITIAL_BINARY_PATH" "$PKTFWD_DESTDIR/$BINARY_NAME"
chmod +x "$PKTFWD_DESTDIR/$BINARY_NAME"

configure () {
    printf "%s: Gateway ID:\n> " "$0"

    read -r gatewayID

    printf "%s: Gateway Key:\n> " "$0"

    read -r -s gatewayKey
}

if [[ -z "$gatewayID" && -z "$gatewayKey" ]] ; then
    echo "$0: If you haven't registered your gateway yet, register it on the console or with \`ttnctl\`, using the gateway connector protocol." &> "$OUTPUT"

    if [[ -f "$HOME/.pktfwd.yml" ]] ; then
        while true; do
            read -r -p "$0: Local packet forwarder configuration found (in $HOME/.pktfwd.yml). Do you want to include it in the package? " yn
            case $yn in
                [Yy]* ) cp "$HOME/.pktfwd.yml" "$config_file"; COPIED_CONFIG="1"; echo "$0: Local configuration included." &> "$OUTPUT"; break;;
                [Nn]* ) echo "$0: Local packet forwarder configuration not copied, please enter the new configuration." &> "$OUTPUT"; configure "$PKTFWD_DESTDIR/$CFG_FILENAME"; break;;
                * ) echo "Please answer [y]es or [n]o." &> "$OUTPUT";;
            esac
        done
    else
        configure
    fi

    echo "$0: Configuration packaged." &> "$OUTPUT"
fi

if [[ -z "$COPIED_CONFIG" ]] ; then
    echo "id: \"${gatewayID}\"
key: \"${gatewayKey}\"" > "$config_file"
fi

echo "$0: Generating control file" &> "$OUTPUT"
echo "Package: ttn-pkt-fwd
Version: 2.0.0
Description: TTN Packet Forwarder
Section: console/utils
Priority: optional
Maintainer: The Things Industries <eric@thethingsindustries.com>
License: MIT
Architecture: arm926ejste
OE: ttn-pkt-fwd
Homepage: https://github.com/TheThingsNetwork/packet_forwarder
Depends: libmpsse (>= 1.3), libc6 (>= 2.19)
Source: git://github.com/TheThingsNetwork/packet_forwarder.git;protocol=git" > "$WORKDIR/control"

echo "$0: Generating service script" &> "$OUTPUT"
echo "#!/bin/bash

NAME=\"$BINARY_NAME\"
ENABLED=\"yes\"

[ -f /etc/default/\$NAME ] && source /etc/default/\$NAME

run_dir=/var/run/ttn-pkt-fwd
conf_dir=/usr/cfg
pkt_fwd_dir=$BASE
pkt_fwd=\$pkt_fwd_dir/$BINARY_NAME
pkt_fwd_log=/var/log/ttn-pkt-fwd.log
pkt_fwd_pidfile=\$run_dir/ttn-pkt-fwd.pid

read_card_info() {
    # product-id of first lora card
    lora_id=\$(mts-io-sysfs show lora/product-id 2> /dev/null)
    lora_eui=\$(mts-io-sysfs show lora/eui 2> /dev/null)
    # remove all colons
    lora_eui_raw=\${lora_eui//:/}
}

card_found() {
    if [ \"\$lora_id\" = \"\$lora_us_id\" ] || [ \"\$lora_id\" = \"\$lora_eu_id\" ]; then
        echo \"Found lora card \$lora_id\"
        return 1
    else
        return 0
    fi
}

do_start() {
    read_card_info

    if ! card_found; then
        echo \"\$0: MTAC-LORA not detected\"
        exit 1
    fi

    # wait for internet connection to become available
    COUNTER=0
    while : ; do
	ping -c1 google.com > /dev/null 2> /dev/null
	if [ \$? -eq 0 ]
	then
		break
    else
        if [ \$COUNTER -gt 10 ] ; then
            echo \"Couldn't connect to Internet, aborting.\"
            exit 1
        fi
		echo \"No internet connection (\$COUNTER out of 10 tries), waiting...\"
		sleep 20
        let COUNTER=COUNTER+1
	fi
    done

    echo -n \"Starting \$NAME: \"
    mkdir -p \$run_dir

    start-stop-daemon --start --background --make-pidfile \
        --pidfile \$pkt_fwd_pidfile --exec \$pkt_fwd -- start --config=\$conf_dir/config.yml
    echo \"OK\"
}

do_stop() {
    echo -n \"Stopping \$NAME: \"
    start-stop-daemon --stop --quiet --oknodo --pidfile \$pkt_fwd_pidfile --retry 5
    rm -f \$pkt_fwd_pidfile
    echo \"OK\"
}

if [ \"\$ENABLED\" != \"yes\" ]; then
    echo \"\$NAME: disabled in /etc/default\"
    exit
fi

configure() {
    multitech-installer.sh
    mkdir -p \$conf_dir
    if [[ ! -f \"\$conf_dir/config.yml\" ]] ; then
        touch \"\$conf_dir/config.yml\"
        \$pkt_fwd configure \"\$conf_dir/config.yml\" --config=\"\$conf_dir/config.yml\"
    fi
    update-rc.d ttn-pkt-fwd defaults
    exit
}

case \"\$1\" in
    \"start\")
        do_start
        ;;
    \"stop\")
        do_stop
        ;;
    \"restart\")
        ## Stop the service and regardless of whether it was
        ## running or not, start it again.
        do_stop
        do_start
        ;;
    \"configure\")
        ## Configure the service
        configure
        ;;
    *)
        ## If no parameters are given, print which are avaiable.
        echo \"Usage: \$0 {start|stop|restart|configure}\"
        exit 1
        ;;
esac" > "$WORKDIR/etc/init.d/$BINARY_NAME"
chmod +x "$WORKDIR/etc/init.d/$BINARY_NAME"

echo "chmod +x \"$BASE/$BINARY_NAME\"
echo \"**********************************************
YOU NEED TO CONFIGURE YOUR GATEWAY BY EXECUTING /etc/init.d/$BINARY_NAME configure
**********************************************\"
update-rc.d -f ttn-pkt-fwd remove > /dev/null 2> /dev/null
update-rc.d ttn-pkt-fwd defaults 95 30 > /dev/null 2> /dev/null" > "$WORKDIR/postinst"
chmod +x "$WORKDIR/postinst"

cp "$multitech_installer_file" "$PKTFWD_DESTDIR"
chmod +x "$PKTFWD_DESTDIR/multitech-installer.sh"

FILENAME="ttn-pkt-fwd$random_string.ipk"

pushd "$WORKDIR" &> /dev/null
tar -czvf "data.tar.gz" "etc" "var" "usr" &> /dev/null
tar -czvf "control.tar.gz" "control" "postinst" &> /dev/null
tar -czvf "$FILENAME" "data.tar.gz" "control.tar.gz" &> /dev/null
popd &> /dev/null

release_folder="multitech-release$random_string"
mkdir "$release_folder"
mv "$WORKDIR/$FILENAME" "$PWD/$release_folder"

rm -rf "$WORKDIR"

echo "$0: package available at $PWD/$release_folder/$FILENAME" &> "$OUTPUT"

if [[ ! -z "$4" ]] ; then
    printf "%s/%s/%s" "$PWD" "$release_folder" "$FILENAME"
fi
