#!/bin/bash
# Copyright (c) Huawei Technologies Co., Ltd. 2026. All rights reserved.
set -e

CUR_DIR="$(dirname "${BASH_SOURCE[0]}")"
MINDIO_DIR="$(realpath "${CUR_DIR}/..")"

cd "$MINDIO_DIR"/acp/build
dos2unix *.sh && chmod +x *
./build.sh
cd "$MINDIO_DIR"/tft/build
dos2unix *.sh && chmod +x *
./build.sh

mkdir -p "$MINDIO_DIR"/output
cp "$MINDIO_DIR"/acp/output/*.whl "$MINDIO_DIR"/output
cp "$MINDIO_DIR"/tft/output/*.whl "$MINDIO_DIR"/output

echo "Successfully built mindio_acp and mindio_ttp."