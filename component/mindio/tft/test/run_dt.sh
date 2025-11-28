#!/bin/bash
# ***********************************************************************
# Copyright: (c) Huawei Technologies Co., Ltd. 2024. All rights reserved.
# script for run dt
# version: 1.0.0
# change log:
# ***********************************************************************

set -e

CURRENT_PATH=$(cd "$(dirname "$0")"; pwd)
cd "${CURRENT_PATH:?}"

main()
{
  hdt clean && hdt build && hdt run "--args=\"--gtest_output=xml:report.xml\""
  hdt report
  echo "done"
}

main