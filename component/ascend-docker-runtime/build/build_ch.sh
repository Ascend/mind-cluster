#!/bin/bash

set -e

Ascend="Alan"
ascend="alan"

ROOT=$(cd $(dirname $0); pwd)/..

modify_files=(${ROOT}/build/scripts/run_main.sh ${ROOT}/build/scripts/uninstall.sh ${ROOT}/build/makeself-header/makeself-header.sh)
for cur_file in "${modify_files[@]}"
do
  sed -i "s/RT_LOWER_CASE=\"ascend-docker-runtime\"/RT_LOWER_CASE=\"${ascend}-docker-runtime\"/g" ${cur_file}
  sed -i "s/RT_FIRST_CASE=\"Ascend-docker-runtime\"/RT_FIRST_CASE=\"${Ascend}-docker-runtime\"/g" ${cur_file}
done

sed -i "s/ascend-docker-runtime/${ascend}-docker-runtime/g" ${ROOT}/build/scripts/help.info

sed -i "s#/var/log/ascend-docker-runtime/#/var/log/${ascend}-docker-runtime/#g" ${ROOT}/cli/src/logger.c

chmod +x build.sh && dos2unix build.sh
source build.sh

mv ${OUTPUT}/${RUN_PKG_NAME} ${OUTPUT}/${Ascend}-docker-runtime