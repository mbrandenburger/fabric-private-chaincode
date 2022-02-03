#!/bin/bash -eu
#
# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

GOROOT='/opt/go'
GO_VERSION=1.16.7

# ----------------------------------------------------------------
# Install Golang
# ----------------------------------------------------------------
GO_TAR=go${GO_VERSION}.linux-amd64.tar.gz
GO_URL=https://dl.google.com/go/${GO_TAR}
mkdir -p $GOROOT
curl -sL "$GO_URL" | (cd $GOROOT && tar --strip-components 1 -xz)

# ----------------------------------------------------------------
# Setup environment
# ----------------------------------------------------------------
cat <<EOF > /etc/profile.d/goroot.sh
export GOROOT=$GOROOT
export PATH=\$PATH:$GOROOT/bin
EOF
