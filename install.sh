#!/bin/bash

set -eufo pipefail

BASE=./absolute

rm -rf ${BASE}

mkdir -p ${BASE}/opt/dynip/bin \
		 ${BASE}/opt/dynip/etc \
	     ${BASE}/opt/dynip/var \
	     ${BASE}/etc/systemd/system

echo "*** installing binary ***"
cd pkg/version
export COMMIT_SHA=$( git rev-parse HEAD )
go generate
cd ../../

go install
cp $GOBIN/dynip-ng ${BASE}/opt/dynip/bin

echo "*** writing system files ***"
cp addon/dynip.service ${BASE}/etc/systemd/system/dynip.service
cp addon/dynip-ng.yml.example ${BASE}/opt/dynip/etc/dynip-ng.yml.example

echo "*** creating deployable archive ***"
# create an archive
tar cjf dynip.tar.bz2 ${BASE} 2>&1 > /dev/null
