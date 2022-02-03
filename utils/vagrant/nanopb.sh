#!/bin/bash -eu
#
# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

NANOPB_PATH=/usr/local/nanopb/
NANOPB_VERSION=0.4.3

export DEBIAN_FRONTEND=noninteractive

# Install protobuf
apt-get -qq update
apt-get install -y \
    python \
    protobuf-compiler \
    python-protobuf

# Install nanopb
git clone \
  -c advice.detachedHead=false \
  https://github.com/nanopb/nanopb.git \
  ${NANOPB_PATH}

cd ${NANOPB_PATH} \
  && git checkout nanopb-${NANOPB_VERSION} \
  && cd generator/proto \
  && make
echo "export NANOPB_PATH=${NANOPB_PATH}" >> /etc/profile.d/nanopb.sh
