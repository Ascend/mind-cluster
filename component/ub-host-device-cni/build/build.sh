#!/bin/bash
# Perform build ub-host-device CNI plugin
# Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.


set -e
CUR_DIR=$(dirname $(readlink -f $0))
TOP_DIR=$(realpath "${CUR_DIR}"/..)
export GO111MODULE="on"
arch=$(arch 2>&1)
echo "Build Architecture is" "${arch}"

OUTPUT_NAME="ub-host-device"
build_version="v26.2.0"

GIT_COMMIT=$(git rev-parse --verify HEAD 2>/dev/null || echo "unknown")
GIT_BRANCH=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")
GO_VERSION=$(go version | awk '{print $3}')

function clean() {
  rm -rf "${TOP_DIR}"/output
  mkdir -p "${TOP_DIR}"/output
}

function build() {
  cd "${TOP_DIR}"/cmd
  export CGO_ENABLED=1
  export CGO_CFLAGS="-fstack-protector-strong -D_FORTIFY_SOURCE=2 -O2 -fPIC -ftrapv"
  export CGO_CPPFLAGS="-fstack-protector-strong -D_FORTIFY_SOURCE=2 -O2 -fPIC -ftrapv"
  go build -mod=mod -buildmode=pie \
    -ldflags "-buildid none -s -w -extldflags=-Wl,-z,relro,-z,now,-z,noexecstack \
              -X ascend-common/common-utils/version.Version=${build_version} \
              -X ascend-common/common-utils/version.GitCommit=${GIT_COMMIT} \
              -X ascend-common/common-utils/version.GitBranch=${GIT_BRANCH} \
              -X ascend-common/common-utils/version.BuildOS=linux \
              -X ascend-common/common-utils/version.BuildArch=${arch} \
              -X ascend-common/common-utils/version.GoVersion=${GO_VERSION}" \
    -o "${OUTPUT_NAME}" \
    -trimpath .
  ls "${OUTPUT_NAME}"
  if [ $? -ne 0 ]; then
    echo "fail to find ub-host-device"
    exit 1
  fi
}

function mv_file() {
  mv "${TOP_DIR}"/cmd/${OUTPUT_NAME} "${TOP_DIR}"/output
  cp "${TOP_DIR}"/build/nad-cni-inherit-config.yaml "${TOP_DIR}"/output/nad-cni-inherit-config.yaml
  chmod 500 "${TOP_DIR}"/output/${OUTPUT_NAME}
}

function main() {
  clean
  build
  mv_file
}

main
