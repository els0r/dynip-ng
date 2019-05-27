#!/bin/bash

set -eufo pipefail

BASE=$PWD/absolute

rm -rf ${BASE}

mkdir -p ${BASE}/opt/dynip/bin \
		 ${BASE}/opt/dynip/etc \
	     ${BASE}/opt/dynip/var \
	     ${BASE}/etc/systemd/system

cp $GOBIN/dynip-ng ${BASE}/opt/dynip/bin

# systemd service
echo "[Unit]
Description=Dynamic IP Updater
After=network.target syslog.target

[Service]
ExecStart=/opt/dynip/bin/dynip-ng run -c /opt/dynip/etc/dynip-ng.yml
StandardOutput=syslog
Restart=on-failure

[Install]
WantedBy=multi-user.target
Alias=dnyip.service
" > ${BASE}/etc/systemd/system/dynip.service

# write example config
echo "iface: eth0 			# interface to listen on
zone: example.com 			# zone to change record in
record: dyn-ip	  			# name or DNS record to change
interval: 5		  			# check every 5 minutes
state_file: /opt/dynip/var/.dyn-ip 		# state file
cloudflare_api:
    key: 13550350a8681c84c861aac2e5b440161c2b33a3e4f302ac680ca5b686de48de
    email: user@example.com
" > ${BASE}/opt/dynip/etc/dynip-ng.yml.example \

# create an archive
tar cjf dynip.tar.bz2 ${BASE} 2>&1 > /dev/null

