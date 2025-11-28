#!/usr/bin/env bash
# ***********************************************************************
# Copyright (c) Huawei Technologies Co., Ltd. 2023. All rights reserved.
# script for Huawei hyperio to build pkg
# version: 1.0.0
# change log:
# ***********************************************************************
set -e

PROJECT_HOME="$( cd "$( dirname "$0" )"/.. && pwd  )"
DT_EXE_PATH=${PROJECT_HOME}/output/bin
GENERATE_DIR=${PROJECT_HOME}/dtfuzz_result

cd "${PROJECT_HOME}"/test
hdt comp -i
if [ 0 != $? ];then
  echo "Failed to install components!"
	exit 1
fi
cd -

find "${PROJECT_HOME}"/Build/Debug -type f -name "*.gcda" -print0 | xargs -0 rm -rf

bash "${PROJECT_HOME}"/build.sh -t debug --dtfuzz ON
if [ 0 != $? ];then
  echo "Failed to build ockio!"
	exit 1
fi

chmod 500 -R "${DT_EXE_PATH}"

cd "${PROJECT_HOME}"/output

export ASAN_OPTIONS="log_path=${PROJECT_HOME}/output/dtfuzz.log:detect_leaks=1"
export HCOM_FILE_PATH_PREFIX=${PROJECT_HOME}/output
export LD_LIBRARY_PATH=${PROJECT_HOME}/output/lib
./bin/dfs_dt_fuzz -fsanitize-coverage=trace-pc

[ -d "${GENERATE_DIR}" ] && rm -rf "${GENERATE_DIR}"
mkdir -p "${GENERATE_DIR}"

# coverage
lcov --d "${PROJECT_HOME}"/Build/Debug/src --c --output-file "${GENERATE_DIR}"/coverage.info --rc lcov_branch_coverage=1 --rc lcov_excl_br_line="LCOV_EXCL_BR_LINE|MFS_LOG*|ASSERT_*|BKG_*|gLastErrorMessage"
if [ 0 != $? ];then
  echo "Failed to generate all coverage info"
  exit 1
fi

lcov -r "${GENERATE_DIR}"/coverage.info "/usr/*" -o "${GENERATE_DIR}"/coverage.info --rc lcov_branch_coverage=1
if [ 0 != $? ];then
  echo "Failed to remove /usr/* from coverage info"
  exit 1
fi

lcov -r "${GENERATE_DIR}"/coverage.info "/opt/buildtools/*" -o "${GENERATE_DIR}"/coverage.info --rc lcov_branch_coverage=1
if [ 0 != $? ];then
  echo "Failed to remove /opt/buildtools/* from coverage info"
  exit 1
fi

lcov -r "${GENERATE_DIR}"/coverage.info "*/test/*" -o "${GENERATE_DIR}"/coverage.info --rc lcov_branch_coverage=1
if [ 0 != $? ];then
  echo "Failed to remove */tests/* from coverage info"
  exit 1
fi

lcov -r "${GENERATE_DIR}"/coverage.info "*/src/util/*" -o "${GENERATE_DIR}"/coverage.info --rc lcov_branch_coverage=1
if [ 0 != $? ];then
  echo "Failed to remove */tests/* from coverage info"
  exit 1
fi

lcov -r "${GENERATE_DIR}"/coverage.info "*/src/sdk/memfs/python_sdk/*" -o "${GENERATE_DIR}"/coverage.info --rc lcov_branch_coverage=1
if [ 0 != $? ];then
  echo "Failed to remove */tests/* from coverage info"
  exit 1
fi

lcov -r "${GENERATE_DIR}"/coverage.info "*/src/sdk/*" -o "${GENERATE_DIR}"/coverage.info --rc lcov_branch_coverage=1
if [ 0 != $? ];then
  echo "Failed to remove */tests/* from coverage info"
  exit 1
fi

lcov -r "${GENERATE_DIR}"/coverage.info "*/src/common/*" -o "${GENERATE_DIR}"/coverage.info --rc lcov_branch_coverage=1
if [ 0 != $? ];then
  echo "Failed to remove */tests/* from coverage info"
  exit 1
fi

lcov -r "${GENERATE_DIR}"/coverage.info "*/src/memfs/fuse/*" -o "${GENERATE_DIR}"/coverage.info --rc lcov_branch_coverage=1
if [ 0 != $? ];then
  echo "Failed to remove */tests/* from coverage info"
  exit 1
fi

lcov -r "${GENERATE_DIR}"/coverage.info "*/3rdparty/*" -o "${GENERATE_DIR}"/coverage.info --rc lcov_branch_coverage=1
if [ 0 != $? ];then
  echo "Failed to remove */3rdparty/* from coverage info"
  exit 1
fi

lcov -r "${GENERATE_DIR}"/coverage.info "*_generated.h" -o "${GENERATE_DIR}"/coverage.info --rc lcov_branch_coverage=1
if [ 0 != $? ];then
  echo "Failed to remove *_generated.h from coverage info"
  exit 1
fi

genhtml -o "${GENERATE_DIR}"/result "${GENERATE_DIR}"/coverage.info --show-details --legend --rc lcov_branch_coverage=1
if [ 0 != $? ];then
  echo "Failed to generate all coverage info with html format"
  exit 1
fi

echo
echo "Done generating tarball!"
echo
echo Success