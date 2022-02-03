#!/bin/bash -eu
#
# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

UBUNTU_VERSION=20.04
UBUNTU_NAME=focal

SGX_VERSION=2.12
SGXSSL_VERSION=2.10_1.1.1g
OPENSSL_VERSION=1.1.1g
PROTO_VERSION=3.11.4

export DEBIAN_FRONTEND="noninteractive" TZ="UTC"

apt-get install -y -q \
    basez \
    ca-certificates \
    curl \
    gnupg2 \
    unzip \
    python \
    pkg-config \
    perl \
    wget

# Install SGX PSW packages
echo "deb [arch=amd64] https://download.01.org/intel-sgx/sgx_repo/ubuntu ${UBUNTU_NAME} main" >> /etc/apt/sources.list
wget -qO - https://download.01.org/intel-sgx/sgx_repo/ubuntu/intel-sgx-deb.key | apt-key add -
apt-get update -q
apt-get install -y -q --no-install-recommends \
    libssl-dev \
    libcurl4-openssl-dev \
    libprotobuf-dev \
    make \
    libsgx-urts \
    libsgx-launch \
    libsgx-epid

# Install SGX SDK
mkdir -p /opt/intel && cd /opt/intel
SGX_SDK_BIN_REPO=https://download.01.org/intel-sgx/sgx-linux/${SGX_VERSION}/distro/ubuntu${UBUNTU_VERSION}-server
SGX_SDK_BIN_FILE=$(cd /tmp; wget --spider --recursive --level=1 --no-parent ${SGX_SDK_BIN_REPO} 2>&1 | perl  -ne 'if (m|'${SGX_SDK_BIN_REPO}'/(sgx_linux_x64_sdk.*)|) { print "$1\n"; }')
wget -q ${SGX_SDK_BIN_REPO}/${SGX_SDK_BIN_FILE} \
  && chmod +x ${SGX_SDK_BIN_FILE} \
  && echo -e "no\n/opt/intel" | ./${SGX_SDK_BIN_FILE} \
  && rm ${SGX_SDK_BIN_FILE}

SGX_SDK='/opt/intel/sgxsdk'

cat <<EOF > /etc/profile.d/sgx.sh
export SGX_SDK=${SGX_SDK}
export PATH=\${PATH:+\${PATH}:}\$SGX_SDK/bin:\$SGX_SDK/bin/x64
export PKG_CONFIG_PATH=\$SGX_SDK/pkgconfig
export LD_LIBRARY_PATH=\${LD_LIBRARY_PATH:+\${LD_LIBRARY_PATH}:}\$SGX_SDK/sdk_libs
EOF
source /etc/profile.d/sgx.sh

# LVI mitigations, needed to compile sgxssl, requires a
#   recent version of binutils (>= 2.32). Ubuntu 18.04 only
#   has 2.30 but Intel ships binary distro for 2.32.51.20190719
#   As sgx ships tools also for 20.04, use these for simplicity
#   and uniformity reason
SGX_SDK_BINUTILS_REPO=https://download.01.org/intel-sgx/sgx-linux/${SGX_VERSION}
SGX_SDK_BINUTILS_FILE=$(cd /tmp; wget --spider --recursive --level=1 --no-parent ${SGX_SDK_BINUTILS_REPO} 2>&1 | perl  -ne 'if (m|'${SGX_SDK_BINUTILS_REPO}'/(as.ld.objdump.*)|) { print "$1\n"; }')
wget -q ${SGX_SDK_BINUTILS_REPO}/${SGX_SDK_BINUTILS_FILE} \
  && mkdir -p sgxsdk.extras \
  && cd sgxsdk.extras \
  && tar -zxf ../${SGX_SDK_BINUTILS_FILE} \
  && rm ../${SGX_SDK_BINUTILS_FILE} \
  && (cd /opt/intel/sgxsdk.extras/external/toolset/ && \
      for f in $(ls | grep -v ${UBUNTU_VERSION}); do rm -rf ${f}; done)
# Note: above install file contains binutitls for _all_ supported distros
#   and are fairly large, so clean out anything we do not need
echo "export PATH=/opt/intel/sgxsdk.extras/external/toolset/ubuntu${UBUNTU_VERSION}\${PATH:+:\${PATH}}" >> /etc/profile.d/sgx.sh
source /etc/profile.d/sgx.sh

# install SGX SSL
SGX_SSL='/opt/intel/sgxssl'
git clone   -c advice.detachedHead=false \
  'https://github.com/intel/intel-sgx-ssl.git' ${SGX_SSL} \
  && cd ${SGX_SSL} \
  && . ${SGX_SDK}/environment \
  && git checkout lin_${SGXSSL_VERSION} \
  && cd openssl_source \
  && wget -q https://www.openssl.org/source/openssl-${OPENSSL_VERSION}.tar.gz \
  && cd ../Linux \
  && make SGX_MODE=SIM DESTDIR=${SGX_SSL} all test \
  && make install
cat <<EOF >> /etc/profile.d/sgx.sh
export SGX_SSL=${SGX_SSL}
export OPENSSL_VERSION=${OPENSSL_VERSION}
export SGXSSL_VERSION=${SGXSSL_VERSION}
EOF

# install custom protoc
PROTO_DIR='/usr/local/proto3'
PROTO_ZIP=protoc-${PROTO_VERSION}-linux-x86_64.zip
PROTO_REPO=https://github.com/google/protobuf/releases/download
wget -q ${PROTO_REPO}/v${PROTO_VERSION}/${PROTO_ZIP} \
  && unzip ${PROTO_ZIP} -d ${PROTO_DIR} \
  && rm ${PROTO_ZIP}
cat <<EOF >> /etc/profile.d/sgx.sh
export PROTO_DIR=${PROTO_DIR}
export PROTOC_CMD=${PROTO_DIR}/bin/protoc
EOF
