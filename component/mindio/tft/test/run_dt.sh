#!/bin/bash
# ***********************************************************************
# Copyright: (c) Huawei Technologies Co., Ltd. 2024. All rights reserved.
# script for run dt
# version: 1.0.0
# change log:
# ***********************************************************************

set -e

PROJECT_HOME="$( cd "$( dirname "$0" )"/.. && pwd  )"
bash $PROJECT_HOME/build/build.sh -t debug --ut ON
if [ 0 != $? ];then
    echo "Failed to build ockio!"
    exit 1
fi

CMAKE_FILE="${PROJECT_HOME}/CMakeLists.txt"
BACKUP_FILE="${CMAKE_FILE}.backup"
TARGET_LINE="add_compile_definitions(_GLIBCXX_USE_CXX11_ABI=0)"

comment_cmake_config()
{
    # 备份原文件
    cp "${CMAKE_FILE}" "${BACKUP_FILE}"
    # 注释目标行
    sed -i "s/^[[:space:]]*${TARGET_LINE}/# &/" "${CMAKE_FILE}"
}

restore_cmake_config()
{
    if [ -f "${BACKUP_FILE}" ]; then
        cp "${BACKUP_FILE}" "${CMAKE_FILE}"
        rm "${BACKUP_FILE}"
    fi
}

comment_cmake_config

CURRENT_PATH=$(cd "$(dirname "$0")"; pwd)
cd "${CURRENT_PATH:?}"

main()
{
  hdt clean && hdt build && hdt run "--args=\"--gtest_output=xml:report.xml\""
  hdt report
  echo "done"
}

main

restore_cmake_config