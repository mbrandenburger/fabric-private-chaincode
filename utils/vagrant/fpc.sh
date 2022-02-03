#!/bin/bash -eu
#
# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

FABRIC_REPO=https://github.com/hyperledger/fabric.git
FABRIC_VERSION=2.3.3

FABRIC_REL_PATH='src/github.com/hyperledger/fabric'
FPC_REL_PATH='src/github.com/hyperledger/fabric-private-chaincode'

GOPATH="/home/vagrant/projects"
FPC_PATH="${GOPATH}/${FPC_REL_PATH}"
FPC_VERSION=main
SGX_MODE=SIM

# get fabric
FABRIC_PATH=${GOPATH}/${FABRIC_REL_PATH}
git clone \
  -c advice.detachedHead=false \
  --branch v${FABRIC_VERSION} \
  ${FABRIC_REPO} \
  ${FABRIC_PATH}

cat <<EOF >> /home/vagrant/.bashrc
export GOPATH=${GOPATH}
export PATH=\${PATH:+\${PATH}:}\$GOPATH/bin
export FPC_PATH=${FPC_PATH}
export FPC_VERSION=${FPC_VERSION}
export SGX_MODE=${SGX_MODE}
export FABRIC_PATH=${FABRIC_PATH}

cd \$FPC_PATH
EOF
