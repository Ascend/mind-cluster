#!/bin/bash

# Perform build clusterd
# Copyright @ Huawei Technologies CO., Ltd. 2024-2024. All rights reserved
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
# http://www.apache.org/licenses/LICENSE-2.0
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
# ============================================================================

set -e
CUR_DIR=$(dirname $(readlink -f $0))
TOP_DIR=$(realpath "${CUR_DIR}"/..)
export GO111MODULE="on"
VER_FILE="${TOP_DIR}"/service_config.ini
build_version="v6.0.0"
output_name="clusterd"
if [ -f "$VER_FILE" ]; then
  line=$(sed -n '1p' "$VER_FILE" 2>&1)
  #cut the chars after ':' and add char 'v', the final example is v3.0.0
  build_version="v"${line#*=}
fi

arch=$(arch 2>&1)
echo "Build Architecture is" "${arch}"
os_type=$(arch)

docker_zip_name="Ascend-mindxdl-clusterd_${build_version:1}_linux-${arch}.zip"


function clean() {
  rm -rf "${TOP_DIR}"/output
  mkdir -p "${TOP_DIR}"/output
}

function build() {
  cd "${TOP_DIR}"
  go mod tidy
  export CGO_ENABLED=1
  export CGO_CFLAGS="-fstack-protector-strong -D_FORTIFY_SOURCE=2 -O2 -fPIC -ftrapv"
  export CGO_CPPFLAGS="-fstack-protector-strong -D_FORTIFY_SOURCE=2 -O2 -fPIC -ftrapv"
  go build -mod=mod -buildmode=pie -ldflags "-X main.BuildName=${output_name} \
              -X main.BuildScene=${build_scene} \
              -X main.BuildVersion=${build_version}_linux-${os_type} \
              -buildid none \
              -s \
              -linkmode=external \
              -extldflags=-Wl,-z,relro,-z,now,-z,noexecstack" \
              -o "${output_name}"  \
              -trimpath
  ls "${output_name}"
  if [ $? -ne 0 ]; then
      echo "fail to find clusterd"
      exit 1
  fi
}

function mv_file() {
  mv "${TOP_DIR}/${output_name}" "${TOP_DIR}"/output
  cd "${TOP_DIR}"
  sed -i "s/clusterd:.*/clusterd:${build_version}/" "$CUR_DIR"/clusterd.yaml
  cp "$CUR_DIR"/Dockerfile "$TOP_DIR"/output/
  cp "$CUR_DIR"/faultDuration.json "$TOP_DIR"/output/
  cp "$CUR_DIR"/relationFaultCustomization.json "$TOP_DIR"/output/
  cp "$CUR_DIR"/publicFaultConfiguration.json "$TOP_DIR"/output/
  cp "$CUR_DIR"/clusterd.yaml "$TOP_DIR"/output/clusterd-"${build_version}".yaml
  sed -i "s#output/clusterd#clusterd#" "$TOP_DIR"/output/Dockerfile
  change_mod
}

function change_mod() {
  chmod 400 "$TOP_DIR"/output/*
  chmod 500 "${TOP_DIR}/output/${output_name}"
}

function main() {
  clean
  build
  mv_file
}

if [ "$1" = clean ]; then
  clean
  exit 0
fi
main
