#!/bin/bash

set -e
CUR_DIR=$(dirname "$(readlink -f "$0")")
TOP_DIR=$(realpath "${CUR_DIR}"/..)

source build.sh
docker_zip_name="${docker_zip_name/Ascend/Alan}"

sed -i 's/ascendjobs/alanjobs/g' "$TOP_DIR"/output/clusterd-*.yaml